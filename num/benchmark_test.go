package num

import (
	"encoding/json"
	"testing"
)

func BenchmarkFromString(b *testing.B) {
	for b.Loop() {
		FromString("123.456789")
	}
}

func BenchmarkFromInt64(b *testing.B) {
	for b.Loop() {
		FromInt64(123456)
	}
}

func BenchmarkFromFloat64(b *testing.B) {
	for b.Loop() {
		FromFloat64(123.456789)
	}
}

func BenchmarkArithmeticChain(b *testing.B) {
	a := FromString("100.50")
	c := FromString("10")
	d := FromString("3")
	for b.Loop() {
		_ = a.Add(c).Mul(d).Sub(a).Div(c)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	n := FromString("123.456789")
	for b.Loop() {
		_, _ = json.Marshal(n)
	}
}

func BenchmarkUnmarshalJSON_Bare(b *testing.B) {
	data := []byte("123.456789")
	for b.Loop() {
		var n Num
		_ = n.UnmarshalJSON(data)
	}
}

func BenchmarkUnmarshalJSON_Quoted(b *testing.B) {
	data := []byte(`"123.456789"`)
	for b.Loop() {
		var n Num
		_ = n.UnmarshalJSON(data)
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		var n Num
		_ = n.Scan("123.456789")
	}
}

func BenchmarkValue(b *testing.B) {
	n := FromString("123.456789")
	for b.Loop() {
		_, _ = n.Value()
	}
}
