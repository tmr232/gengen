Hello!

My name is Tamir, and I am here to talk to you about generators!

But first, a little about me.
You see, this is a Go talk, but I am relatively new to Go.
In fact, this is my largest Go project to date...

For the most part, I am a Python programmer.
But with a strong interest in programming languages.

A few months back, I started learning Go, knowing that it will soon take a
significant part in my professional career.

So I started reading and experimenting.
Watching talks and listening to podcasts.
Coming from Python, I wanted to be able to write Go, not Gython...

And while it's not easy changing programming styles, I liked a lot of what I saw.

First and foremost, Go's tooling is phenomenal. 
Being able to run, lint, and test my code...
To manage dependencies and to package code...
Even profiling and fuzzing!
And everything Just Works(TM)

True, some things are still weird or foreign to me, and I still miss some Python features
when writing in Go. 
But in return - I get a simple, consistent language.
When I look at a codebase I did not write, I can quite easily find my way around it.
When I see I line of code - I know what it does.
There's no magic and voodoo everywhere.
(Which, I admit, I like a lot. Just not for production code...)

But while Go's receivers and interfaces do a good job at replacing classes & inheritance,
and errors, despite the verbosity, do a good job at replacing exceptions...
I still sorely miss Generators. 
In fact, I think that the lack of generators is making some parts of the language more complicated.
It adds complexity, and reduces maintainability.

Let me show you.

---

I'll share a bit of my workflow.

When I start working with a new API - be it a remote, REST API, or a local API or data structure,
I usually write a _very bad_ piece of mangled code just to see that it's working.

If, for example, I want to get all the items of a paginated API, I'll start by doing everything in a single function.
Initialization, querying the API, iterating, and printing:

```go
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
```

This is more or less what my code will look like - panics and all.
I am not yet writing production code - I am just trying to see that things work.

Now, once things do work, and I see all the items printed as expected, starts the hard part.
You see, I want to have a single instance of this code, and just iterate over the results.

In Go, this means one of two options.
One - collect all the results in a slice, and return that. 
This works, but only if the number of items is small.
Additionally, it may unneeded requests across the network if we only need _some_ of the items.
The code will look something like this:

```go
items, err := GetAllItems(client)
if err != nil {
	panic(err)
}
for _, item := range items {
	fmt.Println(item)
}
```

Two - create an iterator. With an iterator, we can iterate over the items seamlessly, and know 
that we're not doing any extra work.
The code will be as follows:

```go
itemIterator := IterItems(client)
for itemIterator.Next() {
	item := itemIterator.Value()
	fmt.Println(item)
}
if itemIterator.Err() != nil {
	panic(err)
}
```

And here the trouble starts.
Some of you may be familiar with 
```go
type Iterator[T any] interafce {
	Next() bool
	Value() T
	Err() error
}
```
Which is great.
But looking online I found several other variants.
One is the Java-style iterator

```go
type Iterator[T any] interface {
    HasNext() bool
	Next() T
}
```

And even the suggestion to use channels for iteration.
Which, of course, you should not do.
It will also be the last time in this talk I mention channels.

Now, once we settle on an interface for iteration, we need to implement it.

An iterator will generally be comprised of a struct, to hold it's state;
a constructor, to initialize it; and 3 methods to implement it.
Throughout this talk the baseline for the struct will be 
```go
type IteratorImpl[T any] struct {
	value T
	err error
}
```
So that both `Value()` and `Err()` are trivial getters and we never need to mention them again.

In our case, the iterator struct will contain the following:

```go
type ThingIterator struct {
	client *ApiClient
	result *Result
	index  int
	value  Thing
	err    error
}
```

The client we're querying; the _current_ result (as a pointer, since we don't have one when we start iteration);
an index into `result.Things`, and the previously mentioned `value` and `err` members.

Our `Next()` function will look as follows:

```go
func (it *ThingIterator) Next() bool {
	if it.result == nil {
		result, err := it.client.GetThings(nil)
		if err != nil {
			it.err = err
			return false
		}
		it.result = &result
	}
	
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
	
	return it.Next()
}
```
First - if we don't have any result yet, we get it.
If we encounter an error, we report it and finish.
Then, we check if the current index fits in the current result - if it does, set `value`.
If we consumed the current result, we check if there's another one.
If there isn't, we terminate.
Once we get the new result, we just call `Next()` again to get the actual value.

