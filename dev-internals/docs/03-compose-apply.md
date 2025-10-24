# Compose and Apply Operations

This document explains how to compose sequential operations and apply operations to text. These are essential algorithms for operation history compression, undo functionality, and actually transforming document text.

**Prerequisites**: Read [01-operations.md] and [02-transform.md] first to understand the basic operations and the transform algorithm.

---

## Table of Contents

1. [What is Compose?](#what-is-compose)
2. [Why Compose?](#why-compose)
3. [Compose vs Transform](#compose-vs-transform)
4. [Compose Cases](#compose-cases)
5. [Compose Algorithm](#compose-algorithm)
6. [Apply Operation to Text](#apply-operation-to-text)
7. [Invert Operation (Undo)](#invert-operation-undo)
8. [Use Cases](#use-cases)

---

## What is Compose?

**The Compose Property**:

```pseudocode
Given:
    - Document state S
    - Sequential operations A then B
      (where A.targetLen == B.baseLen)

Compose produces C such that:
    apply(S, compose(A, B)) = apply(apply(S, A), B)

// Single operation C equivalent to A followed by B
```

**Visual Example**:

```
Original: "hello"

Operation A: Insert("X") at position 0
Operation B: Insert("Y") at position 1

Sequential application:
    apply("hello", A) = "Xhello"
    apply("Xhello", B) = "XYhello"

Composed:
    C = compose(A, B)
    apply("hello", C) = "XYhello"

// Same result, but C is a single operation
```

Compose merges two consecutive operations into one operation while preserving the changes of both. The composed operation produces the same result as applying both operations sequentially, but it's more efficient to store and transmit.

---

## Why Compose?

Compose is critical for several real-world use cases in collaborative editing systems.

### Use Case 1: Operation History Compression

```pseudocode
// User types "hello" character by character
history = [
    Insert("h"),
    Insert("e"),
    Insert("l"),
    Insert("l"),
    Insert("o")
]

// Compose into single operation
composed = compose_all(history)
// composed = Insert("hello")

// Benefits:
// - Smaller history (1 operation vs 5)
// - Faster to send over network
// - Less memory usage
```

Without composition, every keystroke would remain as a separate operation in the history. A document with thousands of edits would have an enormous operation history. Composition allows us to merge sequential operations, dramatically reducing storage and bandwidth requirements.

### Use Case 2: Client Pending Buffer

```pseudocode
CLIENT state:
    pending_operations = []  // Operations not yet acknowledged by server

ON user types "h":
    new_op = Insert("h")
    pending_operations.push(new_op)
    send_to_server(new_op)

ON user types "e" (before "h" acknowledged):
    new_op = Insert("e")
    // Compose with pending operations
    pending_operations = [compose(pending_operations[0], new_op)]
    // Keep pending buffer as single operation
```

In collaborative editing, clients may send operations faster than the server can acknowledge them. Composing pending operations keeps the pending buffer small and efficient.

### Use Case 3: Network Bandwidth Optimization

Instead of sending many small operations individually, clients can compose them into larger batches before transmission:

```pseudocode
// Without composition (bad)
send([Insert("h")])     // 20 bytes
send([Insert("e")])     // 20 bytes
send([Insert("l")])     // 20 bytes
// Total: 60 bytes, 3 network round trips

// With composition (good)
composed = compose_all([Insert("h"), Insert("e"), Insert("l")])
send([Insert("hel")])   // 25 bytes
// Total: 25 bytes, 1 network round trip
```

---

## Compose vs Transform

It's crucial to understand the difference between compose and transform, as they solve different problems.

**Key Differences**:

```pseudocode
Transform:
    - Concurrent operations (same baseLen)
    - A and B both based on same document state
    - Produces A' and B' that can be applied in either order
    - Used for conflict resolution

Compose:
    - Sequential operations (A.targetLen == B.baseLen)
    - B is based on state after applying A
    - Produces C = A followed by B
    - Used for operation compression
```

**Error Conditions**:

```pseudocode
Transform requires: A.baseLen == B.baseLen
    // Both operations start from same document state

Compose requires: A.targetLen == B.baseLen
    // B starts where A ends
```

**When to Use Which**:

- **Transform**: When two users edit concurrently and you need to resolve conflicts
- **Compose**: When you have sequential operations from the same user or want to compress history

**Example Showing the Difference**:

```
Document: "hello"

Scenario 1 - Concurrent (use Transform):
    User A: Insert("X") at position 0 → "Xhello"
    User B: Insert("Y") at position 0 → "Yhello"
    Both start from "hello" (same baseLen = 5)
    Need: transform(A, B) to get A' and B'

Scenario 2 - Sequential (use Compose):
    User types "X": Insert("X") at position 0 → "Xhello"
    Then types "Y": Insert("Y") at position 1 → "XYhello"
    Second operation starts from "Xhello" (baseLen = 6)
    Need: compose(A, B) to get single operation C
```

---

## Compose Cases

Compose handles six different case combinations. Each case has specific logic for how to merge the operations.

### Case 1: Delete from First Operation

```pseudocode
compose(Delete(n), any_operation):
    // Delete always goes through unchanged
    result.Delete(n)
    // Continue with rest of A
```

**Rationale**: Deletes from the first operation affect the original document. They must be preserved in the composed result because they happen first.

**Example**:

```
A = [Delete(3), Retain(2)]
B = [Insert("X"), Retain(2)]

compose(A, B):
    Step 1: Delete(3) from A → add Delete(3) to result
    Step 2: Process remaining operations
```

### Case 2: Insert from Second Operation

```pseudocode
compose(any_operation, Insert(s)):
    // Insert always goes through unchanged
    result.Insert(s)
    // Continue with rest of B
```

**Rationale**: Inserts from the second operation happen after the first operation completes. They must be preserved as they represent new changes made to the already-transformed document.

**Example**:

```
A = [Retain(5)]
B = [Retain(3), Insert("Y"), Retain(2)]

compose(A, B):
    Step 1: Process Retain operations
    Step 2: Insert("Y") from B → add Insert("Y") to result
```

### Case 3: Retain vs Retain

```pseudocode
compose(Retain(n1), Retain(n2)):
    min_n = min(n1, n2)
    result.Retain(min_n)

    // Process remainder
    IF n1 > n2:
        remainder_a = Retain(n1 - n2)
    ELSE IF n2 > n1:
        remainder_b = Retain(n2 - n1)
```

**Rationale**: When both operations retain, we're skipping over unchanged text in both the original and intermediate document. We process them piecewise.

**Example**:

```
A = [Retain(10)]
B = [Retain(3), Insert("X"), Retain(7)]

compose(A, B):
    Step 1: Retain(3) vs Retain(10)
        → result.Retain(3)
        → remainder_a = Retain(7)
    Step 2: Insert("X") from B
        → result.Insert("X")
    Step 3: Retain(7) vs Retain(7)
        → result.Retain(7)

Result: [Retain(3), Insert("X"), Retain(7)]
```

### Case 4: Insert vs Delete

```pseudocode
compose(Insert(s), Delete(n)):
    text_len = length(s)

    IF text_len < n:
        // Delete entire insert plus more
        // Insert is canceled out (not added to result)
        remainder_delete = Delete(n - text_len)
    ELSE IF text_len == n:
        // Insert completely canceled
        // Both operations removed
    ELSE:
        // Delete part of insert
        result.Insert(s[n:])  // Keep remaining inserted text
```

**Rationale**: If the first operation inserts text and the second deletes it (or part of it), we can optimize by not inserting it in the first place.

**Example 1 - Complete Cancellation**:

```
A = [Insert("hello")]  // Insert 5 characters
B = [Delete(5)]        // Delete 5 characters

compose(A, B):
    Insert("hello") vs Delete(5)
    → Both cancel out completely

Result: [] (empty operation)
```

**Example 2 - Partial Deletion**:

```
A = [Insert("hello world")]  // Insert 11 characters
B = [Delete(6), Retain(5)]   // Delete "hello " (6 chars including space)

compose(A, B):
    Insert("hello world") vs Delete(6)
    → Delete first 6 characters of inserted text
    → Keep "world"

Result: [Insert("world")]
```

**Example 3 - Delete More Than Insert**:

```
A = [Retain(5), Insert("XX"), Retain(3)]
B = [Retain(5), Delete(4), Retain(1)]  // Delete the XX plus 2 more

compose(A, B):
    Retain(5) vs Retain(5) → Retain(5)
    Insert("XX") vs Delete(4) → Insert canceled, Delete(2) remains
    Delete(2) vs Retain(1) → Delete(1), Retain(1)

Result: [Retain(5), Delete(1), Retain(1)]
```

### Case 5: Insert vs Retain

```pseudocode
compose(Insert(s), Retain(n)):
    text_len = length(s)

    IF text_len <= n:
        result.Insert(s)  // Insert retained in full
    ELSE:
        // Only retain part of the insert
        result.Insert(s[0:n])  // Partially retained
        remainder_a = Insert(s[n:])
```

**Rationale**: When the first operation inserts and the second retains, the inserted text survives in the composed operation.

**Example**:

```
A = [Insert("hello")]  // Insert 5 characters
B = [Retain(5)]        // Retain all 5

compose(A, B):
    Insert("hello") vs Retain(5)
    → Insert goes through

Result: [Insert("hello")]
```

### Case 6: Retain vs Delete

```pseudocode
compose(Retain(n), Delete(m)):
    min_n = min(n, m)
    result.Delete(min_n)  // Retained text is now deleted

    // Process remainder
    IF n > m:
        remainder_a = Retain(n - m)
    ELSE IF m > n:
        remainder_b = Delete(m - n)
```

**Rationale**: If the first operation retains text (keeps it unchanged) and the second operation deletes it, the composed result should delete it.

**Example**:

```
Original: "hello world"

A = [Retain(11)]           // Keep all text
B = [Delete(6), Retain(5)] // Delete "hello ", keep "world"

compose(A, B):
    Retain(11) vs Delete(6)
        → result.Delete(6)
        → remainder_a = Retain(5)
    Retain(5) vs Retain(5)
        → result.Retain(5)

Result: [Delete(6), Retain(5)]
```

---

## Compose Algorithm

Here's the complete composition algorithm:

```pseudocode
FUNCTION Compose(op_a, op_b):
    // Validate: A's output must match B's input
    IF op_a.targetLen != op_b.baseLen:
        ERROR "incompatible lengths"

    result = NewOperationSeq()

    iter_a = iterator(op_a.operations)
    iter_b = iterator(op_b.operations)

    component_a = iter_a.next()
    component_b = iter_b.next()

    WHILE component_a OR component_b:
        // Priority 1: Delete from A always goes first
        IF component_a is Delete(n):
            result.Delete(n)
            component_a = iter_a.next()
            CONTINUE

        // Priority 2: Insert from B always goes last
        IF component_b is Insert(s):
            result.Insert(s)
            component_b = iter_b.next()
            CONTINUE

        // Both components exhausted
        IF component_a is nil AND component_b is nil:
            BREAK

        // One is nil but other isn't - error
        IF component_a is nil OR component_b is nil:
            ERROR "incompatible lengths"

        // Process pairs (see cases above)
        CASE (component_a, component_b):
            (Retain(n1), Retain(n2)):
                // Case 3: Process piecewise

            (Insert(s), Delete(n)):
                // Case 4: Potential cancellation

            (Insert(s), Retain(n)):
                // Case 5: Insert passes through

            (Retain(n), Delete(m)):
                // Case 6: Retain becomes delete

    RETURN result
```

**Key Insights**:

1. **Priority ordering**: Deletes from A go first, Inserts from B go last. This ensures the composed operation maintains the correct semantics.

2. **Piecewise processing**: Similar to transform, we process operations piece by piece to handle cases where operation lengths don't match.

3. **Optimization opportunities**: Insert vs Delete can cancel out completely, reducing the size of the composed operation.

**Complete Example Walkthrough**:

```
Original: "hello"

A = [Retain(5), Insert(" world")]
B = [Retain(6), Insert("beautiful "), Retain(5)]

Compose A and B:

Initial state:
    component_a = Retain(5)
    component_b = Retain(6)

Step 1: Retain(5) vs Retain(6)
    → result.Retain(5)
    → component_a = Insert(" world")
    → component_b = Retain(1)  // remainder

Step 2: Insert(" world") (from A, not Delete, not at end)
         vs Retain(1) (from B)
    → This is Case 5: Insert vs Retain
    → Insert(" world") has length 6
    → Retain is only 1
    → result.Insert(" ")  // first 1 char of insert
    → component_a = Insert("world")  // remainder
    → component_b = next = Insert("beautiful ")

Step 3: Insert("world") vs Insert("beautiful ")
    → Insert from B has priority (goes last)
    → result.Insert("beautiful ")
    → component_b = next = Retain(5)

Step 4: Insert("world") vs Retain(5)
    → Case 5: Insert vs Retain
    → result.Insert("world")
    → component_a = next = nil
    → component_b = Retain(5)

Step 5: nil vs Retain(5)
    → This would be error, but actually Retain should consume
    from the text that was there

Result: [Retain(5), Insert(" "), Insert("beautiful "), Insert("world")]
       = [Retain(5), Insert(" beautiful world")]  // after merging

Apply to "hello":
    → Retain(5): "hello"
    → Insert(" beautiful world"): "hello beautiful world"
```

---

## Apply Operation to Text

Applying an operation transforms the input text to the output text. This is how we actually execute the changes described by an operation.

**Core Algorithm**:

```pseudocode
FUNCTION Apply(operation, text):
    // Validate
    IF operation.baseLen != length(text):
        ERROR "incompatible lengths"

    result = ""
    cursor = 0

    FOR EACH component IN operation.operations:
        CASE component:
            Retain(n):
                // Copy n characters from input
                result += text[cursor : cursor+n]
                cursor += n

            Delete(n):
                // Skip n characters from input
                cursor += n

            Insert(s):
                // Add inserted text
                result += s
                // cursor doesn't move (not consuming input)

    RETURN result
```

**Key Points**:

1. **Cursor tracking**: The cursor tracks our position in the input text
2. **Retain copies**: Retain operations copy characters from input to output
3. **Delete skips**: Delete operations skip characters (move cursor without copying)
4. **Insert doesn't move cursor**: Insert operations add text without consuming input

**Example Walkthrough**:

```
Input text: "hello world"
Operation: [Retain(6), Insert("beautiful "), Delete(1), Retain(4)]

Step 1: Retain(6)
    Copy 6 characters: "hello "
    result = "hello "
    cursor = 6

Step 2: Insert("beautiful ")
    Add inserted text: "beautiful "
    result = "hello beautiful "
    cursor = 6 (unchanged - Insert doesn't consume input)

Step 3: Delete(1)
    Skip 1 character (skip "w")
    result = "hello beautiful " (no change)
    cursor = 7

Step 4: Retain(4)
    Copy 4 characters: "orld"
    result = "hello beautiful orld"
    cursor = 11

Final result: "hello beautiful orld"
```

**Unicode Handling**:

Remember that all lengths are in Unicode codepoints, not bytes:

```
Input: "hello 😀"  // 7 codepoints (emoji is 1 codepoint)
Operation: [Retain(6), Delete(1)]  // Delete the emoji

Apply:
    Retain(6): Copy "hello " (6 codepoints)
    Delete(1): Skip "😀" (1 codepoint, but 4 bytes in UTF-8)

Result: "hello "
```

The implementation uses Go's rune type to handle Unicode correctly:

```go
func (o *OperationSeq) Apply(s string) (string, error) {
    // ... validation ...

    var result strings.Builder
    runes := []rune(s)  // Convert to runes (codepoints)
    idx := 0

    for _, op := range o.ops {
        switch v := op.(type) {
        case Retain:
            // Copy n runes
            for i := uint64(0); i < v.N && idx < len(runes); i++ {
                result.WriteRune(runes[idx])
                idx++
            }
        case Delete:
            // Skip n runes
            idx += int(v.N)
        case Insert:
            // Add inserted text
            result.WriteString(v.Text)
        }
    }

    return result.String(), nil
}
```

---

## Invert Operation (Undo)

Invert computes the inverse of an operation. The inverse reverts the effects of the operation, enabling undo functionality.

**Purpose**: Create an inverse operation for undo functionality

**Inverse Rules**:

```pseudocode
Retain(n)  → Retain(n)      // Retain is its own inverse
Delete(n)  → Insert(text)   // Insert back the deleted text
Insert(s)  → Delete(len(s)) // Delete the inserted text
```

**Algorithm**:

```pseudocode
FUNCTION Invert(operation, original_text):
    inverse = NewOperationSeq()
    cursor = 0

    FOR EACH component IN operation.operations:
        CASE component:
            Retain(n):
                // Retain stays Retain
                inverse.Retain(n)
                cursor += n

            Insert(s):
                // Insert becomes Delete
                inverse.Delete(length(s))
                // cursor doesn't move (Insert doesn't consume input)

            Delete(n):
                // Delete becomes Insert (with deleted text)
                deleted_text = original_text[cursor : cursor+n]
                inverse.Insert(deleted_text)
                cursor += n

    RETURN inverse
```

**Why We Need Original Text**:

To invert a Delete operation, we need to know what text was deleted so we can insert it back. That's why the Invert function requires the original text as a parameter.

**Example**:

```
Original text: "hello world"
Operation: [Retain(6), Insert("beautiful "), Delete(1), Retain(4)]

Result after applying operation: "hello beautiful orld"

Invert the operation:

Step 1: Retain(6)
    → inverse.Retain(6)
    cursor = 6

Step 2: Insert("beautiful ")
    → inverse.Delete(10)  // Delete 10 characters
    cursor = 6 (unchanged)

Step 3: Delete(1)
    → deleted_text = original[6:7] = "w"
    → inverse.Insert("w")
    cursor = 7

Step 4: Retain(4)
    → inverse.Retain(4)
    cursor = 11

Inverse operation: [Retain(6), Delete(10), Insert("w"), Retain(4)]

Apply inverse to "hello beautiful orld":
    Retain(6): "hello "
    Delete(10): "hello " (delete "beautiful ")
    Insert("w"): "hello w"
    Retain(4): "hello world"

Result: "hello world" ✓ (back to original)
```

**Implementation**:

```go
func (o *OperationSeq) Invert(s string) *OperationSeq {
    inverse := NewOperationSeq()
    runes := []rune(s)
    idx := 0

    for _, op := range o.ops {
        switch v := op.(type) {
        case Retain:
            inverse.Retain(v.N)
            idx += int(v.N)
        case Insert:
            inverse.Delete(uint64(charCount(v.Text)))
        case Delete:
            // Insert the deleted characters back
            deleted := string(runes[idx : idx+int(v.N)])
            inverse.Insert(deleted)
            idx += int(v.N)
        }
    }

    return inverse
}
```

**Undo/Redo Stack**:

```pseudocode
// Undo/Redo implementation using Invert

undo_stack = []
redo_stack = []
document = "initial text"

FUNCTION apply_operation(op):
    inverse = Invert(op, document)
    undo_stack.push(inverse)
    redo_stack.clear()  // Clear redo when new operation applied
    document = Apply(op, document)

FUNCTION undo():
    IF undo_stack is empty:
        RETURN

    inverse_op = undo_stack.pop()
    redo_op = Invert(inverse_op, document)
    redo_stack.push(redo_op)
    document = Apply(inverse_op, document)

FUNCTION redo():
    IF redo_stack is empty:
        RETURN

    redo_op = redo_stack.pop()
    undo_op = Invert(redo_op, document)
    undo_stack.push(undo_op)
    document = Apply(redo_op, document)
```

---

## Use Cases

Here's a summary of when to use each algorithm:

### Compose

**Use when**: You have sequential operations that you want to merge

**Examples**:
1. **Operation history compression**: Merge old operations to save space
   ```pseudocode
   // Compress history from last hour
   old_ops = [op1, op2, op3, ..., op100]
   compressed = compose_all(old_ops)  // Now just 1 operation
   ```

2. **Client pending buffer management**: Keep pending operations small
   ```pseudocode
   pending_op = compose(pending_op, new_local_op)
   ```

3. **Reducing network bandwidth**: Batch operations before sending
   ```pseudocode
   batch = []
   ON user_edit:
       batch.append(new_op)

   EVERY 100ms:
       IF batch not empty:
           send(compose_all(batch))
           batch.clear()
   ```

### Apply

**Use when**: You need to execute an operation on text

**Examples**:
1. **Updating document text**: Apply incoming operations
   ```pseudocode
   ON receive_operation(op):
       document_text = Apply(op, document_text)
       update_editor(document_text)
   ```

2. **Testing**: Verify operation correctness
   ```pseudocode
   TEST_CASE compose_property:
       result1 = Apply(Apply(text, A), B)
       result2 = Apply(text, Compose(A, B))
       ASSERT result1 == result2
   ```

3. **Initial state**: Creating document from operation history
   ```pseudocode
   FUNCTION restore_from_history(operations):
       text = ""
       FOR EACH op IN operations:
           text = Apply(op, text)
       RETURN text
   ```

### Invert

**Use when**: You need undo/redo functionality

**Examples**:
1. **Undo functionality**: Revert user changes
   ```pseudocode
   ON user_undo:
       inverse = Invert(last_op, current_text)
       current_text = Apply(inverse, current_text)
   ```

2. **Redo**: Undo the undo
   ```pseudocode
   ON user_redo:
       inverse_of_undo = Invert(last_undo_op, current_text)
       current_text = Apply(inverse_of_undo, current_text)
   ```

3. **Time-travel debugging**: Step through operation history
   ```pseudocode
   FUNCTION goto_revision(target_revision):
       WHILE current_revision > target_revision:
           inverse = Invert(history[current_revision], text)
           text = Apply(inverse, text)
           current_revision -= 1
   ```

---

## Summary

- **Compose**: Merges sequential operations (A.targetLen == B.baseLen) into one operation
- **Apply**: Executes an operation on text to produce the transformed text
- **Invert**: Creates the inverse operation for undo functionality

**Key Principles**:

1. Compose is for sequential operations; Transform is for concurrent operations
2. Apply is how you actually execute operations on text
3. Invert enables undo by creating the reverse operation
4. All three algorithms work together in a collaborative editing system

**Related Documentation**:

- For operation basics: [01-operations.md]
- For transform algorithm: [02-transform.md]
- For serialization: [04-serialization.md]

---

**Implementation Reference**:

- Compose: See `compose.go` in the codebase
- Apply: See `apply.go` in the codebase
- Tests: See `compose_test.go` and `operation_test.go` for examples

This is a direct port from the Rust operational-transform crate:
https://github.com/spebern/operational-transform-rs
