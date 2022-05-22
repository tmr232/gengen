# gengen
Generating Generators!


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