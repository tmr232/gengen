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

```go
// Source
func If(flag bool) gengen.Generator[string] {
    if flag {
        yield("true")
    } else {
		yield("false")
    }
	
	return nil
}

// Generated
func If(flag bool) gengen.Generator[string] {
	__next := 0
	return &gengen.GeneratorFunction[string]{
		Advance: func() (hasValue bool, value string, err error) {
			switch __next {
			case 0:
				goto __next_0
            case 1:
				goto __next_1
            }   
			
			__next_0:
				// if flag
				if flag {
				    goto __then_0
                } else {
					goto __else_0
                }               
            
            __then_0:
				// yield true
				__next = 1
				return true, "true", nil
				
            __else_0:
				// yield false
				__next = 1
				return true, "false", nil
				
            __next_1:
				// return
				return
        }
	}
}
```