# Implementation Lineage: Why Port from Rust Rather Than Use ot.go?

> **Purpose**: This document explains the decision to create a new Go implementation by porting from Rust (`operational-transform-rs`) rather than using the existing `ot.go` library. It examines the lineage of OT implementations and the rationale behind choosing a "port of a port" approach.

---

## Table of Contents

1. [The Lineage Tree](#the-lineage-tree)
2. [The Three Implementations](#the-three-implementations)
3. [Why Not Use ot.go?](#why-not-use-otgo)
4. [Why Choose the Rust Implementation?](#why-choose-the-rust-implementation)
5. [Key Differences Between Implementations](#key-differences-between-implementations)
6. [Design Philosophy Comparison](#design-philosophy-comparison)
7. [The Rust Advantage](#the-rust-advantage)
8. [Trade-offs and Considerations](#trade-offs-and-considerations)
9. [Summary](#summary)

---

## The Lineage Tree

The operational transformation ecosystem has evolved from academic theory through multiple language implementations, each building on lessons learned from predecessors:

```
Sun & Ellis (1998) - OT Theory
  ├─ Formal transformation properties
  ├─ Convergence guarantees
  └─ Algorithm foundations
  ↓
ot.js (2012)
  ├─→ ot.go (2015, petejkim)
  │     └─→ Direct JavaScript → Go port
  │         ├─ Server-focused backend
  │         ├─ Limited to ot.js compatibility
  │         └─ Last updated: 2015
  │
  └─→ operational-transform-rs (spebern)
        ├─→ JavaScript → Rust port
        ├─ Modern Rust idioms
        ├─ Type-safe design
        ├─ Active development
        └─→ operational-transformation-go (this project, 2025)
              └─→ Rust → Go port
                  ├─ Modern Go idioms
                  ├─ Complete feature set
                  └─ Active development
```

**The Question**: Why create a new Go implementation (Rust → Go) instead of using the existing one (JavaScript → Go)?

**The Answer**: Architecture, maintainability, type safety, and design philosophy.

---

## The Three Implementations

### 1. ot.js (The Original)

**Language**: JavaScript
**Created**: 2012
**Status**: Seeking new maintainer (as of 2025)
**GitHub**: https://github.com/Operational-Transformation/ot.js

**Characteristics**:
- **Purpose**: Real-time collaborative editing in browsers
- **Architecture**: Client-focused, designed for CodeMirror integration
- **Type system**: Dynamic (JavaScript)
- **Design**: Prototype-based objects, mutable state
- **Wire format**: JSON array notation (`[1, -1, "abc"]`)

**Strengths**:
- Battle-tested in production
- 2.1k GitHub stars
- Interactive visualization and documentation
- Wide ecosystem compatibility

**Weaknesses**:
- JavaScript's dynamic typing makes correctness harder to verify
- Limited static analysis
- Maintenance concerns (project seeking new maintainer)
- Designed for browser, not necessarily optimal for server

### 2. ot.go (The Direct Port)

**Language**: Go
**Created**: 2015 (Nitrous, Inc.)
**Status**: Unmaintained (last update 2015)
**GitHub**: https://github.com/petejkim/ot.go

**Characteristics**:
- **Purpose**: Backend server compatible with ot.js clients
- **Architecture**: Server-side complement to JavaScript clients
- **Type system**: Static (Go)
- **Design**: Direct translation from JavaScript idioms
- **Feature set**: Basic compatibility layer

**Strengths**:
- Native Go server implementation
- ot.js wire format compatibility
- Deployed in production (Nitrous)

**Weaknesses**:
- **Unmaintained**: No updates since 2015 (10 years ago)
- **Limited scope**: Focused on server backend, not standalone library
- **Minimal documentation**: README only shows copyright
- **Architecture**: Direct JavaScript-to-Go translation preserves JavaScript idioms
- **No modern Go features**: Pre-dates Go 1.13+ improvements
- **No test suite visible**: Limited confidence in correctness
- **License concerns**: "Documentation not displayed due to license restrictions" (Go Packages)

### 3. operational-transform-rs (The Rust Rewrite)

**Language**: Rust
**Created**: ~2018
**Status**: Active (196 stars)
**GitHub**: https://github.com/spebern/operational-transform-rs

**Characteristics**:
- **Purpose**: Modern, type-safe OT implementation
- **Architecture**: Standalone library with serde integration
- **Type system**: Static with strong guarantees
- **Design**: Rust ownership model, immutability-first
- **Feature set**: Complete (Transform, Compose, Apply, Invert)

**Strengths**:
- **Type safety**: Rust's type system prevents entire classes of bugs
- **Modern design**: Clean API, ownership semantics
- **Complete feature set**: All OT operations including Invert
- **Serialization**: First-class serde support
- **Active maintenance**: Regular updates
- **Correctness-focused**: Rust enforces safety guarantees

**Weaknesses**:
- Not Go (requires porting for Go projects)
- Rust learning curve for contributors
- Smaller ecosystem than JavaScript

### 4. operational-transformation-go (This Project)

**Language**: Go
**Created**: 2025
**Status**: Active development
**GitHub**: https://github.com/shiv248/operational-transformation-go

**Characteristics**:
- **Purpose**: Modern, complete OT library for Go
- **Architecture**: Standalone library with standard library integration
- **Type system**: Static (Go)
- **Design**: Port of Rust design to Go idioms
- **Feature set**: Complete (Transform, Compose, Apply, Invert)

**Strengths**:
- **Modern Go idioms**: Uses Go 1.13+ features
- **Complete feature set**: All operations from Rust port
- **Type safety**: Inherits Rust's design decisions
- **Wire format compatibility**: Works with ot.js and Rust
- **Comprehensive docs**: Technical documentation in `dev-internals/`
- **Test coverage**: All Rust tests ported and passing
- **Active development**: Regular updates and improvements

**Design Goal**: Bring Rust's type-safe, modern design to the Go ecosystem.

---

## Why Not Use ot.go?

Despite being a direct Go implementation, `ot.go` was not chosen as the foundation for several critical reasons:

### 1. **Maintenance Status**

```
ot.go last commit: 2015
├─ 10 years without updates
├─ Pre-dates Go modules (2019)
├─ Pre-dates Go 1.13+ improvements
└─ No active maintainer
```

**Implications**:
- **Dependency risk**: Unmaintained dependencies pose security and compatibility risks
- **Go evolution**: Misses 10 years of Go language improvements
- **Bug fixes**: No path for community-discovered issues
- **Modern tooling**: Incompatible with current Go development practices

### 2. **Limited Scope**

**ot.go's stated purpose**: "Backend written in Go compatible with ot.js"

```
ot.go focus:
├─ Server-side operations only
├─ Client-server protocol
├─ ot.js JavaScript client compatibility
└─ Not a standalone library
```

**This project's needs**:
```
operational-transformation-go needs:
├─ Standalone library usable in any Go application
├─ No JavaScript client requirement
├─ Complete OT operations (including Invert)
└─ Server OR client use cases
```

**The mismatch**: ot.go was designed as a server backend for JavaScript clients, not a general-purpose OT library.

### 3. **Incomplete Feature Set**

Based on repository analysis:

| Feature | ot.go | operational-transform-rs | This Project |
|---------|-------|--------------------------|--------------|
| Transform | ✅ | ✅ | ✅ |
| Compose | ✅ | ✅ | ✅ |
| Apply | ✅ | ✅ | ✅ |
| Invert | ❓ | ✅ | ✅ |
| Serialization | Basic | serde | encoding/json |
| Test suite | Minimal | Comprehensive | Comprehensive |
| Documentation | Minimal | Good | Extensive |

**Missing documentation**: ot.go's README contains only copyright notice, making feature verification difficult.

### 4. **Architecture: JavaScript Idioms in Go**

**The fundamental issue**: ot.go is a *direct translation* from JavaScript, not a *Go-idiomatic* design.

**Example differences**:

**JavaScript approach (ot.js style)**:
```pseudocode
// Mutable objects, prototype-based
operation = new Operation()
operation.insert("text")
operation.retain(5)

// Dynamic typing
if (operation.ops[0].type === "insert") { ... }
```

**Rust approach (operational-transform-rs style)**:
```pseudocode
// Immutable sequences, strong typing
let mut operation = OperationSeq::default();
operation.insert("text");
operation.retain(5);

// Type safety via enums
match operation.ops[0] {
    Operation::Insert(text) => { ... }
    Operation::Retain(n) => { ... }
    Operation::Delete(n) => { ... }
}
```

**This project (Go idioms from Rust design)**:
```go
// Clear types, interface-based
op := ot.NewOperationSeq()
op.Insert("text")
op.Retain(5)

// Type safety via interface + type assertions
switch o := operation.Ops()[0].(type) {
case Insert:
    // o.Text
case Retain:
    // o.N
case Delete:
    // o.N
}
```

**Why this matters**:
- **Correctness**: Type-safe operations prevent entire classes of bugs
- **Maintainability**: Go idioms feel natural to Go developers
- **Performance**: Avoid JavaScript translation overhead
- **Testing**: Strong typing enables compile-time verification

### 5. **No Visible Test Suite**

**Risk assessment**:

```
ot.go tests: ❓ (repository shows minimal test files)
operational-transform-rs tests: ✅ Comprehensive
This project tests: ✅ All Rust tests ported + additional

Trust level:
├─ ot.go: Low (can't verify correctness)
├─ Rust: High (196 stars, tested in production)
└─ This project: High (inherits Rust test suite)
```

**OT is notoriously tricky**: Small bugs in transform or compose can cause divergence. Without comprehensive tests, confidence is low.

### 6. **License Concerns**

**Go Packages warning**: "Documentation not displayed due to license restrictions"

**Implications**:
- Unclear licensing situation
- Potential legal barriers
- Corporate use concerns

**This project**: Clear MIT license, same as Rust and JavaScript implementations.

---

## Why Choose the Rust Implementation?

Instead of the direct JavaScript→Go path, this project chose JavaScript→Rust→Go. Why add an intermediate step?

### 1. **Type Safety Bridge**

**The challenge**: Translating dynamic JavaScript to static Go requires making type decisions.

```
JavaScript (ot.js):
├─ operation.ops[i] could be anything
├─ Runtime type checking
└─ Easy to make mistakes

Rust (operational-transform-rs):
├─ enum Operation { Insert(String), Retain(u64), Delete(u64) }
├─ Compile-time guarantees
└─ Type decisions already made

Go (this project):
├─ interface Operation { isOperation() }
├─ Concrete types: Insert, Retain, Delete
└─ Inherits Rust's type design
```

**Benefit**: Rust has already solved the "how to make ot.js type-safe" problem. We inherit those decisions.

### 2. **Modern Design Decisions**

**Rust forced good architectural choices**:

| Design Decision | JavaScript (dynamic) | Rust (forced) | This Project (inherited) |
|----------------|---------------------|---------------|---------------------------|
| Mutation | In-place mutation | Ownership rules | Clear mutability |
| Validation | Runtime checks | Type system | Compile-time checks |
| Error handling | Exceptions/undefined | Result<T, E> | error returns |
| Memory safety | GC + runtime errors | Compile-time | GC + type safety |

**Example - Operation merging**:

```rust
// Rust: Ownership prevents accidental mutation
impl OperationSeq {
    pub fn insert(&mut self, text: &str) {
        // Compiler enforces: either mut or shared, not both
    }
}
```

```go
// Go: Port Rust's design pattern
func (o *OperationSeq) Insert(s string) {
    // Follows Rust's mutability pattern
    // Clear ownership semantics
}
```

### 3. **Complete Feature Set**

**Comparison**:

```
ot.js:
├─ Transform ✅
├─ Compose ✅
├─ Apply ✅
└─ Invert ✅

ot.go:
├─ Transform ✅
├─ Compose ✅
├─ Apply ✅
└─ Invert ❓

operational-transform-rs:
├─ Transform ✅
├─ Compose ✅
├─ Apply ✅
└─ Invert ✅

This project:
├─ Transform ✅ (from Rust)
├─ Compose ✅ (from Rust)
├─ Apply ✅ (from Rust)
└─ Invert ✅ (from Rust)
```

**Invert is critical**: Undo/redo functionality requires operation inversion. ot.go's lack of visible Invert implementation is a significant gap.

### 4. **Active Development & Community**

**Development activity**:

```
ot.js: Seeking maintainer (dormant)
ot.go: Last update 2015 (unmaintained)
operational-transform-rs: Active (regular commits, 196 stars)
This project: Active (2025+)
```

**Community confidence**:
- **Rust implementation**: 196 stars, used in production Rust projects
- **Battle-tested**: Real-world usage validates design decisions
- **Modern best practices**: Benefits from recent OT research

### 5. **Correctness Confidence**

**The OT correctness challenge**:

```
Transform must satisfy:
  apply(apply(S, A), B') == apply(apply(S, B), A')

One bug breaks convergence.
```

**Trust chain**:

```
ot.js → Rust:
  ├─ Rust compiler catches type errors
  ├─ Ownership prevents mutation bugs
  ├─ Pattern matching ensures exhaustiveness
  └─ Tests validate semantics

Rust → Go:
  ├─ Type design preserved
  ├─ All tests ported
  ├─ Behavior verified identical
  └─ Go's type system provides safety
```

**Why trust Rust as source**:
1. **Compile-time verification**: Rust catches bugs at compile time
2. **Test suite**: Comprehensive tests from ot.js all passing
3. **Production usage**: Real-world validation
4. **Type safety**: Prevents entire bug classes

### 6. **Better Foundation for Evolution**

**Design headroom**:

```
ot.go (JavaScript idioms):
├─ Tightly coupled to JavaScript patterns
├─ Hard to extend without breaking idioms
└─ Mutation patterns from JavaScript

Rust design:
├─ Clean separation of concerns
├─ Clear ownership and lifetimes
├─ Extensible architecture
└─ This project inherits this foundation
```

**Example - Adding new features**:

Rust's clear design makes it easy to understand *why* things work:

```rust
// Rust: Clear iterator pattern for transform
let mut iter_a = ops_a.iter().peekable();
let mut iter_b = ops_b.iter().peekable();

// Easy to understand and port to Go
```

---

## Key Differences Between Implementations

### 1. **Type System Architecture**

**JavaScript (ot.js)**:
```javascript
// Dynamic typing
{
  ops: [
    5,           // Retain(5)
    -3,          // Delete(3)
    "hello"      // Insert("hello")
  ]
}

// Runtime type checking required
if (typeof op === 'string') { /* insert */ }
else if (op > 0) { /* retain */ }
else { /* delete */ }
```

**Rust (operational-transform-rs)**:
```rust
// Strong enums
enum Operation {
    Retain(u64),
    Delete(u64),
    Insert(String),
}

// Compile-time exhaustiveness
match op {
    Operation::Retain(n) => { ... }
    Operation::Delete(n) => { ... }
    Operation::Insert(s) => { ... }
    // Compiler error if any case missing!
}
```

**Go (this project)**:
```go
// Interface + concrete types
type Operation interface {
    isOperation()
}

type Retain struct { N uint64 }
type Delete struct { N uint64 }
type Insert struct { Text string }

// Type-safe switches
switch o := op.(type) {
case Retain:
    // o.N
case Delete:
    // o.N
case Insert:
    // o.Text
}
```

### 2. **Serialization Design**

**All implementations** use the same wire format for compatibility:
```json
[1, -1, "abc"]
```

**Implementation differences**:

| Implementation | Serialization | Deserialization |
|---------------|---------------|-----------------|
| ot.js | Native JSON | Native JSON parsing |
| ot.go | Custom marshaling | Custom unmarshaling |
| Rust | serde derive macros | serde derive macros |
| This project | encoding/json | encoding/json |

**This project's advantage**: Standard library integration means compatibility with existing Go JSON tooling.

### 3. **Error Handling**

**JavaScript**:
```javascript
// Exceptions or undefined
function transform(a, b) {
    if (a.baseLen !== b.baseLen) {
        throw new Error("Incompatible lengths");
    }
}
```

**Rust**:
```rust
// Result type (explicit)
fn transform(&self, other: &OperationSeq)
    -> Result<(OperationSeq, OperationSeq), OTError>
{
    if self.base_len() != other.base_len() {
        return Err(OTError::IncompatibleLengths);
    }
    // ...
}
```

**Go (this project)**:
```go
// Multiple return values (idiomatic)
func (o *OperationSeq) Transform(other *OperationSeq)
    (*OperationSeq, *OperationSeq, error)
{
    if o.BaseLen() != other.BaseLen() {
        return nil, nil, ErrIncompatibleLengths
    }
    // ...
}
```

### 4. **Memory Management**

**JavaScript**:
```
├─ Garbage collected
├─ Reference semantics
└─ No ownership concept
```

**Rust**:
```
├─ Ownership model
├─ Compile-time lifetime tracking
└─ Zero-cost abstractions
```

**Go (this project)**:
```
├─ Garbage collected (like JavaScript)
├─ Value/pointer semantics (explicit)
└─ Simpler than Rust, safer than JavaScript
```

**Why this matters**: Go gets JavaScript's ease-of-use with more explicit control than JavaScript provides.

### 5. **Unicode Handling**

All implementations count **Unicode codepoints**, not bytes. Implementation differs:

**JavaScript**:
```javascript
// JavaScript strings are UTF-16
text.length  // Counts UTF-16 code units
// Works for most characters, issues with surrogate pairs
```

**Rust**:
```rust
// Rust strings are UTF-8
text.chars().count()  // Counts Unicode codepoints
// Correct for all Unicode
```

**Go (this project)**:
```go
// Go strings are UTF-8
utf8.RuneCountInString(text)  // Counts Unicode codepoints
// Correct for all Unicode
```

**Advantage**: Both Rust and Go handle Unicode correctly by default.

---

## Design Philosophy Comparison

### JavaScript (ot.js): Pragmatic Simplicity

**Philosophy**: Make it work in browsers, optimize for developer ease.

**Characteristics**:
- ✅ Quick to prototype
- ✅ Flexible (sometimes too flexible)
- ⚠️ Runtime verification only
- ❌ Easy to misuse

**Design priorities**:
1. Browser compatibility
2. Small bundle size
3. Ease of integration

### Go Direct Port (ot.go): Compatibility Focus

**Philosophy**: Provide Go backend for JavaScript clients.

**Characteristics**:
- ✅ Works with ot.js
- ✅ Native Go performance
- ⚠️ Limited scope
- ❌ JavaScript idioms in Go

**Design priorities**:
1. ot.js protocol compatibility
2. Server-side operations
3. Minimal translation from JavaScript

### Rust (operational-transform-rs): Correctness First

**Philosophy**: Make incorrect usage impossible to compile.

**Characteristics**:
- ✅ Compile-time guarantees
- ✅ Impossible to misuse (if it compiles, it's likely correct)
- ✅ Performance + safety
- ⚠️ Steeper learning curve

**Design priorities**:
1. Type safety
2. Compile-time verification
3. Zero-cost abstractions
4. Impossible to misuse

### This Project: Go Idioms + Rust Safety

**Philosophy**: Bring Rust's safety guarantees to Go's simplicity.

**Characteristics**:
- ✅ Type-safe operations
- ✅ Go-idiomatic API
- ✅ Comprehensive tests
- ✅ Easy to use correctly

**Design priorities**:
1. Go idioms (feels natural to Go developers)
2. Type safety (from Rust design)
3. Complete features (all OT operations)
4. Maintainability (clear code, good docs)

---

## The Rust Advantage

Why Rust as the "translation layer" between JavaScript and Go?

### 1. **Type Safety Forcing Function**

**JavaScript → Rust** required solving:
- How to represent operations type-safely?
- How to prevent invalid operation sequences?
- How to enforce length invariants?

**Rust → Go** simply ports the solutions:
- Operations are an interface with concrete types ✅
- OperationSeq tracks invariants ✅
- Validation at compile time where possible ✅

**Alternative path (JavaScript → Go directly)**:
- Every type decision made from scratch
- Easy to miss edge cases
- No battle-tested design to follow

### 2. **Ownership Clarifies Semantics**

**Rust ownership** forces explicit decisions about mutation:

```rust
// Rust: Must declare mutability
let mut op = OperationSeq::default();
op.insert("text");  // OK: op is mut

let op2 = op;  // Move! op is now invalid
op.insert("more");  // Compiler error!
```

This clarifies **when mutations are allowed**, which translates to clear Go semantics:

```go
// Go: Explicit pointer vs value
op := ot.NewOperationSeq()
op.Insert("text")  // OK: receiver is pointer

op2 := op  // Pointer copy
op.Insert("more")  // Both op and op2 affected
```

**Benefit**: No ambiguity about mutation behavior.

### 3. **Pattern Matching Ensures Completeness**

**Rust transform algorithm**:

```rust
match (comp_a, comp_b) {
    (Operation::Insert(s1), Operation::Insert(s2)) => { ... }
    (Operation::Insert(s), Operation::Retain(n)) => { ... }
    (Operation::Insert(s), Operation::Delete(n)) => { ... }
    (Operation::Retain(n), Operation::Insert(s)) => { ... }
    (Operation::Retain(n1), Operation::Retain(n2)) => { ... }
    (Operation::Retain(n), Operation::Delete(m)) => { ... }
    (Operation::Delete(n), Operation::Insert(s)) => { ... }
    (Operation::Delete(n), Operation::Retain(m)) => { ... }
    (Operation::Delete(n1), Operation::Delete(n2)) => { ... }
}
// Compiler error if any case missing!
```

**Rust compiler enforces**: All 9 cases must be handled.

**Go port benefits**: We know we have all cases covered because Rust verified it.

### 4. **Standard Patterns**

**Rust ecosystem** has established patterns for:
- Iterator design
- Error handling (Result type)
- Serialization (serde)
- Testing (cargo test)

**This project** adapts these to Go:
- Iterator → simple index iteration
- Result → multiple return values
- serde → encoding/json
- cargo test → go test

**Advantage**: Well-understood patterns translated to Go idioms.

---

## Trade-offs and Considerations

### Why This Approach Works

**Strengths**:
1. ✅ **Type safety**: Inherits Rust's design decisions
2. ✅ **Completeness**: All features from battle-tested implementation
3. ✅ **Maintainability**: Modern Go code, not JavaScript idioms
4. ✅ **Correctness**: Comprehensive test suite ported
5. ✅ **Documentation**: Extensive technical docs
6. ✅ **Active development**: Not abandoned

**Trade-offs**:
1. ⚠️ **One more translation**: Rust → Go adds indirection
2. ⚠️ **Not 1:1 with ot.js**: Different from JavaScript direct port
3. ⚠️ **Smaller ecosystem**: Less community than ot.js

### When to Use This vs. ot.go

**Use this project if you want**:
- ✅ Complete OT library (including Invert)
- ✅ Modern Go codebase
- ✅ Active maintenance
- ✅ Type-safe operations
- ✅ Comprehensive documentation
- ✅ Standalone library (not tied to JavaScript clients)

**Use ot.go if you**:
- Specifically need 2015-era compatibility
- Have existing ot.go integration
- *(Note: ot.go is unmaintained and not recommended for new projects)*

### Alternative: Direct ot.js Port

**Could we have ported ot.js directly to Go?**

Yes, but:
- ❌ Lose type-safety guidance from Rust
- ❌ JavaScript idioms don't translate cleanly to Go
- ❌ Would repeat all the design work Rust already did
- ❌ No forcing function for correctness
- ✅ Slightly fewer translation steps

**The Rust intermediate step** provides:
- ✅ Type-safe design already worked out
- ✅ Ownership semantics clarify mutation
- ✅ Compile-time verification of completeness
- ✅ Battle-tested in production
- ⚠️ One more translation step (Rust → Go)

**Verdict**: The benefits of Rust's type safety outweigh the cost of an extra translation step.

---

## Summary

### The Decision

**Instead of using ot.go (JavaScript → Go, 2015)**, this project chose to port from Rust (JavaScript → Rust → Go, 2025).

### Key Reasons

1. **Maintenance**: ot.go unmaintained since 2015; Rust implementation actively developed
2. **Completeness**: Rust has full feature set including Invert; ot.go scope limited
3. **Type Safety**: Rust's type system solved the "how to make OT type-safe" problem
4. **Architecture**: Rust design is modern and extensible; ot.go preserves JavaScript idioms
5. **Correctness**: Rust's comprehensive test suite provides confidence
6. **Documentation**: Rust implementation well-documented; ot.go has minimal docs

### The Lineage Advantage

```
JavaScript (ot.js)
    ↓
Rust (operational-transform-rs)
    ↓ [Type safety enforced]
    ↓ [Design decisions validated]
    ↓ [Tests verified]
    ↓
Go (this project)
    ↓ [Go idioms applied]
    ↓ [Standard library integration]
    ↓ [Comprehensive documentation]
```

**Value of Rust intermediate step**:
- Type-safe design worked out
- Ownership semantics clarified
- Compile-time correctness verification
- Battle-tested in production

### Comparison Summary

| Aspect | ot.go | This Project |
|--------|-------|--------------|
| **Maintenance** | Unmaintained (2015) | Active (2025+) |
| **Design source** | JavaScript idioms | Rust type-safe design |
| **Feature set** | Basic compatibility | Complete OT operations |
| **Type safety** | Basic | Strong (from Rust) |
| **Tests** | Minimal | Comprehensive |
| **Documentation** | Minimal | Extensive |
| **Scope** | Server backend | Standalone library |
| **Go idioms** | JavaScript-influenced | Modern Go patterns |

### The Result

This project successfully brings:
- ✅ **Modern Go idioms**: Feels natural to Go developers
- ✅ **Rust's type safety**: Design decisions that prevent bugs
- ✅ **Complete features**: All OT operations including Invert
- ✅ **Wire compatibility**: Works with ot.js and Rust implementations
- ✅ **Maintainability**: Active development, comprehensive docs
- ✅ **Correctness**: Battle-tested design with full test coverage

### When Rust as Intermediate Works

**This pattern (Language A → Rust → Language B) works well when**:
1. Source language (JavaScript) is dynamically typed
2. Target language (Go) is statically typed
3. Rust has solved the type-safety translation
4. Rust implementation is well-tested
5. Target language has similar abstractions to Rust

**For OT specifically**: Rust's ownership model and exhaustive pattern matching make it an ideal intermediate representation for translating dynamic JavaScript OT to static Go OT.

---

### Related Documentation

- **[01-operations.md](01-operations.md)** - Core OT operations and design
- **[02-transform.md](02-transform.md)** - Transform algorithm details
- **[04-serialization.md](04-serialization.md)** - Wire format compatibility

### Academic Foundations

- **Sun, C., & Ellis, C. (1998)**. [Operational transformation in real-time group editors](https://dl.acm.org/doi/10.1145/289444.289469). _Proceedings of the 1998 ACM Conference on Computer Supported Cooperative Work_ (CSCW '98), 59-68.
  - The seminal paper that established the theoretical foundations and formal properties of Operational Transformation
  - Defines the convergence guarantees that all implementations (ot.js, Rust, and this Go port) rely upon

### External References

- [ot.js](https://github.com/Operational-Transformation/ot.js) - Original JavaScript implementation
- [ot.go](https://github.com/petejkim/ot.go) - Direct JavaScript→Go port (2015)
- [operational-transform-rs](https://github.com/spebern/operational-transform-rs) - Rust implementation
- [This project](https://github.com/shiv248/operational-transformation-go) - Rust→Go port (2025)

---

**Questions or Issues?**

If you have questions about this design decision or suggestions for improvement, please file an issue in the repository.
