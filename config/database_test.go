package config

import "testing"

func TestDatabaseValidation(t *testing.T) {
	for _, name := range []string{"", "public", "custom_1", "_private"} {
		db := Database{Schema: name}
		if errors := db.Validate(); len(errors) != 0 {
			t.Fatalf("%q: %v", name, errors)
		}
		if db.Schema == "" {
			t.Fatal("empty default schema")
		}
	}
	for _, name := range []string{"Upper", "a-b", "1bad", "a;DROP SCHEMA public"} {
		db := Database{Schema: name}
		if len(db.Validate()) == 0 {
			t.Fatalf("%q accepted", name)
		}
	}
	db := Database{Type: "sqlite"}
	if len(db.Validate()) == 0 {
		t.Fatal("sqlite accepted")
	}
}
