# Operational Transform (OT) for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/shiv248/operational-transformation-go.svg)](https://pkg.go.dev/github.com/shiv248/operational-transformation-go)
[![CI](https://github.com/shiv248/operational-transformation-go/actions/workflows/ci.yml/badge.svg)](https://github.com/shiv248/operational-transformation-go/actions)
[![codecov](https://codecov.io/gh/shiv248/operational-transformation-go/graph/badge.svg)](https://codecov.io/gh/shiv248/operational-transformation-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/shiv248/operational-transformation-go)](https://goreportcard.com/report/github.com/shiv248/operational-transformation-go)

A Go port of the Rust [`operational-transform`](https://github.com/spebern/operational-transform-rs) library, which itself is a port of [ot.js](https://github.com/Operational-Transformation/ot.js).

## Overview

Operational Transformation (OT) is an algorithm for supporting real-time collaborative editing in distributed systems. It enables multiple users to concurrently edit the same document while maintaining consistency across all copies, without using locks or requiring users to wait for each other.

## Features

- ✅ **Direct port from Rust** - Maintains identical behavior to the battle-tested Rust implementation
- ✅ **UTF-8 character handling** - Correctly counts Unicode codepoints (not bytes)
- ✅ **JSON wire format compatibility** - Compatible with Rust and JavaScript implementations
- ✅ **Complete test coverage** - All tests ported from Rust pass
- ✅ **Type-safe operations** - Go's type system prevents invalid operations

## Installation

```bash
go get github.com/shiv248/operational-transformation-go
```

## Operations

Three basic operations:

- **Retain(n)**: Move cursor forward n positions without changing anything
- **Delete(n)**: Delete n characters at current position
- **Insert(s)**: Insert string at current position

## Usage

```go
import "github.com/shiv248/operational-transformation-go"

// Create operations
op := ot.NewOperationSeq()
op.Retain(5)
op.Insert("world")
op.Delete(3)

// Apply to text
result, err := op.Apply("hello123")
// result = "helloworld"

// Transform concurrent operations
a := ot.NewOperationSeq()
a.Insert("A")

b := ot.NewOperationSeq()
b.Insert("B")

aPrime, bPrime, err := a.Transform(b)
// Both clients converge to same result

// Compose sequential operations
c, err := a.Compose(b)
// c = single operation equivalent to a followed by b

// Invert for undo
inverse := op.Invert("original text")
```

## JSON Serialization

Compatible with Rust/JavaScript wire format:

```go
import "encoding/json"

op := ot.NewOperationSeq()
op.Retain(1)
op.Delete(1)
op.Insert("abc")

data, _ := json.Marshal(op)
// data = [1, -1, "abc"]

var op2 ot.OperationSeq
json.Unmarshal(data, &op2)
```

## Testing

```bash
go test ./...
```

All tests ported from Rust operational-transform pass successfully.

## See It In Action

- [Kolabpad](https://github.com/shiv248/kolabpad) - Real-time collaborative editor using this library
- [OT.js Demo](https://operational-transformation.github.io) - Interactive visualization of Operational Transformation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## References

### Implementations
- [Rust operational-transform](https://github.com/spebern/operational-transform-rs) - Source for this Go port
- [ot.js](https://github.com/Operational-Transformation/ot.js) - Original JavaScript implementation
- [OT FAQ](https://www3.ntu.edu.sg/scse/staff/czsun/projects/otfaq/) - Comprehensive FAQ on Operational Transformation

### Further Reading
- Sun, C., & Ellis, C. (1998). [Operational transformation in real-time group editors](https://dl.acm.org/doi/10.1145/289444.289469). _CSCW '98_, 59-68.
  - Foundational paper on OT theory and algorithms

## License

MIT - Same as the Rust and JavaScript implementations. See [LICENSE](LICENSE) and [NOTICE](NOTICE) for details.
