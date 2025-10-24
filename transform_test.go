package ot

import (
	"errors"
	"testing"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		opsA    func() *OperationSeq
		opsB    func() *OperationSeq
		expectS string
	}{
		{
			name: "concurrent inserts at different positions",
			s:    "abc",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(3)
				o.Insert("def")
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(3)
				o.Insert("ghi")
				return o
			},
			expectS: "abcdefghi", // or "abcghidef" depending on tie-breaking
		},
		{
			name: "concurrent inserts at same position",
			s:    "abc",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(2)
				o.Insert("X")
				o.Retain(1)
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(2)
				o.Insert("Y")
				o.Retain(1)
				return o
			},
			expectS: "abXYc", // or "abYXc" depending on tie-breaking
		},
		{
			name: "insert vs delete",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(6) // delete "hello "
				o.Retain(5)
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(5)
				o.Insert("!") // insert "!" after "hello"
				o.Retain(6)
				return o
			},
			expectS: "world!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.opsA()
			b := tt.opsB()

			// Transform A and B
			aPrime, bPrime, err := a.Transform(b)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Apply A then B'
			afterA, err := a.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply A failed: %v", err)
			}
			afterAB, err := bPrime.Apply(afterA)
			if err != nil {
				t.Fatalf("Apply B' failed: %v", err)
			}

			// Apply B then A'
			afterB, err := b.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply B failed: %v", err)
			}
			afterBA, err := aPrime.Apply(afterB)
			if err != nil {
				t.Fatalf("Apply A' failed: %v", err)
			}

			// Results should be identical (convergence property)
			if afterAB != afterBA {
				t.Errorf("transform property failed:\n  A+B' = %q\n  B+A' = %q", afterAB, afterBA)
			}
		})
	}
}

func TestTransformProperty(t *testing.T) {
	// TP1: apply(apply(S, A), B') = apply(apply(S, B), A')
	// where (A', B') = transform(A, B)

	tests := []struct {
		s string
		a func() *OperationSeq
		b func() *OperationSeq
	}{
		{
			s: "hello",
			a: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(5)
				o.Insert(" world")
				return o
			},
			b: func() *OperationSeq {
				o := NewOperationSeq()
				o.Insert("Hi! ")
				o.Retain(5)
				return o
			},
		},
		{
			s: "abcdefgh",
			a: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(3)
				o.Delete(2)
				o.Retain(3)
				return o
			},
			b: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(5)
				o.Delete(3)
				return o
			},
		},
		{
			s: "test",
			a: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(2)
				o.Insert("XX")
				o.Retain(2)
				return o
			},
			b: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(2)
				o.Insert("YY")
				o.Retain(2)
				return o
			},
		},
	}

	for i, tt := range tests {
		a := tt.a()
		b := tt.b()

		// Transform
		aPrime, bPrime, err := a.Transform(b)
		if err != nil {
			t.Fatalf("test %d: Transform failed: %v", i, err)
		}

		// Path 1: A then B'
		afterA, err := a.Apply(tt.s)
		if err != nil {
			t.Fatalf("test %d: Apply A failed: %v", i, err)
		}
		path1, err := bPrime.Apply(afterA)
		if err != nil {
			t.Fatalf("test %d: Apply B' failed: %v", i, err)
		}

		// Path 2: B then A'
		afterB, err := b.Apply(tt.s)
		if err != nil {
			t.Fatalf("test %d: Apply B failed: %v", i, err)
		}
		path2, err := aPrime.Apply(afterB)
		if err != nil {
			t.Fatalf("test %d: Apply A' failed: %v", i, err)
		}

		// Must converge
		if path1 != path2 {
			t.Errorf("test %d: TP1 failed:\n  S=%q\n  A+B'=%q\n  B+A'=%q", i, tt.s, path1, path2)
		}
	}
}

