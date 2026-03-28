package num

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE test_nums (
		id    INTEGER PRIMARY KEY,
		price TEXT,
		strike TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func sqlRoundtripScaleOf(n Num) int {
	return n.dec.Scale()
}

func TestNumSQLRoundtrip(t *testing.T) {
	db := openTestDB(t)

	original := FromString("123.456789")
	_, err := db.Exec("INSERT INTO test_nums (id, price) VALUES (?, ?)", 1, original)
	if err != nil {
		t.Fatal(err)
	}

	var restored Num
	if err := db.QueryRow("SELECT price FROM test_nums WHERE id = 1").Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Ok() {
		t.Fatalf("scan error: %v", restored.Err)
	}
	if !original.Equal(restored) {
		t.Errorf("roundtrip: got %s, want %s", restored.dec.String(), original.dec.String())
	}
	if got := sqlRoundtripScaleOf(restored); got != Scale() {
		t.Errorf("scale = %d, want %d", got, Scale())
	}
}

func TestNullNumSQLRoundtrip_Valid(t *testing.T) {
	db := openTestDB(t)

	original := NullNum{Num: FromString("99.99"), Valid: true}
	_, err := db.Exec("INSERT INTO test_nums (id, strike) VALUES (?, ?)", 1, original)
	if err != nil {
		t.Fatal(err)
	}

	var restored NullNum
	if err := db.QueryRow("SELECT strike FROM test_nums WHERE id = 1").Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid {
		t.Fatal("expected Valid=true after roundtrip")
	}
	if !restored.Num.Ok() {
		t.Fatalf("scan error: %v", restored.Num.Err)
	}
	if !original.Num.Equal(restored.Num) {
		t.Errorf("roundtrip: got %s, want %s",
			restored.Num.dec.String(), original.Num.dec.String())
	}
}

func TestNullNumSQLRoundtrip_Null(t *testing.T) {
	db := openTestDB(t)

	null := NullNum{Valid: false}
	_, err := db.Exec("INSERT INTO test_nums (id, strike) VALUES (?, ?)", 1, null)
	if err != nil {
		t.Fatal(err)
	}

	var restored NullNum
	if err := db.QueryRow("SELECT strike FROM test_nums WHERE id = 1").Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Errorf("expected Valid=false after NULL roundtrip, got %s", restored.Num.dec.String())
	}
}

func TestNullNumSQLRoundtrip_MixedRow(t *testing.T) {
	db := openTestDB(t)

	price := FromString("42.50")
	strike := NullNum{Valid: false}
	_, err := db.Exec("INSERT INTO test_nums (id, price, strike) VALUES (?, ?, ?)", 1, price, strike)
	if err != nil {
		t.Fatal(err)
	}

	var rPrice Num
	var rStrike NullNum
	if err := db.QueryRow("SELECT price, strike FROM test_nums WHERE id = 1").Scan(&rPrice, &rStrike); err != nil {
		t.Fatal(err)
	}

	if !rPrice.Ok() || !rPrice.Equal(price) {
		t.Errorf("price roundtrip: got %s, want %s", rPrice.dec.String(), price.dec.String())
	}
	if rStrike.Valid {
		t.Error("strike should be NULL")
	}
}
