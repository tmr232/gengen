1:12 - Introducing myself
    Messily talking about starting with Go, and how nice simplicity is
3:09 - I still miss generators, and I want to show you why it's important
3:20 - Start talking about my workflow when working with new APIs
    This is _very_ messy here.
    Need a proper code sample to actually talk about.
    I just spouted nonsense here.
4:20 - Time to write business logic
    Explaining the choices - slices vs. iterators vs. duplicating code
    Need to make it clear that duplication is not a good thing - a bit messy now.
    Need to explain why slices are bad - tighten it.
5:30 - Implementing Iterators
    Introducing the iterator interface.
    Showing iteration
6:30 - Actually implementing
    5 parts of implementing an iterator
    ctor talk is a bit messy
7:00 - SingleFunctionIterator
    A bit messy. What does "duplicated" mean?
    Showing the "Advance()" function
    This explanation is _really_ messy at the moment.
    Need to tighten the transformation & mention the closure
8:40 - Implementing the actual iterator
    Need actual code
    This is messy because I didn't have the code with me.
    Should probably show a nested loop to simplify things
9:40 - Iterator is done, what are the issues?
    This is a bit messy.
    Multiple interfaces.
    Closures vs. Structures
10:20 - Hard to implement and maintain.
    This is very messy still. 
    Need to work this out.
11:40 - Show the code with generators
    See how it is almost identical to the printing loop
    Need to tighten the explanation
    **Should probably explain generator syntax directly**
13:00 - Generator Semantics
    Work this over
    Possibly show directly on the previous example
14:34 - start with implementation
15:54 - end of 20 second pause
    We'll do code transformations
    Go has the tools, but we don't discuss them today
    Transforming a yield 1,2,3 generator
    A bit messy with the explanation of "Advance()".
    Maybe we need a better vocabulary?
    Tighten the transitions here
    Need to add visualization of the flow!
17:30 - If
    We have blocks - can't use goto directly
    Tighten the goto issues
    Rework wording on the code->graph transition. It's a mess
    Maybe show fake goto's in the original code?
20:00 - For
    I need to explain that I'll only cover _some_ structures, due to timing.
    Again - a bit messy and need to be tightened
21:55 - Range For
    Adaptors.
    Still messy.
    **Should probably explain variable issues and block issues when we introduce goto, then move to control flow**
23:25 - Maintaining State
    Initialization of variables and maintaining values
    Issues with goto - feels a bit messy
    Move all definitions outside the closure - get value persistence!
    Variable Naming
24:50 - Return statements - explain them!
    This is probably a bad place, and should be in the first iterator.
25:20 - end of 20 second break
25:28 - Code generation
    We write fake code
    We need to make sure we don't build it
    Using built tags
    Need to discuss go-generate before the build tag
    Mention that the code is copied
    Mention that the gengen tag is inversed for the generated code
27:28 - "Demo" and clapping
End of talk.


Current Timing Estimates

## Intro [12 min]
[2 min] About me
[1 min] My workflow
[1 min] The need for iterators
[4 min] Implementing Iterators (incl. SingleFunctionIterator)
[2 min] Issues with Iterators
[2 min] Generator expo

## Generators [15 min]
[2 min] Generator Semantics
[3 min] Intro to goto
[6 min] Control Flow
[1 min] State
[2 min] Code Generation
[1 min] Demo & Clapping


From Mat Ryer's talk (GopherCon EU 2019)
- Maintainability
- Glancability
- Obvious Code
- Clear to new Gophers or people new to the project
- Self-similar code

- Add disclaimer for code in the talk
  - Simplified to fit into the slide
  - Simple examples for time's sake
- "End up with a ctor, because there is setup that needs to happen"


# New Flow After Guy's Feedback

- Start with intro
- Missing feature from Python - Generators
- Describe what Generators are - a simple and straightforward way to write Iterators
- Describe my workflow
- Code sample with simple API, mention that it is simplified
- Iterators are too complicated and hard to maintain
  - Lost of code, many ways to do it, breaks the logic in non-obvious way
  - No self-similarity
- Show generator solution & explain generator syntax
- From here on - how to get this code (generators) working TODAY
- Explain that the solution is using Code Generation
- Because the syntax I showed is not Go.
- We'll use go-generate and build-tags
- Allow us to transparently separate our Generator definitions from the actual, executed code
- build tag, go-generate line
- We have - what are generators, we do code generation
- Show that our `Yield` is a no-op
- So we'll discuss the Code Transformations required to make them work in Go
- We're using Go tools for analysis & generation, which are awesome.
- Start with the actual `return nil` because error reporting is easy and we want it
    out of the way.
- Start with a simple example and show `goto`
- We need `goto` to return to the line after `Yield`
- We need to mark that line - so we use labels - the easiest way
- Goto's jump to labels, no need to mention other solutions
- Issues with `goto`
- Safety - no blocks, no var-defs
- So we have 2 issues - variable defs, jumping into blocks
- Var-defs - move them into our state-block. Make sure we mention it when discussing iterators!
- All gotos are after the defs, all state is maintained!
- Probably no time to discuss name conflicts
- More complex - blocks!
- Start with `if`
- We cannot jump into it, so we can use `goto` to restructure the `if`
- We can see the graph for the `if`, and use `goto` to reposition parts of it
- Start by marking the parts with labels, adding `goto` to show the original flow
- Then moving the code out of the blocks and add the `goto` to redirect
- With `for` we need to show backlink, `break`, `continue`
- condition for->show the differences
- c-style - lucky to have vars out of the way already!
- Show it as condition for, mention differences in `break` and `continue`
- `range-for` - explain adaptors, transform to condition-for or **c-style**
- We perform this exercise for the rest of the control flow structures in Go
- We have generators, generation, limitations of goto, transformations, error reporting!!!
- Show final example, generated code, results with go-test

# Code Manipulates Data

In essence, code is meant to manipulate data.
So when working with new APIs, I want to know the data I'll be working with.
Documentation and structures are nice, but I find that seeing concrete data
makes it much easier for me to reason about it and know what I need to do.
So when dealing with new APIs, my first order of business is querying for
data and displaying it. Usually using `fmt.PrintLn`.

# Workflow up to Iterators

The first thing we do is query and print the data.
Structures and documentation are nice but seeing the actual helps
me understand it a whole-lot better.
Once we have that, we want to start manipulating the data and write our business logic.
When we do that - we need to make some code choices.
3 options - one is to put the logic in our query function.
Gets tricky as soon as we add filtering, pagination, etc.
We don't want to duplicate it.
Additionally, it puts 2 logically separate actions in the same function and makes it harder to
reason about it.
The second option - query first, manipulate later.
We can query all the data, put it in a slice, and manipulate the slice.
This works for small enough or cheap-enough-to-get data.
When we're making network requests - we'd rather only get the data we want.
The third option, and the one I prefer, is to use iterators.
This allows us to only get the data we need, and iterate over it like a slice.
This is where I show iteration code and explain the interface.
The issue comes from implementing them.
Generally 5 parts.

# Iterators

first - maintain state. Usually done using a struct or a closure.
*Really messy here*
Also need a ctor due to non-obvious initialization
`Value` and `Err` methods as trivial getters
`Next` which is the core of the iterator, and handles the actual iteration.

Instead, we'll use a helper struct, and implement our iterators using closures.