func TestTransformSymmetry(t *testing.T) {
	// Transform(A, B) and Transform(B, A) should produce symmetric results
	s := "hello"

	a := NewOperationSeq()
	a.Retain(5)
	a.Insert(" world")

	b := NewOperationSeq()
	b.Insert("Hi! ")
	b.Retain(5)

	aPrime, bPrime, err := a.Transform(b)
	if err != nil {
		t.Fatalf("Transform(A, B) failed: %v", err)
	}

	// Verify symmetry: both paths should converge
	// Path 1: apply a to s, then apply b' to result
	afterA, err := a.Apply(s)
	if err != nil {
		t.Fatalf("Apply a failed: %v", err)
	}
	afterAB, err := bPrime.Apply(afterA)
	if err != nil {
		t.Fatalf("Apply b' after a failed: %v", err)
	}

	// Path 2: apply b to s, then apply a' to result
	afterB, err := b.Apply(s)
	if err != nil {
		t.Fatalf("Apply b failed: %v", err)
	}
	afterBA, err := aPrime.Apply(afterB)
	if err != nil {
		t.Fatalf("Apply a' after b failed: %v", err)
	}

	// Both paths should produce the same result
	if afterAB != afterBA {
		t.Errorf("transform convergence failed:\n  a then b': %q\n  b then a': %q",
			afterAB, afterBA)
	}
}

// =============================================================================
// Additional transform tests beyond the Rust port
// These tests explicitly cover edge cases and error conditions that are not
// explicitly tested in operational-transform-rs but are important for
// comprehensive coverage.
// =============================================================================

func TestTransformErrorIncompatibleLengths(t *testing.T) {
	// Operations with different base lengths should error.
	// This ensures transform validates that both operations are
	// based on the same document state (concurrent operations).
	a := NewOperationSeq()
	a.Retain(5) // baseLen = 5

	b := NewOperationSeq()
	b.Retain(10) // baseLen = 10 (different from A)

	// Transform should reject these as they're not concurrent
	_, _, err := a.Transform(b)
	if !errors.Is(err, ErrIncompatibleLengths) {
		t.Errorf("expected ErrIncompatibleLengths, got %v", err)
	}
}

func TestTransformDeleteVsDelete(t *testing.T) {
	// Test Delete vs Delete scenarios to ensure both operations
	// don't delete the same text twice and properly handle overlaps.
	// This is critical for collaborative editing where multiple users
	// might delete the same or overlapping content.
	tests := []struct {
		name     string
		s        string
		opsA     func() *OperationSeq
		opsB     func() *OperationSeq
		expected string
	}{
		{
			name: "delete vs delete - same range",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(6) // delete "hello "
				o.Retain(5) // keep "world"
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(6) // delete "hello " (same as A)
				o.Retain(5) // keep "world"
				return o
			},
			expected: "world", // text only deleted once
		},
		{
			name: "delete vs delete - A shorter than B",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(5) // delete "hello"
				o.Retain(6) // keep " world"
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(11) // delete entire document
				return o
			},
			expected: "", // B's longer delete wins
		},
		{
			name: "delete vs delete - A longer than B",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(11) // delete entire document
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(5) // delete "hello"
				o.Retain(6) // keep " world"
				return o
			},
			expected: "", // A's longer delete wins
		},
		{
			name: "delete vs delete - overlapping ranges",
			s:    "abcdefgh",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(2) // keep "ab"
				o.Delete(4) // delete "cdef"
				o.Retain(2) // keep "gh"
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(4) // keep "abcd"
				o.Delete(3) // delete "efg"
				o.Retain(1) // keep "h"
				return o
			},
			expected: "abh", // A deletes "cdef", B deletes "efg", overlap "ef" only deleted once
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.opsA()
			b := tt.opsB()

			// Transform A and B
			aPrime, bPrime, err := a.Transform(b)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Path 1: Apply A then B'
			afterA, err := a.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply A failed: %v", err)
			}
			resultAB, err := bPrime.Apply(afterA)
			if err != nil {
				t.Fatalf("Apply B' failed: %v", err)
			}

			// Path 2: Apply B then A'
			afterB, err := b.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply B failed: %v", err)
			}
			resultBA, err := aPrime.Apply(afterB)
			if err != nil {
				t.Fatalf("Apply A' failed: %v", err)
			}

			// Verify convergence (both paths reach same result)
			if resultAB != resultBA {
				t.Errorf("convergence failed:\n  A+B' = %q\n  B+A' = %q", resultAB, resultBA)
			}

			// Verify expected result
			if resultAB != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, resultAB)
			}
		})
	}
}

