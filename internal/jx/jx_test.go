package jx

import "testing"

func TestDeref_NonNil(t *testing.T) {
	v := 42
	if got := Deref(&v); got != 42 {
		t.Errorf("Deref(&42) = %d, want 42", got)
	}
}

func TestDeref_Nil(t *testing.T) {
	var p *int
	if got := Deref(p); got != 0 {
		t.Errorf("Deref(nil) = %d, want 0", got)
	}
}

func TestDeref_String(t *testing.T) {
	s := "hello"
	if got := Deref(&s); got != "hello" {
		t.Errorf("Deref = %q, want hello", got)
	}
}

func TestDeref_NilString(t *testing.T) {
	var p *string
	if got := Deref(p); got != "" {
		t.Errorf("Deref(nil) = %q, want empty", got)
	}
}

func TestDeref_Bool(t *testing.T) {
	b := true
	if got := Deref(&b); !got {
		t.Error("Deref(&true) = false")
	}
}

func TestDeref_NilBool(t *testing.T) {
	var p *bool
	if got := Deref(p); got {
		t.Error("Deref(nil) = true, want false")
	}
}
