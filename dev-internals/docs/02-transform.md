# Transform Algorithm

> **Purpose**: This document provides a deep dive into the transform algorithm, the heart of Operational Transformation that resolves concurrent operations and ensures all clients converge to the same document state.

---

## Table of Contents

1. [What is Transform?](#what-is-transform)
2. [Why Transform is Needed](#why-transform-is-needed)
3. [The Transform Property](#the-transform-property)
4. [Transform Cases (Comprehensive)](#transform-cases-comprehensive)
5. [The Complete Transform Algorithm](#the-complete-transform-algorithm)
6. [Piecewise Processing](#piecewise-processing)
7. [Iterator Pattern](#iterator-pattern)
8. [Transform Guarantees](#transform-guarantees)
9. [Common Scenarios](#common-scenarios)
10. [Summary](#summary)

---

## What is Transform?

Transform is the core algorithm that enables concurrent editing in Operational Transformation. When two users edit the same document simultaneously, their operations are based on the same document state but may conflict. Transform resolves these conflicts deterministically.

**The fundamental problem**:
```pseudocode
Initial state: "hello"

User A creates: Insert("A") at position 0 → "Ahello"
User B creates: Insert("B") at position 0 → "Bhello"

Without transformation:
  - User A sees: "Ahello" then receives B's operation → "BAhello" ❌
  - User B sees: "Bhello" then receives A's operation → "ABhello" ❌

DIVERGENCE! Different users see different documents.
```

**Transform solves this**:
```pseudocode
Transform produces A' and B' such that:
  - User A applies: A → B' → "ABhello" ✓
  - User B applies: B → A' → "ABhello" ✓

CONVERGENCE! All users see the same document.
```

---

## Why Transform is Needed

### The Collaborative Editing Challenge

In a collaborative editing system, multiple users work on the same document:

```pseudocode
Time 0: Document = "hello", revision = 0

Time 1:
    User A (offline): Insert("A") at position 0
    User B (offline): Insert("B") at position 0

    // Both operations based on revision 0!

Time 2:
    Server receives A first
    Server state: "Ahello", revision = 1

    Server receives B (based on revision 0, but server is at revision 1)
    // Problem: B cannot be applied directly to "Ahello"

    Server TRANSFORMS B against A to get B'
    B' = Insert("B") at position 1  // Shifted to account for A

    Server state: "ABhello", revision = 2

Time 3:
    Server broadcasts to all clients:
        - Operation A with revision 0→1
        - Operation B' with revision 1→2

    All clients converge to: "ABhello"
```

### Without Transform

If we simply applied operations in the order received:

```pseudocode
Client A's view:
    - Applies own operation A: "Ahello"
    - Receives B from server: "BAhello"  ❌

Client B's view:
    - Applies own operation B: "Bhello"
    - Receives A from server: "ABhello"  ❌

Different final states! The system has diverged.
```

### With Transform

```pseudocode
Server:
    - Receives A first, applies it: "Ahello"
    - Receives B, transforms it against A to get B'
    - Applies B': "ABhello"
    - Broadcasts: A (rev 0→1) and B' (rev 1→2)

Client A:
    - Has already applied A: "Ahello"
    - Receives B': applies it → "ABhello"  ✓

Client B:
    - Has already applied B: "Bhello"
    - Receives A, transforms B against A to get A'
    - Undoes B, applies A, applies B': "ABhello"  ✓

All clients converge!
```

---

## The Transform Property

Transform takes two concurrent operations (based on the same document state) and produces two transformed operations that can be applied in either order.

**Mathematical definition**:

```pseudocode
Given:
    - Document state S
    - Concurrent operations A and B where:
        A.baseLen == B.baseLen == length(S)

Transform(A, B) produces (A', B') such that:
    apply(apply(S, A), B') == apply(apply(S, B), A')

In other words:
    S --A--> SA --B'--> SAB'
    |                    ||
    B                    ||  (same result)
    |                    ||
    v                    vv
    SB --A'--> SBA'  ====
```

**Visual Example**:

```pseudocode
Original document: "hello"

Operation A: Insert("X") at position 0
Operation B: Insert("Y") at position 5 (end)

Without coordination:
    apply("hello", A) = "Xhello"
    apply("hello", B) = "helloY"
    // Different starting points!

With transform:
    A', B' = Transform(A, B)

    Path 1: apply(apply("hello", A), B')
          = apply("Xhello", Insert("Y") at position 6)
          = "XhelloY"

    Path 2: apply(apply("hello", B), A')
          = apply("helloY", Insert("X") at position 0)
          = "XhelloY"

    // Same result! Convergence achieved.
```

---

## Transform Cases (Comprehensive)

Transform processes two operation sequences component by component, handling all possible pairs of operations. There are **9 distinct cases** to handle.

### Case 1: Insert vs Insert

When both operations insert text at the same position, we need deterministic tie-breaking.

```pseudocode
Transform(Insert(s1), Insert(s2)):
    // Use lexicographic string comparison
    IF s1 < s2:
        A_prime = Insert(s1)
        B_prime = Retain(length(s1)) + Insert(s2)
        // A goes first, B shifts right by length of s1

    ELSE IF s1 == s2:
        // Both insert identical text - rare but possible
        A_prime = Insert(s1) + Retain(length(s1))
        B_prime = Insert(s2) + Retain(length(s2))
        // Both keep their inserts, each retains the other's

    ELSE:  // s1 > s2
        A_prime = Retain(length(s2)) + Insert(s1)
        B_prime = Insert(s2)
        // B goes first, A shifts right by length of s2
```

**Why string comparison?**
- **Deterministic**: All clients make the same decision without server coordination
- **Fair**: No user is always prioritized
- **Natural ordering**: Alphabetical ordering feels intuitive
- **Consistent**: Matches the Rust and JavaScript implementations

**Example**:
```pseudocode
Document: "hello"

A = [Retain(5), Insert("alpha")]
B = [Retain(5), Insert("beta")]

Transform(A, B):
    "alpha" < "beta" (lexicographically)

    A' = [Retain(5), Insert("alpha")]
    B' = [Retain(5), Retain(5), Insert("beta")]
         // Extra Retain(5) to skip over "alpha"

Result:
    apply("hello", A) = "helloalpha"
    apply("helloalpha", B') = "helloalphabeta"

    apply("hello", B) = "hellobeta"
    apply("hellobeta", A') = "helloalphabeta"

    ✓ Convergence!
```

### Case 2: Insert vs Retain

Insert always "wins" - it happens regardless of what the other operation is retaining.

```pseudocode
Transform(Insert(s), Retain(n)):
    A_prime = Insert(s)
    B_prime = Retain(length(s)) + Retain(n)
    // B must now retain past the newly inserted text
```

**Example**:
```pseudocode
Document: "hello"

A = [Retain(2), Insert("XX")]  // Insert "XX" after "he"
B = [Retain(5)]                // Just retain entire document

Transform(A, B):
    A' = [Retain(2), Insert("XX")]
    B' = [Retain(2), Retain(2), Retain(3)]
         // = [Retain(7)]  (after merging)

Result:
    apply("hello", A) = "heXXllo"
    apply("heXXllo", B') = "heXXllo"  (just retains)

    apply("hello", B) = "hello"  (no change)
    apply("hello", A') = "heXXllo"

    ✓ Both paths reach "heXXllo"
```

### Case 3: Insert vs Delete

Insert happens at the current position. Delete shifts to account for the insert.

```pseudocode
Transform(Insert(s), Delete(n)):
    A_prime = Insert(s)
    B_prime = Retain(length(s)) + Delete(n)
    // B must retain past the inserted text before deleting
```

**Example**:
```pseudocode
Document: "hello"

A = [Insert("XX")]       // Insert "XX" at start
B = [Delete(2), Retain(3)]  // Delete "he", keep "llo"

Transform(A, B):
    A' = [Insert("XX")]
    B' = [Retain(2), Delete(2), Retain(3)]

Result:
    apply("hello", A) = "XXhello"
    apply("XXhello", B') = "XXllo"

    apply("hello", B) = "llo"
    apply("llo", A') = "XXllo"

    ✓ Convergence to "XXllo"
```

### Case 4: Retain vs Retain

Both operations skip the same text. Process the minimum length, keep remainder.

```pseudocode
Transform(Retain(n1), Retain(n2)):
    min_n = min(n1, n2)

    A_prime = Retain(min_n)
    B_prime = Retain(min_n)

    // Handle remainder:
    IF n1 > n2:
        A_remainder = Retain(n1 - n2)
        // Continue processing A_remainder vs next B component

    ELSE IF n2 > n1:
        B_remainder = Retain(n2 - n1)
        // Continue processing next A component vs B_remainder

    // IF n1 == n2, both exhausted, move to next components
```

**Example**:
```pseudocode
A = [Retain(10)]
B = [Retain(3), Insert("X"), Retain(7)]

Step 1: Retain(10) vs Retain(3)
    min = 3
    A' += Retain(3)
    B' += Retain(3)
    A_current = Retain(7)  // Remainder

Step 2: Retain(7) vs Insert("X")
    A' += Retain(1)  // For the insert
    B' += Insert("X")
    A_current = Retain(7)  // Unchanged

Step 3: Retain(7) vs Retain(7)
    A' += Retain(7)
    B' += Retain(7)

Final:
    A' = [Retain(3), Retain(1), Retain(7)] = [Retain(11)]
    B' = [Retain(3), Insert("X"), Retain(7)]
```

### Case 5: Retain vs Delete

The retained text is being deleted by the other operation. Delete takes precedence.

```pseudocode
Transform(Retain(n), Delete(m)):
    min_n = min(n, m)

    A_prime = Delete(min_n)
    // A was retaining, but that text is now deleted

    B_prime = []
    // B's delete is absorbed (A was just retaining that text)

    // Process remainder similar to Retain vs Retain
```

**Example**:
```pseudocode
Document: "hello world"

A = [Retain(11)]           // Keep everything
B = [Delete(6), Retain(5)] // Delete "hello ", keep "world"

Transform(A, B):
    Step 1: Retain(11) vs Delete(6)
        A' += Delete(6)
        B' += []
        A_current = Retain(5)

    Step 2: Retain(5) vs Retain(5)
        A' += Retain(5)
        B' += Retain(5)

    A' = [Delete(6), Retain(5)]
    B' = [Retain(5)]

Result:
    apply("hello world", A) = "hello world"
    apply("hello world", A') = "world"

    apply("hello world", B) = "world"
    apply("world", B') = "world"

    ✓ Both reach "world"
```

### Case 6: Delete vs Delete

Both operations delete the same text. Only delete it once in the result.

```pseudocode
Transform(Delete(n1), Delete(n2)):
    min_n = min(n1, n2)

    // Both deleting same text - don't add to result
    A_prime = []
    B_prime = []

    // Process remainder:
    IF n1 > n2:
        A_remainder = Delete(n1 - n2)
        // A deletes more text

    ELSE IF n2 > n1:
        B_remainder = Delete(n2 - n1)
        // B deletes more text
```

**Example**:
```pseudocode
Document: "hello world"

A = [Delete(7), Retain(4)]  // Delete "hello w"
B = [Delete(6), Retain(5)]  // Delete "hello "

Transform(A, B):
    Step 1: Delete(7) vs Delete(6)
        min = 6
        A' += []
        B' += []
        A_current = Delete(1)  // Still need to delete 1 more

    Step 2: Delete(1) vs Retain(5)
        A' += Delete(1)
        B' += []
        A_current = Retain(4)
        B_current = Retain(4)

    Step 3: Retain(4) vs Retain(4)
        A' += Retain(4)
        B' += Retain(4)

    A' = [Delete(1), Retain(4)]
    B' = [Retain(4)]

Result:
    apply("hello world", A) = "orld"
    apply("orld", B') = "orld"

    apply("hello world", B) = "world"
    apply("world", A') = "orld"

    ✓ Convergence!
```

### Case 7: Delete vs Retain

Delete takes precedence. The text being retained is actually deleted.

```pseudocode
Transform(Delete(n), Retain(m)):
    min_n = min(n, m)

    A_prime = Delete(min_n)
    // A's delete goes through

    B_prime = []
    // B was retaining, but that text is deleted

    // Process remainder
```

This is symmetric to Case 5 (Retain vs Delete).

### Case 8: Delete vs Insert

These operations don't interact - they affect different positions.

```pseudocode
Transform(Delete(n), Insert(s)):
    // This case is handled by the general Insert rules
    // Insert from second operation always appears in B'
    A_prime = Retain(length(s))
    B_prime = Insert(s)
```

### Case 9: Retain vs Insert

Insert happens, Retain must account for new text.

```pseudocode
Transform(Retain(n), Insert(s)):
    A_prime = Retain(length(s))
    B_prime = Insert(s)
```

This is symmetric to Case 2 (Insert vs Retain).

---

## The Complete Transform Algorithm

Here's the full algorithm that handles all cases:

```pseudocode
FUNCTION Transform(operation_a, operation_b):
    // Validation: both operations must be concurrent
    // (based on same document state)
    IF operation_a.baseLen != operation_b.baseLen:
        ERROR "incompatible lengths - operations not concurrent"

    // Initialize result operations
    a_prime = NewOperationSeq()
    b_prime = NewOperationSeq()

    // Create iterators for piecewise processing
    iter_a = iterator(operation_a.operations)
    iter_b = iterator(operation_b.operations)

    component_a = iter_a.next()
    component_b = iter_b.next()

    WHILE component_a OR component_b:
        // Both operations exhausted - done
        IF component_a == null AND component_b == null:
            BREAK

        // Priority 1: Handle Insert vs Insert (tie-breaking needed)
        IF component_a is Insert AND component_b is Insert:
            // ... handle with string comparison (Case 1)
            CONTINUE

        // Priority 2: Handle Insert from first operation
        IF component_a is Insert:
            a_prime.Insert(component_a.text)
            b_prime.Retain(length(component_a.text))
            component_a = iter_a.next()
            CONTINUE

        // Priority 3: Handle Insert from second operation
        IF component_b is Insert:
            a_prime.Retain(length(component_b.text))
            b_prime.Insert(component_b.text)
            component_b = iter_b.next()
            CONTINUE

        // At this point, no Inserts remain
        // One operation being null here is an error
        IF component_a == null OR component_b == null:
            ERROR "incompatible lengths"

        // Handle Retain vs Retain
        IF component_a is Retain AND component_b is Retain:
            min_n = min(component_a.N, component_b.N)
            a_prime.Retain(min_n)
            b_prime.Retain(min_n)

            IF component_a.N < component_b.N:
                component_b = Retain(component_b.N - component_a.N)
                component_a = iter_a.next()
            ELSE IF component_a.N == component_b.N:
                component_a = iter_a.next()
                component_b = iter_b.next()
            ELSE:
                component_a = Retain(component_a.N - component_b.N)
                component_b = iter_b.next()

            CONTINUE

        // Handle Delete vs Delete
        IF component_a is Delete AND component_b is Delete:
            min_n = min(component_a.N, component_b.N)
            // Both delete same text - don't add to result

            IF component_a.N < component_b.N:
                component_b = Delete(component_b.N - component_a.N)
                component_a = iter_a.next()
            ELSE IF component_a.N == component_b.N:
                component_a = iter_a.next()
                component_b = iter_b.next()
            ELSE:
                component_a = Delete(component_a.N - component_b.N)
                component_b = iter_b.next()

            CONTINUE

        // Handle Delete vs Retain
        IF component_a is Delete AND component_b is Retain:
            min_n = min(component_a.N, component_b.N)
            a_prime.Delete(min_n)
            // B's retain is absorbed

            IF component_a.N < component_b.N:
                component_b = Retain(component_b.N - component_a.N)
                component_a = iter_a.next()
            ELSE IF component_a.N == component_b.N:
                component_a = iter_a.next()
                component_b = iter_b.next()
            ELSE:
                component_a = Delete(component_a.N - component_b.N)
                component_b = iter_b.next()

            CONTINUE

        // Handle Retain vs Delete
        IF component_a is Retain AND component_b is Delete:
            min_n = min(component_a.N, component_b.N)
            b_prime.Delete(min_n)
            // A's retain is absorbed

            IF component_a.N < component_b.N:
                component_b = Delete(component_b.N - component_a.N)
                component_a = iter_a.next()
            ELSE IF component_a.N == component_b.N:
                component_a = iter_a.next()
                component_b = iter_b.next()
            ELSE:
                component_a = Retain(component_a.N - component_b.N)
                component_b = iter_b.next()

            CONTINUE

        // Should never reach here if operations are valid
        ERROR "invalid operation types"

    RETURN (a_prime, b_prime)
```

---

## Piecewise Processing

Transform processes operations **component by component** rather than all at once. This is necessary because operations may have different granularities.

### Why Piecewise?

Consider this example:

```pseudocode
A = [Retain(10)]
B = [Retain(3), Insert("X"), Retain(7)]

// A has 1 component, B has 3 components
// Can't process in single step!
```

We need to break down A's `Retain(10)` to match B's components:

```pseudocode
Step 1: Process Retain(3) from A's Retain(10) vs B's Retain(3)
    → A' gets Retain(3)
    → B' gets Retain(3)
    → A's remainder: Retain(7)
    → B advances to Insert("X")

Step 2: Process Retain(7) vs Insert("X")
    → A' gets Retain(1)
    → B' gets Insert("X")
    → A's remainder: Retain(7) (unchanged)
    → B advances to Retain(7)

Step 3: Process Retain(7) vs Retain(7)
    → A' gets Retain(7)
    → B' gets Retain(7)
    → Both exhausted

Final:
    A' = [Retain(3), Retain(1), Retain(7)] = [Retain(11)]
    B' = [Retain(3), Insert("X"), Retain(7)]
```

### The Remainder Pattern

When processing Retain vs Retain or Delete vs Delete:

```pseudocode
IF component_a.N < component_b.N:
    // Process all of component_a
    process(component_a.N)

    // Create remainder for component_b
    component_b = SameType(component_b.N - component_a.N)

    // Advance component_a to next
    component_a = iter_a.next()

ELSE IF component_a.N == component_b.N:
    // Process both completely
    process(component_a.N)

    // Advance both
    component_a = iter_a.next()
    component_b = iter_b.next()

ELSE:  // component_a.N > component_b.N
    // Process all of component_b
    process(component_b.N)

    // Create remainder for component_a
    component_a = SameType(component_a.N - component_b.N)

    // Advance component_b to next
    component_b = iter_b.next()
```

This pattern ensures we process exactly the right amount from each operation.

---

## Iterator Pattern

The transform algorithm uses iterators to traverse operation components. This allows for the piecewise processing described above.

### Iterator Structure

```pseudocode
TYPE OpIterator:
    operations: List[Operation]
    current_index: int

FUNCTION new_iterator(operations):
    RETURN OpIterator{
        operations: operations,
        current_index: 0
    }

METHOD next():
    IF current_index >= length(operations):
        RETURN null

    op = operations[current_index]
    current_index += 1
    RETURN op
```

### Implementation

The Go implementation uses a simple iterator:

```go
type opIterator struct {
    ops []Operation
    idx int
}

func newOpIterator(ops []Operation) *opIterator {
    return &opIterator{ops: ops, idx: 0}
}

func (it *opIterator) next() Operation {
    if it.idx >= len(it.ops) {
        return nil
    }
    op := it.ops[it.idx]
    it.idx++
    return op
}
```

### Why Iterators?

**Advantages**:
1. **Clean abstraction**: Separates iteration logic from transformation logic
2. **Null handling**: Easy to detect when operations are exhausted
3. **No index arithmetic**: Simpler than manual index management
4. **Piecewise processing**: Can "rewrite" current component with remainder

**The key insight**: We don't always advance both iterators. Sometimes we:
- Advance only A's iterator
- Advance only B's iterator
- Advance both iterators
- Rewrite current component without advancing (when processing remainder)

---

## Transform Guarantees

Transform provides several critical guarantees that make OT work:

### 1. Convergence Property (TP1)

**Guarantee**: Both application paths reach the same final state.

```pseudocode
Given (A', B') = Transform(A, B):
    apply(apply(S, A), B') == apply(apply(S, B), A')

// No matter which operation is applied first,
// the final result is identical
```

**Why this matters**: Ensures all clients see the same document even when operations arrive in different orders.

### 2. Commutativity (After Transform)

After transformation, operations commute:

```pseudocode
A' and B' can be applied in either order to produce same result
```

This is actually a consequence of TP1.

### 3. Determinism

**Guarantee**: Transform always produces the same result for the same input.

```pseudocode
Transform(A, B) always produces the same (A', B')

// No randomness, no external state
// Pure function of inputs
```

**Why this matters**: All clients can independently compute the same transformation without server coordination.

### 4. Length Consistency

**Guarantee**: Transformed operations maintain correct length relationships.

```pseudocode
Given (A', B') = Transform(A, B):
    A'.baseLen == A.targetLen
    A'.targetLen == compose(A, B').targetLen

    B'.baseLen == B.targetLen
    B'.targetLen == compose(B, A').targetLen
```

### 5. No Split-Brain

In a client-server architecture:

```pseudocode
SERVER is the single source of truth
    - Receives operations from clients
    - Applies them in order
    - Transforms later operations against earlier ones
    - Broadcasts the canonical history

CLIENTS synchronize with server
    - Send operations based on their current state
    - Receive server's canonical operations
    - Transform their pending operations
```

This prevents the "split-brain" problem where different parts of the system have different views of truth.

---

## Common Scenarios

### Scenario 1: Simultaneous Typing

```pseudocode
Document: "hello"

User A: Types "A" at position 0
User B: Types "B" at position 0

A = [Insert("A"), Retain(5)]
B = [Insert("B"), Retain(5)]

Transform(A, B):
    "A" < "B" lexicographically

    A' = [Insert("A"), Retain(5)]
    B' = [Retain(1), Insert("B"), Retain(5)]

Result:
    All clients converge to: "ABhello"
```

### Scenario 2: Concurrent Delete and Insert

```pseudocode
Document: "hello world"

User A: Deletes "world"
User B: Inserts "!!!" at end

A = [Retain(6), Delete(5)]
B = [Retain(11), Insert("!!!")]

Transform(A, B):
    A' = [Retain(6), Delete(5)]
    B' = [Retain(6), Insert("!!!")]

Result:
    Path A then B': "hello world" → "hello " → "hello !!!"
    Path B then A': "hello world" → "hello world!!!" → "hello !!!"

    ✓ Both converge to "hello !!!"
```

### Scenario 3: Overlapping Deletes

```pseudocode
Document: "hello world"

User A: Deletes characters 3-8 ("lo wo")
User B: Deletes characters 5-10 (" worl")

A = [Retain(3), Delete(5), Retain(3)]
B = [Retain(5), Delete(5), Retain(1)]

Transform(A, B):
    // Complex case with overlapping deletions
    // Transform ensures text is only deleted once

    A' = [Retain(3), Delete(3), Retain(1)]
    B' = [Retain(3), Retain(1)]

Result:
    Path A then B': "hello world" → "held" → "held"
    Path B then A': "hello world" → "hellod" → "held"

    ✓ Convergence!
```

---

## Summary

The Transform algorithm is the heart of Operational Transformation:

**Key Concepts**:
1. **Purpose**: Resolve concurrent operations to ensure convergence
2. **Input**: Two operations with the same baseLen (concurrent on same state)
3. **Output**: Two transformed operations (A', B') that can be applied in either order
4. **Method**: Process operations piecewise, component by component

**Nine Transform Cases**:
1. Insert vs Insert → Use string comparison for tie-breaking
2. Insert vs Retain → Insert wins, Retain shifts
3. Insert vs Delete → Insert wins, Delete shifts
4. Retain vs Retain → Process minimum, keep remainder
5. Retain vs Delete → Delete wins
6. Delete vs Delete → Delete text once, track remainder
7. Delete vs Retain → Delete wins
8. Delete vs Insert → Symmetric to case 3
9. Retain vs Insert → Symmetric to case 2

**Guarantees**:
- **Convergence**: Both paths reach same final state
- **Determinism**: Same inputs always produce same outputs
- **Length consistency**: Transformed operations maintain correct lengths
- **No split-brain**: Server maintains canonical order

**Implementation Pattern**:
- Use iterators for traversal
- Process piecewise (handle components of different sizes)
- Track remainders when components have different lengths
- Maintain baseLen and targetLen invariants

**Cross-References**:
- For operation basics: [01-operations.md](01-operations.md)
- For sequential composition: [03-compose-apply.md](03-compose-apply.md)
- For serialization format: [04-serialization.md](04-serialization.md)

**Further Reading**:
- Sun, C., & Ellis, C. (1998). [Operational transformation in real-time group editors](https://dl.acm.org/doi/10.1145/289444.289469). _CSCW '98_, 59-68.
  - The seminal paper establishing the theoretical foundations of OT

---

Transform enables the magic of real-time collaboration: multiple users editing simultaneously while maintaining a single, consistent document state. Understanding transform is essential for implementing and debugging OT-based collaborative editing systems.