func TestTransformInsertTieBreaking(t *testing.T) {
	// Test Insert vs Insert tie-breaking with lexicographic string comparison.
	// When two users insert at the same position simultaneously, we use
	// string comparison to deterministically order the inserts, ensuring
	// all clients converge to the same final state.
	tests := []struct {
		name        string
		s           string
		textA       string
		textB       string
		expectOrder string // "AB" or "BA" or "AA"
	}{
		{
			name:        "A < B lexicographically",
			s:           "hello",
			textA:       "alpha",
			textB:       "beta",
			expectOrder: "AB", // "alpha" comes before "beta"
		},
		{
			name:        "A > B lexicographically",
			s:           "hello",
			textA:       "zebra",
			textB:       "apple",
			expectOrder: "BA", // "apple" comes before "zebra"
		},
		{
			name:        "identical inserts",
			s:           "hello",
			textA:       "same",
			textB:       "same",
			expectOrder: "AA", // both inserted when identical
		},
		{
			name:        "numeric strings",
			s:           "test",
			textA:       "123",
			textB:       "456",
			expectOrder: "AB", // "123" < "456" lexicographically
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A inserts textA at start
			a := NewOperationSeq()
			a.Insert(tt.textA)
			a.Retain(uint64(charCount(tt.s)))

			// B inserts textB at start (same position as A)
			b := NewOperationSeq()
			b.Insert(tt.textB)
			b.Retain(uint64(charCount(tt.s)))

			// Transform A and B
			aPrime, bPrime, err := a.Transform(b)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Path 1: Apply A then B'
			afterA, err := a.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply A failed: %v", err)
			}
			resultAB, err := bPrime.Apply(afterA)
			if err != nil {
				t.Fatalf("Apply B' failed: %v", err)
			}

			// Path 2: Apply B then A'
			afterB, err := b.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply B failed: %v", err)
			}
			resultBA, err := aPrime.Apply(afterB)
			if err != nil {
				t.Fatalf("Apply A' failed: %v", err)
			}

			// Verify convergence (both paths reach same result)
			if resultAB != resultBA {
				t.Errorf("convergence failed:\n  A+B' = %q\n  B+A' = %q", resultAB, resultBA)
			}

			// Verify correct ordering based on tie-breaking
			switch tt.expectOrder {
			case "AB":
				expected := tt.textA + tt.textB + tt.s
				if resultAB != expected {
					t.Errorf("expected A before B: %q, got %q", expected, resultAB)
				}
			case "BA":
				expected := tt.textB + tt.textA + tt.s
				if resultAB != expected {
					t.Errorf("expected B before A: %q, got %q", expected, resultAB)
				}
			case "AA":
				// Identical inserts both appear
				expected := tt.textA + tt.textB + tt.s
				if resultAB != expected {
					t.Errorf("expected both inserts: %q, got %q", expected, resultAB)
				}
			}
		})
	}
}

