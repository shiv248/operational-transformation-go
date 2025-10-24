# Operations Fundamentals

> **Purpose**: This document introduces the three basic Operational Transformation (OT) operations and explains the OperationSeq structure. This is the entry point for understanding OT in this codebase.

---

## Table of Contents

1. [What is Operational Transformation?](#what-is-operational-transformation)
2. [The Three Operations](#the-three-operations)
3. [OperationSeq Structure](#operationseq-structure)
4. [Operation Merging (Optimization)](#operation-merging-optimization)
5. [Unicode Codepoint Handling](#unicode-codepoint-handling)
6. [Creating Operations](#creating-operations)
7. [Noop Detection](#noop-detection)
8. [Summary](#summary)

---

## What is Operational Transformation?

**Operational Transformation (OT)** is an algorithm for real-time collaborative editing. It enables multiple users to edit the same document concurrently while maintaining consistency across all copies.

**Key characteristics**:
- **Concurrent editing**: Multiple users can edit simultaneously without locks
- **Automatic conflict resolution**: The algorithm resolves conflicting edits deterministically
- **Eventual consistency**: All users converge to the same final document state
- **Alternative to CRDTs**: Different approach to distributed consensus for text editing

**This implementation**:
- Direct port from Rust [`operational-transform`](https://github.com/spebern/operational-transform-rs) crate
- Compatible with JavaScript [ot.js](https://github.com/Operational-Transformation/ot.js)
- Battle-tested algorithm used in production collaborative editors

For a high-level architectural overview of how OT fits into a collaborative editing system, see the architecture documentation.

---

## The Three Operations

All text transformations in OT are built from three primitive operations. Each operation describes a change to a document relative to a cursor position.

### Retain(n)

**Purpose**: Move cursor forward n positions without changing the text.

```pseudocode
OPERATION Retain(n):
    cursor_position += n
    // Text remains unchanged
```

**Use cases**:
- Skipping unchanged portions of text when describing an edit
- Positioning cursor before making changes
- Expressing "no change" for a section of the document

**Example**:
```pseudocode
// To insert at position 10 in a document
operation = [Retain(10), Insert("hello")]

// This means: skip 10 characters, then insert "hello"
```

**Implementation**:
```go
type Retain struct {
    N uint64  // Number of positions to move forward
}
```

### Delete(n)

**Purpose**: Remove n characters at the current cursor position.

```pseudocode
OPERATION Delete(n):
    remove n characters starting at cursor_position
    // Cursor stays at same position after deletion
```

**Use cases**:
- Removing text from the document
- Deleting user selections
- Replacing text (Delete followed by Insert)

**Example**:
```pseudocode
// To delete characters 5-10 in a document
operation = [Retain(5), Delete(5)]

// This means: skip 5 characters, then delete the next 5
```

**Important**: The cursor does NOT advance after a Delete. It remains at the same position, which is now pointing to what was previously the character after the deleted range.

**Implementation**:
```go
type Delete struct {
    N uint64  // Number of characters to delete
}
```

### Insert(text)

**Purpose**: Insert text at the current cursor position.

```pseudocode
OPERATION Insert(text):
    insert text at cursor_position
    cursor_position += length(text)
```

**Use cases**:
- Adding new text to the document
- User typing
- Pasting content

**Example**:
```pseudocode
// To insert "hello" at position 5
operation = [Retain(5), Insert("hello")]

// After insert, cursor is at position 5 + length("hello") = 10
```

**Implementation**:
```go
type Insert struct {
    Text string  // String to insert
}
```

---

## OperationSeq Structure

An **OperationSeq** is a sequence of operations that together describe a complete transformation of a document. It's more than just a list - it tracks critical metadata about the transformation.

### Core Concept

```pseudocode
OperationSeq:
    operations: List of (Retain | Delete | Insert)
    baseLen: Length of input text required
    targetLen: Length of output text after applying operations
```

### Why Track baseLen and targetLen?

These two values are essential for validation and composition:

**1. Validation** - Ensure operation can be applied
```pseudocode
IF operation.baseLen != length(input_text):
    ERROR "Cannot apply operation to text of wrong length"
```

**2. Composition** - Check if operations can be composed
```pseudocode
// To compose A followed by B:
IF A.targetLen != B.baseLen:
    ERROR "Cannot compose: output of A doesn't match input of B"
```

**3. Transformation** - Verify operations are concurrent
```pseudocode
// To transform concurrent operations A and B:
IF A.baseLen != B.baseLen:
    ERROR "Operations not concurrent: different base states"
```

### Calculating baseLen and targetLen

The lengths are computed from the operations:

```pseudocode
FOR EACH operation IN operations:
    CASE operation:
        Retain(n):
            baseLen += n      // Consuming n chars from input
            targetLen += n    // Producing n chars in output

        Delete(n):
            baseLen += n      // Consuming n chars from input
            targetLen += 0    // Producing nothing (deleted)

        Insert(text):
            baseLen += 0      // Consuming nothing from input
            targetLen += length(text)  // Producing new chars
```

### Example

```pseudocode
// Original text: "hello" (length 5)
operation = [Retain(2), Insert("x"), Delete(1), Retain(2)]

// Calculate baseLen:
baseLen = 2 (Retain) + 0 (Insert) + 1 (Delete) + 2 (Retain) = 5

// Calculate targetLen:
targetLen = 2 (Retain) + 1 (Insert) + 0 (Delete) + 2 (Retain) = 5

// This operation transforms a 5-character string to a 5-character string
```

**Applying to "hello"**:
```pseudocode
Step 1: Retain(2)
    cursor at 2, result: "he"

Step 2: Insert("x")
    cursor at 3, result: "hex"

Step 3: Delete(1)
    cursor at 3 (deleted "l"), result: "hex"

Step 4: Retain(2)
    cursor at 5, result: "hexlo"

Final result: "hexlo"
```

### Implementation

```go
type OperationSeq struct {
    ops       []Operation  // Sequence of operations
    baseLen   int         // Required input length
    targetLen int         // Resulting output length
}

// Accessors
func (o *OperationSeq) BaseLen() int
func (o *OperationSeq) TargetLen() int
func (o *OperationSeq) Ops() []Operation
```

---

## Operation Merging (Optimization)

When building operations incrementally (like a user typing), a naive approach would create many tiny operations. OT implementations merge consecutive operations of the same type for efficiency.

### The Problem

```pseudocode
// User types "hello" one character at a time
Insert("h")
Insert("e")
Insert("l")
Insert("l")
Insert("o")

// This creates 5 separate Insert operations
operations = [Insert("h"), Insert("e"), Insert("l"), Insert("l"), Insert("o")]
```

### The Solution: Automatic Merging

```pseudocode
// With merging enabled
Insert("h")  // operations = [Insert("h")]
Insert("e")  // operations = [Insert("he")]    ← merged!
Insert("l")  // operations = [Insert("hel")]   ← merged!
Insert("l")  // operations = [Insert("hell")]  ← merged!
Insert("o")  // operations = [Insert("hello")] ← merged!

// Result: Single operation
operations = [Insert("hello")]
```

### Insert Merging Algorithm

Insert merging is the most complex because Inserts must always appear before Deletes in the canonical operation order.

```pseudocode
FUNCTION Insert(text):
    IF last operation is Insert:
        // Simple case: merge with last Insert
        lastInsert.text += text

    ELSE IF last operation is Delete AND second-to-last is Insert:
        // Merge with Insert before the Delete
        // Maintains canonical order: Insert before Delete
        secondToLast.text += text

    ELSE IF last operation is Delete:
        // Insert BEFORE the Delete (swap)
        // This maintains the invariant that Inserts come before Deletes
        insert_operation_before_delete(text)

    ELSE:
        // Default: append new Insert
        append(Insert(text))
```

**Why the complex logic?** OT maintains a canonical order where Insert operations always appear before Delete operations at the same cursor position. This ensures deterministic operation sequences.

### Delete Merging Algorithm

Delete merging is straightforward:

```pseudocode
FUNCTION Delete(n):
    IF last operation is Delete:
        lastDelete.n += n  // Merge with last Delete
    ELSE:
        append(Delete(n))  // Append new Delete
```

### Retain Merging Algorithm

Retain merging is also simple:

```pseudocode
FUNCTION Retain(n):
    IF last operation is Retain:
        lastRetain.n += n  // Merge with last Retain
    ELSE:
        append(Retain(n))  // Append new Retain
```

### Why Merge Operations?

1. **Memory efficiency**: Smaller operation history
2. **Network efficiency**: Smaller JSON payloads over WebSocket
3. **Processing speed**: Fewer operations to iterate over
4. **Semantic clarity**: Typing "hello" is one conceptual action

**Example impact**:
```pseudocode
// Without merging: 1000-character paste
operations = [Insert("a"), Insert("b"), ..., Insert("z"), ...]  // 1000 ops
JSON size ≈ 1000 * 10 bytes = 10KB

// With merging:
operations = [Insert("ab...z")]  // 1 op
JSON size ≈ 1000 bytes = 1KB

// 10x size reduction!
```

### Implementation

The merging logic is built into the Insert, Delete, and Retain methods of OperationSeq:

```go
// From operation.go
func (o *OperationSeq) Insert(s string) {
    // ... automatic merging logic ...
}

func (o *OperationSeq) Delete(n uint64) {
    // ... automatic merging logic ...
}

func (o *OperationSeq) Retain(n uint64) {
    // ... automatic merging logic ...
}
```

Users don't need to think about merging - it happens automatically.

---

## Unicode Codepoint Handling

A critical design decision in this OT implementation is **counting Unicode codepoints, not bytes**.

### The Challenge

```pseudocode
text = "hello 😀"

// Different counting methods:
byte_count = 10         // Emoji is 4 bytes in UTF-8
codepoint_count = 7     // Emoji is 1 codepoint
grapheme_count = 7      // Emoji is 1 grapheme cluster

// Which one to use?
```

### The Decision: Codepoints

```pseudocode
// OT uses codepoint counting
text = "hello 😀"
operation.baseLen = 7     // 6 ASCII chars + 1 emoji codepoint
operation.targetLen = 7
```

### Why Codepoints?

**1. JavaScript compatibility**
- JavaScript (and thus Monaco editor) uses UTF-16 code units
- For most text, UTF-16 code units ≈ Unicode codepoints
- Ensures consistent behavior with browser-based editors

**2. Rust compatibility**
- The Rust `operational-transform` crate counts codepoints
- Direct port maintains identical behavior

**3. User perception**
- Users perceive emoji as "one character"
- Codepoint counting matches user intuition for most cases

**4. Editor compatibility**
- Monaco editor (used in VS Code) uses UTF-16 offsets
- Most editors conceptually work with codepoints

### Implementation

Go's `utf8` package makes this straightforward:

```go
// From operation.go
func charCount(s string) int {
    return utf8.RuneCountInString(s)
}
```

This function is used everywhere character counts are needed:

```go
func (o *OperationSeq) Insert(s string) {
    o.targetLen += charCount(s)  // Count codepoints, not bytes
    // ...
}
```

### Example

```pseudocode
// Inserting emoji
text = "hello"
operation = [Retain(5), Insert(" 😀")]

// Calculation:
baseLen = 5
targetLen = 5 + charCount(" 😀") = 5 + 2 = 7

// Apply to "hello":
result = "hello 😀"
length(result) = 7 codepoints ✓
```

### Edge Cases

**Multi-codepoint emoji** (emoji with modifiers):
```pseudocode
emoji = "👨‍👩‍👧‍👦"  // Family emoji
codepoint_count = 7     // Multiple codepoints
grapheme_count = 1      // Single grapheme cluster

// OT counts: 7 (codepoints)
// User sees: 1 character (grapheme)
// This is a known limitation
```

For typical use cases (standard emoji, regular text), codepoint counting works well. Edge cases with complex grapheme clusters are rare in collaborative editing scenarios.

---

## Creating Operations

Operations are built incrementally by calling methods on an OperationSeq instance.

### Basic Pattern

```pseudocode
// Create a new operation sequence
operation = NewOperationSeq()

// Build the operation by calling methods
operation.Retain(5)
operation.Insert("world")
operation.Delete(3)
operation.Retain(10)

// Operations are automatically merged
// operation.operations = [Retain(5), Insert("world"), Delete(3), Retain(10)]
```

### Example: Replacing Text

```pseudocode
// Original: "hello world"
// Goal: "hello universe"

operation = NewOperationSeq()
operation.Retain(6)          // Skip "hello "
operation.Delete(5)          // Delete "world"
operation.Insert("universe") // Insert "universe"

// operation = [Retain(6), Delete(5), Insert("universe")]
// baseLen = 11 (original length)
// targetLen = 14 (result length)
```

### Example: Insert at Beginning

```pseudocode
// Original: "world"
// Goal: "hello world"

operation = NewOperationSeq()
operation.Insert("hello ")   // Insert at start
operation.Retain(5)          // Keep "world"

// operation = [Insert("hello "), Retain(5)]
// baseLen = 5
// targetLen = 11
```

### Example: Append at End

```pseudocode
// Original: "hello"
// Goal: "hello world"

operation = NewOperationSeq()
operation.Retain(5)          // Keep "hello"
operation.Insert(" world")   // Append at end

// operation = [Retain(5), Insert(" world")]
// baseLen = 5
// targetLen = 11
```

### Implementation

```go
import "github.com/shiv248/operational-transformation-go"

// Create operation
op := ot.NewOperationSeq()

// Build incrementally
op.Retain(5)
op.Insert("world")
op.Delete(3)

// Check properties
fmt.Println(op.BaseLen())    // Required input length
fmt.Println(op.TargetLen())  // Output length
fmt.Println(op.IsNoop())     // Check if no-op
```

---

## Noop Detection

A **noop** (no-operation) is an operation that has no effect on the document. Detecting noops is important for optimization.

### What is a Noop?

```pseudocode
FUNCTION IsNoop(operation):
    // Empty operation is a noop
    IF operation is empty:
        RETURN true

    // Single Retain is a noop (just moving cursor)
    IF operation has exactly one component:
        IF that component is Retain:
            RETURN true  // Just moving cursor, no change

    // All other cases are not noops
    RETURN false
```

### Examples

```pseudocode
// Noop: Empty operation
operation = []
IsNoop(operation) = true

// Noop: Single Retain
operation = [Retain(10)]
IsNoop(operation) = true

// NOT a noop: Has Insert
operation = [Retain(5), Insert("x")]
IsNoop(operation) = false

// NOT a noop: Has Delete
operation = [Delete(3)]
IsNoop(operation) = false

// NOT a noop: Multiple Retains
// (This never happens due to merging, but would not be a noop)
operation = [Retain(5), Retain(5)]
IsNoop(operation) = false
```

### Why Detect Noops?

**1. Network optimization**
```pseudocode
BEFORE sending operation over WebSocket:
    IF operation.IsNoop():
        // Don't send - saves bandwidth
        RETURN
```

**2. History compression**
```pseudocode
WHEN adding operation to history:
    IF operation.IsNoop():
        // Don't add to history - saves memory
        RETURN
```

**3. Client-side efficiency**
```pseudocode
WHEN user makes no actual change:
    operation = compute_operation(old_text, new_text)
    IF operation.IsNoop():
        // No need to broadcast or update state
        RETURN
```

### Why is Single Retain a Noop?

A Retain just moves the cursor - it doesn't change the document:

```pseudocode
// Apply Retain(10) to "hello world"
operation = [Retain(10)]

result = "hello world"  // Unchanged!
// The cursor moved, but the text is identical
```

Compare to a Retain followed by other operations:
```pseudocode
// Apply [Retain(5), Insert("x")] to "hello world"
operation = [Retain(5), Insert("x")]

result = "hellox world"  // Changed! Not a noop
```

### Implementation

```go
func (o *OperationSeq) IsNoop() bool {
    if len(o.ops) == 0 {
        return true
    }
    if len(o.ops) == 1 {
        if _, ok := o.ops[0].(Retain); ok {
            return true
        }
    }
    return false
}
```

---

## Summary

This document covered the fundamentals of operations in the OT system:

**Core Operations**:
- **Retain(n)**: Move cursor n positions forward (no change to text)
- **Delete(n)**: Remove n characters at cursor position
- **Insert(text)**: Insert text at cursor position

**OperationSeq Structure**:
- Sequence of operations describing a transformation
- Tracks `baseLen` (input length) and `targetLen` (output length)
- Enables validation, composition, and transformation

**Key Optimizations**:
- **Automatic merging**: Consecutive operations of same type are merged
- **Noop detection**: Empty operations and single Retains are identified as no-ops
- **Unicode handling**: Codepoint counting (not bytes) for compatibility

**Design Decisions**:
- Direct port from Rust `operational-transform` crate
- Compatible with JavaScript `ot.js`
- Canonical operation ordering: Insert before Delete

### Next Steps

Now that you understand the basic operations, you can learn about:

- **[02-transform.md]** - How concurrent operations are transformed to ensure convergence
- **[03-compose-apply.md]** - How sequential operations are composed and applied to text
- **[04-serialization.md]** - How operations are serialized to JSON for network transmission

### Related Documentation

- **Architecture overview** - High-level OT architecture in collaborative editing system
- **Transform algorithm** - Deep dive into the transformation algorithm (see [02-transform.md])
- **JSON wire format** - Serialization details (see [04-serialization.md])

---

**Questions or Issues?**

If you find errors or have suggestions for improving this documentation, please file an issue in the repository.
