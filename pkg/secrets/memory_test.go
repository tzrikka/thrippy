package secrets

import (
	"testing"
)

func TestInMemoryProvider(t *testing.T) {
	m := NewTestManager()

	v1, err := m.Get(t.Context(), "key")
	if err != nil {
		t.Errorf("inMemoryProvider.Get(missing key) error = %v", err)
	}
	if v1 != "" {
		t.Errorf("inMemoryProvider.Get(missing key) = %q, want %q", v1, "")
	}

	v2 := "val1"
	if err := m.Set(t.Context(), "key", v2); err != nil {
		t.Errorf("inMemoryProvider.Set() error = %v", err)
	}

	v2 = "val2"
	if err := m.Set(t.Context(), "key", v2); err != nil {
		t.Errorf("inMemoryProvider.Set() error = %v", err)
	}

	v1, err = m.Get(t.Context(), "key")
	if err != nil {
		t.Errorf("inMemoryProvider.Get() error = %v", err)
	}
	if v1 != v2 {
		t.Errorf("inMemoryProvider.Get() = %q, want %q", v1, v2)
	}

	if err := m.Delete(t.Context(), "key"); err != nil {
		t.Errorf("inMemoryProvider.Delete() error = %v", err)
	}

	v1, err = m.Get(t.Context(), "key")
	if err != nil {
		t.Errorf("inMemoryProvider.Get(missing key) error = %v", err)
	}
	if v1 != "" {
		t.Errorf("inMemoryProvider.Get(missing key) = %q, want %q", v1, "")
	}

	if err := m.Delete(t.Context(), "key"); err != nil {
		t.Errorf("inMemoryProvider.Delete(missing key) error = %v", err)
	}
}
