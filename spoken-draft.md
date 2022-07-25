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
