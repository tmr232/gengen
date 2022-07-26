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
This is simple and straightforward, but may get intractable if our data is too big or if it requires
network-communication.

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
package Example

type MyIterator[T any] struct{}
func NewIterator[T any]() *MyIterator[T] {
	return new(MyIterator[T])
}
func (it* MyIterator[T]) Value() T
func (it* MyIterator[T]) Err() error
func (it* MyIterator[T]) Next() bool

func init() {
	NewIterator[int]()
}
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
When we run `go generate -tags gengen`, we'll copy the file over, transforming the pretend-Go into real Go along the
way.
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

First, we'll look at a trivial sample - the empty generator

```go
package Example

func Empty() gengen.Geneartor[int] {
	return nil
}
```

To implement it, we copy the body of the function into a closure-iterator's `Advance()` function

```go
package Example

func Empty() ClosureIterator[int] {
	return ClosureIterator[int]{
		Advance: func(...) bool {
			return nil
		}
	}
}
```

And replace `return nil` with `return exhausted()`, as we have neither values or errors.

```go
package Example

func Empty() ClosureIterator[int] {
	return ClosureIterator[int]{
		Advance: func(...) bool {
			return exhausted()
		}
	}
}
```

If we were to return an error instead

```go
package Example

func Error() gengen.Geneartor[int] {
	return MyError{}
}
```

We'd replace the `exhausted()` call with `withError(MyError{})`.

```go
package Example

func Empty() ClosureIterator[int] {
	return ClosureIterator[int]{
		Advance: func(...) bool {
			return withEror(MyError{})
		}
	}
}
```

---

# Yielding Values

Now that we handled stopping iteration and returning errors, we need to handle values!

```go
package Example

func HelloWorld() gengen.Generator[string] {
	gengen.Yield("Hello, World!")
	return nil
}
```

Once more, we copy the body into the `Advance` function:

```go
package Example

func Count() ClosureIterator[string] {
	return ClosureIterator[string]{
		Advance: func(...) bool {
			gengen.Yield("Hello, World!")
			return nil
		},
	}
}
```

And now it's time to transform `gengen.Yield` (which is actually an empty function)

```go
package Example

func Yield[T any](T) {}
```

Into meaningful Go code.

When `Next()` will be called, we need to yield a value, so we start with that:

```go
package Example

func Count() ClosureIterator[string] {
	return ClosureIterator[string]{
		Advance: func(...) bool {
			// gengen.Yield("Hello, World!")
			return withValue("Hello, World!")
			return nil
		},
	}
}
```

Next, we need to ensure that the next time we call `Next()`, we'll continue right after that.
So we add a label, a `goto`, and a bit of state

```go
package Example

func Count() ClosureIterator[string] {
	next := 0
	return ClosureIterator[string]{
		Advance: func(...) bool {
			switch next {
			case 0:
				goto Label0
			case 1:
				goto Label1
			}
		Label0:
			// gengen.Yield("Hello, World!")
			next = 1
			return withValue("Hello, World!")
		Label1:
			return nil
		},
	}
}
```

The first time we go into the function - we start from the top - `Label0`
The second time, `next` will be set to `1` so we'll start with `Label1`.

Last but not least - we transform `reutrn nil` like we did before, and we're done:

```go
package Example

func Count() ClosureIterator[string] {
	next := 0
	return ClosureIterator[string]{
		Advance: func(...) bool {
			switch next {
			case 0:
				goto Label0
			case 1:
				goto Label1
			}
		Label0:
			// gengen.Yield("Hello, World!")
			next = 1
			return withValue("Hello, World!")
		Label1:
			return exhausted()
		},
	}
}
```

---

# Using Goto

As you've all just seen - we're using `goto` in our code.
Yes, there are other ways to implement this, but `goto` and labels are by far the simplest.
And with autogenerated code - we're ok with that.

As we'll see in a moment - `goto` in Go, being a safe language, has two limitations we need to handle

1. It cannot jump over variable definitions. If it did - the variable would be in an un-initialized
   state and we'd lose safety.
2. It cannot jump into a block. If it did - we won't know whether the block's invariants hold,
   and may have all sorts of bugs.

**Show examples!**

Since we _will_ have both variables and blocks in our generators, we need to deal with those limitations.
---

# Using Goto - Variable Definitions

First, variables.
To circumvent the issue, we move _all_ variable definitions from the generator body and into the state-block
in our implementation.
Since the `goto`s only happen _inside_ the `Advance()` function, we will never jump over definitions.

This also serves a second purpose - by declaring all variables in the state-block we can maintain
their state across calls to `Next()`.

**Need code sample!!!**
   
---

# Using Goto - Blocks

Go (like many other languages) uses blocks for both control-flow and scoping of variables.
Since we moved all variable definitions into the state-block, effectively eliminating scoping,
blocks are now only used for control-flow.

With that in mind, we can go ahead and transform those blocks away!

# Control Flow - If

The first construct we'll handle is an `if` statement:

```go
package Example

func Count(alpha bool) gengen.Generator[string] {
	if alpha {
		gengen.Yield("a")
		gengen.Yield("b")
		gengen.Yield("c")
	} else {
		gengen.Yield("1")
		gengen.Yield("2")
		gengen.Yield("3")
	}
	return nil
}
```

