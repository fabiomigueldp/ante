package storage

import (
	"bytes"
	"testing"
	"time"
)

func TestEncodeCanonicalIgnoresMapInsertionOrder(t *testing.T) {
	first := map[string]int{"a": 1, "b": 2, "c": 3}
	second := map[string]int{"c": 3, "a": 1, "b": 2}

	encodedFirst, err := EncodeCanonical(first)
	if err != nil {
		t.Fatalf("EncodeCanonical first error: %v", err)
	}
	encodedSecond, err := EncodeCanonical(second)
	if err != nil {
		t.Fatalf("EncodeCanonical second error: %v", err)
	}
	if !bytes.Equal(encodedFirst, encodedSecond) {
		t.Fatal("canonical encoding should ignore map insertion order")
	}
}

func TestEncodeCanonicalNormalizesTimeInstants(t *testing.T) {
	utc := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	other := utc.In(time.FixedZone("UTC+1", 60*60))

	encodedUTC, err := EncodeCanonical(utc)
	if err != nil {
		t.Fatalf("EncodeCanonical utc error: %v", err)
	}
	encodedOther, err := EncodeCanonical(other)
	if err != nil {
		t.Fatalf("EncodeCanonical other error: %v", err)
	}
	if !bytes.Equal(encodedUTC, encodedOther) {
		t.Fatal("same instant should encode identically")
	}
}

func TestCanonicalSHA256ChangesWithValue(t *testing.T) {
	a, err := CanonicalSHA256(struct {
		Name string
		Age  int
	}{Name: "A", Age: 10})
	if err != nil {
		t.Fatalf("CanonicalSHA256 first error: %v", err)
	}
	b, err := CanonicalSHA256(struct {
		Name string
		Age  int
	}{Name: "A", Age: 11})
	if err != nil {
		t.Fatalf("CanonicalSHA256 second error: %v", err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("different values should not hash identically")
	}
}

func TestEncodeCanonicalRejectsUnsupportedKinds(t *testing.T) {
	_, err := EncodeCanonical(1.25)
	if err == nil {
		t.Fatal("expected error for float canonical encoding")
	}
}