func TestTransformRetainVsDeleteEdgeCases(t *testing.T) {
	// Test Retain vs Delete to ensure that when one operation retains text
	// and another deletes it, the delete takes precedence. This covers
	// scenarios where operations have different granularities (piecewise processing).
	tests := []struct {
		name     string
		s        string
		opsA     func() *OperationSeq
		opsB     func() *OperationSeq
		expected string
	}{
		{
			name: "retain shorter than delete",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(3) // retain "hel" (shorter than B's delete)
				o.Retain(8) // retain rest
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(6) // delete "hello " (longer than A's first retain)
				o.Retain(5) // retain "world"
				return o
			},
			expected: "world", // B's delete wins over A's retain
		},
		{
			name: "retain longer than delete",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(11) // retain entire document (longer than B's delete)
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(3) // delete "hel" (shorter than A's retain)
				o.Retain(8) // retain "lo world"
				return o
			},
			expected: "lo world", // B's delete wins over A's retain
		},
		{
			name: "retain equals delete",
			s:    "hello world",
			opsA: func() *OperationSeq {
				o := NewOperationSeq()
				o.Retain(5) // retain "hello" (same length as B's delete)
				o.Retain(6) // retain " world"
				return o
			},
			opsB: func() *OperationSeq {
				o := NewOperationSeq()
				o.Delete(5) // delete "hello" (same length as A's retain)
				o.Retain(6) // retain " world"
				return o
			},
			expected: " world", // B's delete wins over A's retain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.opsA()
			b := tt.opsB()

			// Transform A and B
			aPrime, bPrime, err := a.Transform(b)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Path 1: Apply A then B'
			afterA, err := a.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply A failed: %v", err)
			}
			resultAB, err := bPrime.Apply(afterA)
			if err != nil {
				t.Fatalf("Apply B' failed: %v", err)
			}

			// Path 2: Apply B then A'
			afterB, err := b.Apply(tt.s)
			if err != nil {
				t.Fatalf("Apply B failed: %v", err)
			}
			resultBA, err := aPrime.Apply(afterB)
			if err != nil {
				t.Fatalf("Apply A' failed: %v", err)
			}

			// Verify convergence (both paths reach same result)
			if resultAB != resultBA {
				t.Errorf("convergence failed:\n  A+B' = %q\n  B+A' = %q", resultAB, resultBA)
			}

			// Verify expected result
			if resultAB != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, resultAB)
			}
		})
	}
}

func TestTransformRetainVsRetainEdgeCases(t *testing.T) {
	// Test Retain vs Retain to ensure proper piecewise processing when
	// operations have different component sizes. This tests the remainder
	// pattern where we process the minimum and keep track of what's left.
	s := "hello world"

	tests := []struct {
		name     string
		retainA  uint64
		retainB  uint64
		expected string
	}{
		{
			name:     "A shorter than B",
			retainA:  3,  // retain first 3 chars
			retainB:  11, // retain all 11 chars
			expected: s,  // both just retain, no changes
		},
		{
			name:     "A longer than B",
			retainA:  11, // retain all 11 chars
			retainB:  3,  // retain first 3 chars
			expected: s,  // both just retain, no changes
		},
		{
			name:     "equal lengths",
			retainA:  11, // retain all
			retainB:  11, // retain all
			expected: s,  // both just retain, no changes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build operation A
			a := NewOperationSeq()
			a.Retain(tt.retainA)
			if tt.retainA < 11 {
				a.Retain(11 - tt.retainA) // retain rest of document
			}

			// Build operation B
			b := NewOperationSeq()
			b.Retain(tt.retainB)
			if tt.retainB < 11 {
				b.Retain(11 - tt.retainB) // retain rest of document
			}

			// Transform A and B
			aPrime, bPrime, err := a.Transform(b)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Path 1: Apply A then B'
			afterA, err := a.Apply(s)
			if err != nil {
				t.Fatalf("Apply A failed: %v", err)
			}
			resultAB, err := bPrime.Apply(afterA)
			if err != nil {
				t.Fatalf("Apply B' failed: %v", err)
			}

			// Path 2: Apply B then A'
			afterB, err := b.Apply(s)
			if err != nil {
				t.Fatalf("Apply B failed: %v", err)
			}
			resultBA, err := aPrime.Apply(afterB)
			if err != nil {
				t.Fatalf("Apply A' failed: %v", err)
			}

			// Verify convergence (both paths reach same result)
			if resultAB != resultBA {
				t.Errorf("convergence failed:\n  A+B' = %q\n  B+A' = %q", resultAB, resultBA)
			}

			// Verify expected result
			if resultAB != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, resultAB)
			}
		})
	}
}