And this is where our second, and more pressing issue is.
Now, you may be far better programmers than I am.
Or more experienced with this style of iteration.
But while the original printing loop took me a single attempt to get right, 
this one took me multiple tries, and significantly more time.
When I was sure it works, I tested it, and it still didn't.

You may also want to have a look at an iterator for nested slices:

```go
func (it *LibraryBookIterator) Next() bool {
	it.bookIndex++
	for it.bookIndex >= len(it.library.Rooms[it.roomIndex].Shelves[it.shelfIndex].Books) {
		it.bookIndex = 0
		it.shelfIndex++
		for it.shelfIndex >= len(it.library.Rooms[it.roomIndex].Shelves) {
			it.shelfIndex = 0
			it.roomIndex++
			if it.roomIndex >= len(it.library.Rooms) {
				return false
			}
		}
	}

	return true
}
```

To me, that's not readable, intuitive, or maintainable.
It's a difficult pattern to work with.
Not just because the lack of familiarity, but because it causes a split in the control flow.
We have a function that does a _single_ iteration, and the actual loop is only present in the call-site.
As a reader, I never really see the whole picture. 
And, unlike loops - there are many different ways to implement this code. 
Many styles, and different considerations.
And if the API changes, you need to work out the kinks in this complex piece of code.


Now, with Generators, on the other hand, the code will look as follows:

```go
func IterThings(client *ApiClient) gengen.Generator[Thing] {
	var result Result
	var err error
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
	return nil
}
```

Compare it with the print function:

```go
func PrintThings(client *ApiClient) {
	var result Result
	var err error
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
```

The return type has changed, as we're now creating a Generator.
Instead of panicking, we're returning the errors.
And instead of printing the item, we yield it.

Now, `IterThings` gives as an iterator which is, functionally, completely equivalent to the one we implemented before.
But now, the function is clear.
We can see the loop, we can see the state, and we don't need to bend our mind to break each iteration into a function.
The desired flow is clearly visible. 
And while this adds a bit of syntax and semantics we need to learn, I find it significantly simpler than what we had to
implement before.

In the rest of this talk, I'll show you how we make that happen.

---

Now, before we dig too deep, we need to establish a common language.

First, the syntax of Generators.

A function is a generator if and only if it A) has a `gengen.Generator` return type
and B) yields values using `gengen.Yield`.

Generators use `gengen.Yield` to yield values & suspend their execution,
and `return` to report errors and complete their execution.

`gengen.Generator` implements the iterator interface we described earlier.

---

Under the hood, Generators are implemented using Iterators. 
As we've seen before, the only interesting parts of an Iterator are its state struct 
and its `Next()` function.

With that being the case, we can represent Iterators as closures, with a small helper struct.

```go
type SingleFunctionIterator[T any] struct {
	Advance func() (hasValue bool, value T, err error)
	value   T
	err     error
}

func (it *SingleFunctionIterator[T]) Next() bool {
	hasValue, value, err := it.Advance()
	it.value = value
	it.err = err
	return hasValue
}

func (it *SingleFunctionIterator[T]) Value() T {
	return it.value
}

func (it *SingleFunctionIterator[T]) Err() error {
	return it.err
}
```

```go
func Fibonacci() SingleFunctionIterator[int] {
	a := 1
	b := 1
	return SingleFunctionIterator[int]{
		Advance: func() (hasValue bool, value int, err error) {
			value = a
			a, b = b, a+b
			return true, value, nil
		},
	}
}
```

With that, we can move all the implementation of a specific iterator into a single function.
We'll have a State-Block at the top replacing the struct we used earlier.
And and the `Advance` function replacing `Next`. 
Here, instead of setting the `value` and `err` members in our state struct, we just return them.
The helper struct, `SingleFunctionIterator[T]` will handle them for us.

This new form simplifies things significantly.
It's now easier to talk about the code, and easier to fit it into a slide.

So during the talk, I'll be mentioning the `Advance` function and the State-Block quite often.

---

With that bit of prep-work behind us, it is time to start digging deep.