# gengen
Generating Generators!

## Usage

To use gengen you need to create a new Go file:

```go
//go:build gengen

package myPackage
```

You MUST use the `gengen` build tag, or the project will not work.

In this file, write your generators, and _only_ your generators.


```go
//go:build gengen

package myPackage

func Fibonacci() gengen.Generator[int] {
	a := 1
	b := 1
	for {
		gengen.Yield(a)
		a, b = b, a+b
	}
}
```

For a function to be a generator, two requirements must be met:

1. The function must contain a call to `gengen.Yield(t)`;
2. The function must return a `gengen.Generator[T]` type;

Requirement (1) is what exists in Python.
Requirement (2) is needed because Go uses static typing, and it must be declared.

To generate your generators, run 

```bash
go run github.com/tmr232/gengen/cmd/gengen
```

This can be done manually from the command line, or using `//go:generate`.
Note that this only needs to be run once per package, but multiple runs won't cause any harm.

When executed, the gengen command will generate `myPackage_gengen.go` in your package directory.

The file will look something like this:

```go
//go:build !gengen

package myPackage

import (
	"github.com/tmr232/gengen"
)

func Fibonacci() gengen.Generator[int] {
    // ...
}
```

That looks suspicious. 

Our newly generated `Fibonacci` generator is in the same package,
and it has the exact same name and signature.
Shouldn't it create a name conflict?

Well, this is (the first place) our little `gengen` build tag comes into play.
By using opposing build tags in both files,
we guarantee that they are never visible at the same time,
and prevent the name conflict!

This also lets us do some other neat tricks.
But we'll get to that.
First, we need to discuss the generator interface.

## Generator Interface

The generator interface is pretty straight-forward.

```go
package gengen
type Generator[T any] interface {
	Next() bool
	Value() T
	Error() error
}
```

The `Next` function lets us check whether a value exists,
the `Value` function returns said value,
and the `Error` function reports errors.

Iterating over a generator is as follows:

```go
func main() {
	gen := Fibonacci()
	for gen.Next() {
	    fmt.Println(gen.Value())	
    }
	if gen.Error() != nil {
		fmt.Println("Encountered an error!")
    }
}
```

If an error is encountered, the generator can report it, and stop iteration.

This means that a generator has 2 output channels - values and an error.
We already know that `gengen.Yield(t)` is used to output values,
but how can it report an error?

## Reporting Errors

To report an error, a generator simply returns it.

```go
func ChunkReader(reader io.Reader, bufferSize int) gengen.Generator[[]byte] {
	buffer := make([]byte, bufferSize)
	for {
        n, err := reader.Read(&buffer)
		if n != 0 {
            gengen.Yield()
        }
		if err == nil {
			continue
        } else if err == io.EOF {
            return nil
        } else {
			return err
        }        
    }
}
```

But wait! 
How does that work? 
We can clearly see that a generator returns a `gengen.Generator`!

Well, this is the second place our `gengen` tag comes into play.
You see, `gengen.Generator` is defined twice:

```go
//go:build !gengen

package gengen

type Generator[T any] interface {
	Next() bool
	Value() T
	Error() error
}
```

```go
//go:build gengen

package gengen

type Generator[T any] error

func (g Generator[T]) Next() bool   { return true }
func (g Generator[T]) Value() T     { return *new(T) }
func (g Generator[T]) Error() error { return nil }
```

The first definition (with `!gengen` as a requirement) contains the actual generator type.
The second one (with `gengen`), uses `gengen.Generator[T]` as an alias for `error`.
This way, the handwritten and the generated code look as if they are using the same types,
when in fact they are not.
This allows both humans and machine to be happy.

----------

The current style is discovery-based and not planning-based.
I think it should be more planned than discovered.

We can start with Python generators, and the way they use `yield` and exceptions.
Then we discuss Go's syntactical limitations, and the lack of exceptions,
to come up with a suitable syntax.
We must also remember that there's a need to stop a generator. Python uses `return`.

This leads us to using a function to emulate `yield`,
and using `return` to terminate iteration.

Then we need to cover errors.
In Python we have `raise` and we can just use exceptions.
Go doesn't have exceptions, so we need a different method.
Luckily, we still have that `return`, which is a great channel for that.

Now we're using a function for `yield` and `return nil` to just return,
`return err` to "raise" an error.


