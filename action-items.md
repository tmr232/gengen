- Choose proper code samples for all slides
- Evaluate the usefulness of `Advance()` and relevant vocabulary
- Start with explaining `goto` issues _before_ explaining control-flow
- Guy suggests saying what generators are (a simple way to write iterators)
  right at the state - before the workflow demo
- Need to make it clear that we're generating code, and not wait for the end with it.
- Guy had no idea that this is what we're actually doing.
- We need to move the build-tag and go-generate part to the top, before code transformations.
- That will allow us to explain _why_ we're doing code transformations and where they come in.

## Why Code Transformations
This generator syntax is nice and all, but this is not Go.
We currently have no generator functions, and our `gengen.Yield` function is a no-op

```go
package WhoCares
func Yield[T any](T) {}
```

To actually get our generator function to work, we need to generate real Go code from it.

## Explaining the code transformations

- We need to transform yield into 2 parts - yield value, and note where to go next.
- yielding the value is done by returning the right values from `Next()`, `Value()`, and `Err()`.
- Then, we place a label right after our `yield` so that we know where to go next.
- Place `__next` in our "state block" to maintain it
- Use `switch` to decide which label to `goto` to
- As we'll see in a moment - `goto` in Go, being a safe language, has two limitations we need to handle
  1. It cannot jump over variable definitions. If it did - the variable would be in an un-initialized 
     state and we'd lose safety.
  2. It cannot jump into a block. If it did - we won't know whether the block's invariants hold,
     and may have all sorts of bugs.
- First, we handle the first issue. We do this by moving _all_ variable definitions outside the closure
  and into our state-block. If we have no variable definitions - there's nothing to jump over.
- Additionally, this guarantees that state will be maintained on consective calls to the function,
  as it is captured by the closure and not defined in it.
- **We're ignoring the renaming issue. Audience is free to ask later.**
- The next issue we need to resolve is the block-issue.
- All control-flow in Go is done using blocks, which is a good thing.
- Structure and nesting are _very_ useful.
- But as we moved variable definitions outside our blocks, they no longer serve a purpose other
  than control flow.
### From Code to Control Flow Graph
- To reason about control flow more easily, we'll use a graph form.
- The following `if` statement can be represented in the following graph.
```go
if alpha {
	gengen.Yield("a")
	gengen.Yield("b")
	gengen.Yield("c")
} else {
	gengen.Yield("1")
	gengen.Yield("2")
	gengen.Yield("3")
}


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
- We can add a few labels without changing the flow
- Then, we add `goto`s that match the current flow
- And finally, we can just move the code outside the blocks,
  leaving only a `goto label` behind for the redirection.
