# Translation Tables

Here we describe roughly what every instruction or structure gets translated to!

```go
// Source
func Empty() gengen.Generator[int] {
    return nil
}

// Generated
func Empty() gengen.Generator[int] {
	return &gengen.GeneratorFunction[int]{
		Advance: func() (hasValue bool, value int, err error) {
			return
        }
	}
}
```

```go
// Source
func Error() gengen.Generator[int] {
    return SomeError{}
}

// Generated
func Error() gengen.Generator[int] {
	return &gengen.GeneratorFunction[int]{
		Advance: func() (hasValue bool, value int, err error) {
			return false, *new(int), SomeError{}
        }
	}
}
```

```go
// Source
func Yield() gengen.Generator[int] {
    yield(42)
	return nil
}

// Generated
func Yield() gengen.Generator[int] {
	__next := 0
	return &gengen.GeneratorFunction[int]{
		Advance: func() (hasValue bool, value int, err error) {
			switch __next {
			case 0:
				goto __next_0
            case 1:
				goto __next_1
            }   
			
			__next_0:
				__next = 1
				return true, 42, nil
            
            __next_1:
				return 
        }
	}
}
```