Now for the return type.
Generators return generator objects, so we need to return one here as well.
This means that we return `Generator[T]`. Great.
We also describe the interface at this point and the usage.
This sadly deviates from the `range` iteration pattern, but we can't help it.

Now things seem to be conflicting.
We want to return a generator object, but we also want to return an error.

One option is to use `error` when we write the code, and `Generator[T]` 
in the generated code.
But this means that we can't clearly see the type when we look at our code.
We can do better.
We'll get to that later.

In the meantime, we have a bigger issue - we need to name the generated functions.
One way would be to namespace them. Either with a prefix/suffix, or with a new package.
Both of these options are problematic.
If the generators are in a new package, they won't have access to the same variables.
If we use a prefix/suffix, everything works, but we are forced to see that new name everywhere.
It is both ugly and confusing.

Luckily, Go has build tags.
If we put one build tag in the handwritten code, and the inverse of it in the generated code, 
we have a strong guarantee that there will be no conflicts.
Just need to make sure that what actually compiles is the generated code.

With that done - it's back to the type conflict.
But now, with build tags, we can solve it too!
In the handwritten code, it is an error type.
In the generated code, it is a generator object.
And because the handwritten code never executed, 
we can pretend that the error type has the same interface.


With the syntax decided, we need to go to the actual implementation - what is the structure of a generator?

Here we need to discuss the 2 main parts of a generator - a function, and a state block.
We can show a simple manual example - like Fibonacci - where there is only a single exit/entry.

Once we're done with that example, seeing how things interact, we need to show a more complex example.
Funnily enough, yielding consecutive values is a much more difficult one to pull off.
It also makes it clear that we need to know which line to get to.

We can show a hacky solution around that as well, and then show a more complex one so that the audience will
understand that it is only going to get harder to manage it manually with hacks.

Then we introduce goto. Using goto, with no blocks, we can jump to the relevant lines and execute them.
We introduce the state mechanism, and how each `yield` sets the next entry point.

Then we add actual flow control, and run into issues - `goto` is limited.
It can't jump into blocks, can't jump over variable declarations (need refs here)(probably to start with blocks, and get
to vars later as it is a completely different solution)

This is where we introduce assembly-like flow-control, and "flatten" all the blocks. 
It is important to note that we _are_ allowed to have blocks, just not to jump into them.
So we use `if` blocks for all flow control.

Once we go through `if`, we need to introduce new concepts.
An infinite `for` is possible, but for everything else - we need more.

- c-style loops require variables
- range-style loops require adapters

So we go with an infinite loop and implement fibonacci.
Then we need to discuss the issues with variables, and how we must 
define them all ahead of time. We can also bring up the variable-name-conflict issue.
But that might be better later, if we have an actual example.

With that, we have a working fibonacci generator, and we can use it.

This introduces the most critical concepts. From now on, everything is a bit more optional.

The first thing to introduce is more flow-control structures.
If we want other loop types - we need to discuss `range`.
Since `range` doesn't have an exposed state variable, we need to create one.
We do that using a generator adapter.
Those are different for slices (just an index), maps (convert to slice first),
and chan (I guess I need an adapter that keeps taking values?).

----------

This is the core of the explanation. 
Everything beyond that is implementation details.
I need to see how long this all takes before deciding whether to delve deeper for Berlin.

----------

Then there are a few more things to cover:

- Cool examples to get everyone excited, comparing the generator code to what you'd have to write otherwise
- Benchmarks 
- Future directions, known issues or missing features
- A short pitch for needing generators + iteration protocol in the language
- LIVE DEMO! If I can start a brand-new project, write a generator, run the tool and show it executing
  it'd be absolutely awesome.

----------

## Ideas

Actually, multi-result generators are perfectly possible.
Since `Next()` can have multiple results, and `yield` can take
any number of arguments.
The only tricky bit is getting the generics on `Generator[T]` to work
right, as there are no variadic generics.

The reasonable option is probably to go with:

- `Generator[int]`
- `Generator1[int]`
- `Generator2[int, int]`
- `Generator3[int, int, int]`

To denote the different types of generators.
It is a bit ugly, but should get the job done.

It shouldn't be too hard to pull off in the code,
but it is extra code and can be done once everything else is
already in order.