package secrets

import (
	"testing"
)

func TestFileProvider(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_DATA_HOME", d)

	p, err := newFileProvider(t.Context())
	if err != nil {
		t.Fatalf("newFileProvider() error = %v", err)
	}
	m := &genericWrapper{provider: p, namespace: "test"}

	v1, err := m.Get(t.Context(), "id/field")
	if err != nil {
		t.Errorf("fileProvider.Get(missing key) error = %v", err)
	}
	if v1 != "" {
		t.Errorf("fileProvider.Get(missing key) = %q, want %q", v1, "")
	}

	v2 := "val1"
	if err := m.Set(t.Context(), "id/field", v2); err != nil {
		t.Errorf("fileProvider.Set() error = %v", err)
	}

	v2 = "val2"
	if err := m.Set(t.Context(), "id/field", v2); err != nil {
		t.Errorf("fileProvider.Set() error = %v", err)
	}

	v1, err = m.Get(t.Context(), "id/field")
	if err != nil {
		t.Errorf("fileProvider.Get() error = %v", err)
	}
	if v1 != v2 {
		t.Errorf("fileProvider.Get() = %q, want %q", v1, v2)
	}

	if err := m.Delete(t.Context(), "id/field"); err != nil {
		t.Errorf("fileProvider.Delete() error = %v", err)
	}

	v1, err = m.Get(t.Context(), "id/field")
	if err != nil {
		t.Errorf("fileProvider.Get(missing key) error = %v", err)
	}
	if v1 != "" {
		t.Errorf("fileProvider.Get(missing key) = %q, want %q", v1, "")
	}

	if err := m.Delete(t.Context(), "id/field"); err != nil {
		t.Errorf("fileProvider.Delete(missing key) error = %v", err)
	}
}
