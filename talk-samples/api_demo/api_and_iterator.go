//go:build gengen

package main

import (
	"fmt"
	"github.com/tmr232/gengen"
	"log"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

type Id int

type Thing string

func (t Thing) String() string {
	return string(t)
}

type Result struct {
	Next   *Id
	Things []Thing
}

type ApiClient struct {
	results []Result
	index   Id
}

func Ptr[T any](value T) *T {
	return &value
}

func NewApiClient() *ApiClient {
	return &ApiClient{
		index: 0,
		results: []Result{
			Result{Next: Ptr(Id(1)),
				Things: []Thing{"First thing", "Another thing", "Things are awesome!"},
			},
			Result{Next: Ptr(Id(2)),
				Things: []Thing{"Yet another thing!"},
			},
			Result{Next: nil,
				Things: []Thing{"I'm a thing, too!", "One last thing!"},
			},
		},
	}
}

func (api *ApiClient) Auth(token string) error {
	return nil
}

func (api *ApiClient) GetThings(next *Id) (Result, error) {
	if next == nil {
		return api.results[0], nil
	}
	return api.results[*next], nil
}

func PrintThings() {
	client := NewApiClient()
	err := client.Auth("my-token")
	if err != nil {
		panic(err)
	}

	var result Result
	for {
		result, err = client.GetThings(result.Next)
		if err != nil {
			panic(err)
		}
		for _, thing := range result.Things {
			fmt.Println(thing)
		}
		if result.Next == nil {
			break
		}
	}
}

func IterThings(client *ApiClient) gengen.Generator[Thing] {
	var result Result
	var err error
	var x = 1
	for {
		result, err = client.GetThings(result.Next)
		if err != nil {
			return err
		}
		for _, thing := range result.Things {
			gengen.Yield(thing)
		}
		if result.Next == nil {
			break
		}
	}
	fmt.Println(x)
	return nil
}

type ThingIterator struct {
	client *ApiClient
	result *Result
	index  int
	value  Thing
	err    error
}

func NewThingIterator(client *ApiClient) *ThingIterator {
	return &ThingIterator{client: client}
}

func (it *ThingIterator) Next() bool {
	if it.result == nil {
		result, err := it.client.GetThings(nil)
		if err != nil {
			it.err = err
			return false
		}
		it.result = &result
	}
	for {
		if it.index < len(it.result.Things) {
			it.value = it.result.Things[it.index]
			it.index++
			return true
		}
		it.index = 0

		if it.result.Next == nil {
			return false
		}

		result, err := it.client.GetThings(it.result.Next)
		if err != nil {
			it.err = err
			return false
		}
		it.result = &result
	}
}

func (it *ThingIterator) Value() Thing {
	return it.value
}

func (it *ThingIterator) Err() error {
	return it.err
}

func main() {
	client := NewApiClient()
	err := client.Auth("my-token")
	if err != nil {
		log.Fatal(err)
	}

	thingIterator := NewThingIterator(client)
	for thingIterator.Next() {
		fmt.Println(thingIterator.Value())
	}
	if thingIterator.Err() != nil {
		log.Fatal(err)
	}

	PrintThings()

	client = NewApiClient()
	thingGen := IterThings(client)
	for thingGen.Next() {
		fmt.Println(thingGen.Value())
	}
	if thingGen.Error() != nil {
		log.Fatal(err)
	}
}