It can be represented with the following graph: (show graph)

We can add a few labels just to make the various parts:
```go
if alpha {
thenLabel:
   gengen.Yield("a")
   gengen.Yield("b")
   gengen.Yield("c")
} else {
elseLabel:
   gengen.Yield("1")
   gengen.Yield("2")
   gengen.Yield("3")
}
afterLabel:
```

And even a couple of `goto`s to make the flow "explicit"

```go
if alpha {
thenLabel:
    gengen.Yield("a")
    gengen.Yield("b")
    gengen.Yield("c")
    goto afterLabel
} else {
elseLabel:
    gengen.Yield("1")
    gengen.Yield("2")
    gengen.Yield("3")
    goto afterLabel
}
afterLabel:
```

Then, we can move our code out of the blocks while maintaining the exact same flow

```go
if alpha {
    goto thenLabel
} else {
    goto elseLabel
}
thenLabel:
    gengen.Yield("a")
    gengen.Yield("b")
    gengen.Yield("c")
    goto afterLabel
elseLabel:
    gengen.Yield("1")
    gengen.Yield("2")
    gengen.Yield("3")
    goto afterLabel
afterLabel:
```

Now that we have our code out of the blocks, we can continue with the transformations we saw before,
and get a working iterator.

---

# Control Flow - Forever

With `if` done, we proceed to transforming `for` loops in a similar manner.

```go
n := 0
for {
	gengen.Yield(n)
	n++
}
```

We add the relevant labels:

```go
n := 0
for {
 loopHead:
	gengen.Yield(n)
	n++
}
afterLoop:
```

Goto's

```go
n := 0
for {
 loopHead:
	gengen.Yield(n)
	n++
	goto loopHead 
}
afterLoop:
```

And move the code out of the loop entirely:

```go
n := 0
loopHead:
 gengen.Yield(n)
 n++
 goto loopHead 
afterLoop:
```

A `break` will be transformed to `goto afterLoop`, and a `continue` to `goto loopHead`

---

# Control Flow - While

The next `for` loop to transform is the while-equivalent

```go
n := 0
for n < 10 {
	gengen.Yield(n)
	n++
}
```
Now, we can start building on what we've done before, and transform it to a `forever` loop with an `if` statement:

```go
n := 0
for {
	if n < 10 {
       gengen.Yield(n)
       n++
    } else {
		break
    }
}
```

Which will eventually transform into

```go
n := 0

loopHead:
	if n < 10 {
		goto loopBody
    } else {
		goto afterLoop
    }
loopBody:
     gengen.Yield(n)
     n++
afterLoop:
```

---

# Control Flow - C-Style Loop

A C-style loop can now go through a similar transformation:

```go
for n := 0; n < 10;  n++ {
	gengen.Yield(n)
}
```

Becomes

```go
n := 0
for ;;  n++ {
	if n < 10 {
      gengen.Yield(n)
   }
}
```

But we need to be careful about the increment when dealing with `continue`.
So the end result is:

```go
n := 0

loopHead:
   if n < 10 {
        goto loopBody
   } else {
        goto afterLoop
   }
loopBody:
    gengen.Yield(n)
loopIncrement:
    n++
afterLoop:
```

With a `continue` translated as `goto loopIncrement` and not `goto loopHead`.

---

# Control Flow - for-range

Last but not least - we have our `for range` loops.

```go
for index, value := range numbers {
	
}
```

Those pose a new problem, as there is no visible state to preserve as we iterate through them.

To solve this, we introduce adaptors.

A slice adaptor wraps a slice in an iterator, allowing us to iterate through it:

```go
iter := SliceAdaptor(slice)
for iter.Next() {
	index, value := iter.Value()
}
```

And from this point, we can use the previous transformations to finish the implementation.

The same goes for maps (though we are forced to copy all keys and values into a slice first, to guarantee order).
And same for channels.

---

# Control Flow - To Be Continued

Go has more control-flow structures, and they can all be transformed in a similar manner.
After all, Go is not _entirely_ without syntactic sugar.

---
# Demo

So we've seen why we want generators.
We've talked about code generation, and walked through the transformations.
Now it's time to see it in action!

```go
//go:build gengen

package sample

import (
	"github.com/tmr232/gengen"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

func Fibonacci() gengen.Generator[int] {
	a := 1
	b := 1
	for {
		gengen.Yield(a)
		a, b = b, a+b
	}
}
```

We run `go generate -tags gengen`, and get

```go
//go:build !gengen

// AUTOGENERATED DO NOT MODIFY

package sample

import (
	"github.com/tmr232/gengen"
)

func Fibonacci() gengen.Generator[int] {

	var a int

	var b int

	__next := 0
	return &gengen.GeneratorFunction[int]{
		Advance: func() (__hasValue bool, __value int, __err error) {
			switch __next {

			case 0:
				goto __Next0

			case 1:
				goto __Next1

			}

		__Next0:

			a = 1
			b = 1

		__Head1:

			__hasValue = true
			__value = a
			__next = 1
			return

		__Next1:

			a, b = b, a+b
			goto __Head1

		},
	}
}

```

And as we run a simple test to verify it - it works!