package main

import "testing"

func TestRiskAction(t *testing.T) {
	cases := []struct {
		name string
		meta map[string]string
		want string
	}{
		{"high risk", map[string]string{"risk": "high"}, "review"},
		{"reversed payment", map[string]string{"status": "reversed"}, "review"},
		{"normal payment", map[string]string{"risk": "low", "status": "settled"}, "allow"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := riskAction(tc.meta); got != tc.want {
				t.Fatalf("riskAction() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsCollectionExists(t *testing.T) {
	if !isCollectionExists(&apiError{Code: "VECTOR_COLLECTION_EXISTS", Message: "already exists"}) {
		t.Fatal("expected VECTOR_COLLECTION_EXISTS to be accepted")
	}
	if isCollectionExists(&apiError{Code: "UNAUTHORIZED", Message: "bad key"}) {
		t.Fatal("unexpected API error was accepted")
	}
}
