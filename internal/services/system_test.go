package services

import "testing"

func TestCheckFiltersCompability_EmptyFilters(t *testing.T) {
	ok, err := checkFiltersCompability(InfoTypes.CPU, []InfoFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected compatible=true for empty filters")
	}

	ok, err = checkFiltersCompability(InfoTypes.RAM, []InfoFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected compatible=true for empty filters")
	}

	ok, err = checkFiltersCompability(InfoTypes.Net, []InfoFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected compatible=true for empty filters")
	}

	ok, err = checkFiltersCompability(InfoTypes.File, []InfoFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected compatible=true for empty filters")
	}

}
