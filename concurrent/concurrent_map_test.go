package concurrent

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	item1       = "item"
	value1      = "value"
	nKeys       = 100000
	ngoroutines = 100
)

func TestAddOrUpdate(t *testing.T) {

	Convey("When an item is added", t, func() {

		cmap := NewConcurrentMap()
		cmap.AddOrUpdate(item1, value1, func(k string, v interface{}) interface{} { return v })

		Convey("It should be in the key list", func() {
			So(item1, ShouldBeIn, cmap.GetKeys())
		})

		Convey("It should be available to Get", func() {
			value, ok := cmap.Get(item1)
			So(value.(string), ShouldEqual, value1)
			So(ok, ShouldBeTrue)
		})

		Convey("When the item is removed", func() {
			cmap.Delete(item1)

			Convey("It shouldn't be available to Get", func() {
				value, ok := cmap.Get(item1)
				So(ok, ShouldBeFalse)
				So(value, ShouldBeNil)

			})

			Convey("The key list must be empty", func() {
				So(cmap.GetKeys(), ShouldBeEmpty)
			})

		})
	})

	Convey("When 2 goroutines uses the same map to addorupdate", t, func() {

		cmap := NewConcurrentMap()
		keys := mockedKeyList(nKeys)
		var wg sync.WaitGroup

		// create the goroutines functions
		addinmap := func() {
			for _, key := range keys {
				rand.Seed(int64(time.Now().Nanosecond()))
				cmap.AddOrUpdate(key, rand.Float64(), func(k string, val interface{}) interface{} { return val })
			}
			wg.Done()
		}
		wg.Add(1)
		go addinmap()
		wg.Add(1)
		go addinmap()

		Convey("All items must exists when the goroutines have finished ", func() {
			wg.Wait()
			sort.Strings(keys)
			mapKeys := cmap.GetKeys()
			sort.Strings(mapKeys)
			So(len(mapKeys), ShouldEqual, len(keys))
			So(keys, ShouldResemble, mapKeys)
		})
	})

	SkipConvey("When n goroutines uses the same map to add parts of a key's set", t, func() {
		cmap := NewConcurrentMap()
		keys := mockedKeyList(nKeys)

		var wg sync.WaitGroup

		splitedKeys := map[int][]string{}

		qtdSlice := nKeys / ngoroutines

		for i := 0; i < ngoroutines; i++ {
			splitedKeys[i] = keys[i*qtdSlice : (i*qtdSlice)+qtdSlice]
		}

		addGoroutine := func(id int) {
			defer wg.Done()
			for _, key := range splitedKeys[id] {
				rand.Seed(int64(time.Now().Nanosecond()))
				cmap.AddOrUpdate(key, rand.Float64(), func(k string, val interface{}) interface{} { return val })

			}
			Println("added: ", len(splitedKeys[id]))
		}

		// run all n goroutines
		for i := 0; i < ngoroutines; i++ {
			wg.Add(1)
			go addGoroutine(i)
		}

		time.Sleep(1 * time.Millisecond)
		Println("len antes de ler: ", len(cmap.GetValues()))

		Convey("And n others goroutines uses the same map to get the key list and store in another map, deleting these keys from map", func() {

			cmapStoredKeys := NewConcurrentMap()
			removedQnt := 0

			getGoroutine := func(id string) {
				defer wg.Done()
				fromMapKeys := cmap.GetKeys()
				cmapStoredKeys.AddOrUpdate(id, fromMapKeys, func(k string, val interface{}) interface{} { return val })
				Println("removed: ", removedQnt, "len :", len(fromMapKeys))
				removedQnt += len(fromMapKeys)
				for _, v := range fromMapKeys {
					cmap.Delete(v)
				}
			}

			i := 0
			for ; removedQnt < (nKeys/ngoroutines)*ngoroutines; i++ {
				wg.Add(1)
				go getGoroutine(strconv.Itoa(i))
				time.Sleep(10 * time.Millisecond)
				//Println(i, removedQnt)
			}
			Println("Quantity of goroutines to read/delete created:", i)

			Convey("All added keys should be values in the map  used to store retrieved keys", func() {
				wg.Wait()
				storedKeys := []string{}
				for _, v := range cmapStoredKeys.GetValues() {
					storedKeys = append(storedKeys, v.([]string)...)
				}
				sort.Strings(storedKeys)
				sort.Strings(keys)
				So(len(storedKeys), ShouldEqual, len(keys))
				So(storedKeys, ShouldResemble, keys)
			})
		})

	})

}

func mockedKeyList(n int) []string {
	keys := []string{}
	//prepare key list
	for i := 0; i < n; i++ {
		keys = append(keys, item1+strconv.Itoa(i))
	}
	return keys
}
