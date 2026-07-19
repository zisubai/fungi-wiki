package migrations

import "testing"

func TestNamesAreOrdered(t *testing.T) {
	names, err := Names()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 4 {
		t.Fatalf("expected at least four migrations, got %v", names)
	}
	for index := 1; index < len(names); index++ {
		if names[index-1] >= names[index] {
			t.Fatalf("migrations are not ordered: %v", names)
		}
	}
}

func TestChecksumIsStable(t *testing.T) {
	first := checksumOf([]byte("SELECT 1;"))
	second := checksumOf([]byte("SELECT 1;"))
	if first != second || len(first) != 64 {
		t.Fatalf("unexpected checksum %q", first)
	}
}
