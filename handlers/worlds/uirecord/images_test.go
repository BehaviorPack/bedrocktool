package uirecord

import (
	"strings"
	"testing"
)

func TestImagePath(t *testing.T) {
	for _, c := range []struct {
		url, want string
		ok        bool
	}{
		{"https://cdn.playhive.com/icons/locker/build/cape/sweet-treat.png", "icons/locker/build/cape/sweet-treat", true},
		{"https://cdn.playhive.com/icons/locker/locked.png?v=2", "icons/locker/locked", true},
		{"https://example.com/a/../../b.png", "b", true},
		{"file:///etc/passwd", "", false},
		{"textures/ui/x", "", false},
		{"https://cdn.playhive.com/", "", false},
		{"https://cdn.playhive.com/a%20b.png", "", false},
	} {
		got, ok := ImagePath(c.url)
		if got != c.want || ok != c.ok {
			t.Errorf("ImagePath(%q) = %q, %v; want %q, %v", c.url, got, ok, c.want, c.ok)
		}
	}
}

func TestScriptShowsURLImages(t *testing.T) {
	r := New(nil)
	r.forms["f1"] = &Form{Key: "f1", Data: map[string]any{
		"type":  "form",
		"title": "PETS§m§8",
		"buttons": []any{
			map[string]any{"text": "Cape", "image": map[string]any{"type": "url", "data": "https://cdn.playhive.com/icons/locker/build/cape/sweet-treat.png"}},
			map[string]any{"text": "Vanilla", "image": map[string]any{"type": "path", "data": "textures/ui/x"}},
			map[string]any{"text": "None"},
		},
	}}
	r.formOrder = append(r.formOrder, "f1")

	images := r.Images()
	if len(images) != 1 || images["https://cdn.playhive.com/icons/locker/build/cape/sweet-treat.png"] != "icons/locker/build/cape/sweet-treat" {
		t.Errorf("Images() = %v", images)
	}
	script := r.Script()
	for _, s := range []string{
		`const IMAGES = {`,
		`"https://cdn.playhive.com/icons/locker/build/cape/sweet-treat.png": "icons/locker/build/cape/sweet-treat"`,
		`IMAGES[b.image.data]`,
		`PETS§m§8`,
	} {
		if !strings.Contains(script, s) {
			t.Errorf("script lacks %s", s)
		}
	}
}
