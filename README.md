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