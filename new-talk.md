# Title Slide

---

# I'm New to Go

So a bit about me.

I'm a relatively new Gopher - I've only been working with Go for the last year.

The majority of my background is Python, with a bit of C++ and Assembly sprinkled in.

Moving from those languages to Go is quite the experience.

Go has a far simpler syntax and ecosystem.
It's easy to get started, to get productive, and to deploy your code.
And when looking at someone else's code - be it a colleague, an open source project,
or the standard library - it is relatively easy to get around and see what's going on.
That, especially, was a new experience for me.

True, I had to let go of some features.
Templates, exceptions, lots of syntactic sugar...
I can't say it was easy, but I managed.
That said, there's one thing that I just kept trying to reach for.
That feature is Generators.

In 2 words, since we'll dig into that later - Generators are a simple and straightforward way to implement Iterators.

It seems that I wanted Generators badly enough to create a tool and a talk...

But before we dig into that, I want to talk to you about my workflow.

---

# My Workflow

When we encounter a new API, we always need to go through a process of learning it.
Reading the docs, going over the public functions, structures, etc.

Now, in every project, beyond the syntax and algorithms and tooling - we work with data.
However we frame it, the purpose of our code is to manipulate data.

So when faced with a new API the first thing I do is try to understand the data.
Documentation and structures are nice, but what helps me most is seeing some concrete data.

That's why the first thing I do is query the API for data, and print it out.
Just loop over it, panicking at any error, and print.

This allows me to see that I'm using the API correctly, and get an understanding of the data itself.


Once that is done, and I have a basic understanding, it's time to start writing the business logic.
But now, we face a choice. 
How do we consume the data?

One option is to stick all the logic in the function we just wrote.
It works, kinda, but smells of bad design.
It mixes querying and processing, and will need to be copied into any
new function handling the data.

Another option is to query first, and process later.
We can query _all_ the data, put that in a slice, and then process the slice.
This is simple and straightforward, but may get intractable if our data is too big or if it requires network-communication.

```go
data, err := getData()
if err != nil {
	panic(err)
}
for _, value := range data {
	fmt.Println(value)
}
```

Our third option, and my preferred one, is to write Iterators.

---

# Iterators

Iterators allow us to loop over our data _as if_ it was in a slice, 
but only get the next item on demand, when we actually need it.

```go
iterator := getIterator()
for iterator.Next() {
	value := iterator.Value()
	fmt.Println(value)
}
if iterator.Err() != nil {
	panic(iterator.Err())
}
```

We use `Next()` to check if another value exists in the iterator.
`Value()` to get that value, if one exists.
When iteration is done, we check `Err()` to see if we stopped because of an error.

They allow separating the query from the processing, without incurring any significant runtime cost. 

---

# Implementing Iterators

The problem with iterators, then, comes from implementing them.

Generally, an iterator is made up of 5 parts.

```go
type MyIterator[T any] struct{...}
func NewIterator[T any]() *MyIterator[T]
func (it* MyIterator) Value() T
func (it* MyIterator) Err() error
func (it* MyIterator) Next() bool
```

A struct to maintain the iterator state;
A constructor, as we often have some non-obvious initialization;
`Value()` and `Err()` methods to get the values and errors, which are usually implemented as trivial getters;
And a `Next()` method, which does all the heavy lifting.

For the duration of the talk, we'll be using a helper struct called `ClosureIterator[T any]` that allows 
us to collapse all that into a function and closure:

```go
package Example

func rangeIterator(stop int) ClosureIterator[int] {
	current := 0

	return ClosureIterator[int]{
		Advance: func(withValue func(value int) bool, withError func(err error) bool, exhausted func() bool) bool {
			if current < stop {
				retval := current
				current++
				return withValue(retval)
			}
			return exhausted()
		},
	}
}
```

Our new function acts as a constructor, returning an initialized iterator.
At the top of the function, we have our state-block, holding the state of the iterator, in-place of a struct.
Then, our `Advance()` function takes the role of `Next`.
Either returning with a new value, continuing the iteration;
returning with an error, stopping the iteration;
Or just reporting that the iteration has been exhausted.

---

# Implementing Iteartors - Real Sample

Now, as we look at the iterator implementation of our original sample - we can see the issue.

This code is now significantly more complex than what we originally had.
The original looping structure is gone, and we have a lot of checks just to maintain
the state that was automatically handled before.

Now, it might be my lack of experience, but implementing this was _hard_  for me.
While the original loop worked on the first try, this took several failed attempts,
and I could only get it properly working using TDD.

To me - this looks like a maintainability nightmare.
It's hard to reason about, as the flow is awkwardly broken up,
and the difficulty in implementing it guarantees that changes will be similarly painful.

---

# Implementing Iterators - Generators

This is why I want Generators so badly.

With generators, the code looks as follows:

**Show sample here**

Now compare it with the original printing loop.

We changed the return value, as we now return a generator;
We changed `panic` to `return` to report errors;
And we changed `print` to `yield`.
That's it.
No extra code, no extra logic.

When we call our generator function - it returns a generator, but does not execute any code yet.
Then, when we call the generator's `Next()` method,
the function will run until we reach `Yield`, yielding a value that'll be available via the `Value()` method.
When we call `Next()` again, we'll continue right after the `Yield`.
When we reach a `return` statement iteration will end.
We use `return err` to report errors.

---

# Generating Generators

Now, as nice as that generator syntax is, we still have one problem - it isn't Go.

We can write this code, but we cannot run it.

To circumvent this issue, we're going to use code generation.
We'll take our generators, written in pretend-Go, and automatically generate real-Go implementations.

After all, the talk _is_ called "Generating Generators"


To generate our implementation code, we'll use some of Go's fantastic tooling for code-analysis and code-generation.

---

# Generating Generators - Build Tricks

To generate our code, we'll be using `go-generate`.
But we also need to make sure the pretend-Go code does not get built.

To do that, we'll be using a build tag.
Every file with generator definitions needs to be tagged with `//go:build gengen`.
When we run `go generate -tags gengen`, we'll copy the file over, transforming the pretend-Go into real Go along the way.
The resulting file will have a `//go:build !gengen` tag, ensuring they won't conflict;
It will also have all the normal Go code from our original file, and the implementation of our generators.

When we build and run our project - only the real-Go generator-implementations will be included.
The pretend-Go definitions will be entirely ignored.

---

# Generating Generators - Code Transformations

Now that we're done with the introductions, it's time to get to the core of generating generators - 
Code Transformations.

We'll gradually walk over various code transformations, building on what we cover,
to achieve fully functional generators.