package app_settings

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dan-sherwin/go-app-settings/db"
	"github.com/dan-sherwin/go-app-settings/db/models"
)

// test helpers
func resetGlobals() {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	settings = []*Setting{}
	defaultSettings = []*models.AppSetting{}
	socketPath = ""
}

func tempDBPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.db")
}

// Registers a simple string setting bound to the provided pointer
func registerStringSetting(name string, desc string, prop *string) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: desc,
		GetFunc:     func() string { return *prop },
		SetFunc:     func(s string) error { *prop = s; return nil },
	})
}

func TestGetSetting_FoundAndNotFound(t *testing.T) {
	resetGlobals()
	var foo string
	registerStringSetting("foo", "Foo setting", &foo)

	// Found
	s, err := GetSetting("foo")
	if err != nil {
		t.Fatalf("expected to find setting, got err: %v", err)
	}
	if s == nil || s.Name != "foo" {
		t.Fatalf("unexpected setting returned: %#v", s)
	}

	// Not found
	if _, err := GetSetting("bar"); err == nil {
		t.Fatalf("expected error for missing setting")
	}
}

func TestSetup_SettingPersistenceAndVars(t *testing.T) {
	resetGlobals()
	var foo string
	foo = "defaultFoo"
	registerStringSetting("foo", "Foo setting", &foo)

	// Setup with temp DB
	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Defaults copied into defaultSettings
	if len(defaultSettings) != 1 || defaultSettings[0].Key != "foo" || defaultSettings[0].Value != "defaultFoo" {
		t.Fatalf("defaultSettings not initialized correctly: %#v", defaultSettings)
	}

	// Set new value and ensure it is saved in DB
	if err := SetSetting("foo", "newFoo"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	found, err := db.AppSetting.Find()
	if err != nil {
		t.Fatalf("db find failed: %v", err)
	}
	if len(found) != 1 || found[0].Key != "foo" || found[0].Value != "newFoo" {
		t.Fatalf("unexpected DB contents: %#v", found)
	}

	// SettingsVars should reflect the in-memory value
	vars := SettingsVars()
	if vars["foo"] != "newFoo" {
		t.Fatalf("SettingsVars did not reflect value, got %q", vars["foo"])
	}
}

func TestRetrieveAppSettings_LoadsFromDB(t *testing.T) {
	resetGlobals()
	var foo string
	foo = "defaultFoo"
	registerStringSetting("foo", "Foo setting", &foo)

	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Save different value directly and reload
	if err := db.AppSetting.Save(&models.AppSetting{Key: "foo", Value: "persisted"}); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Overwrite local value to ensure Retrieve updates it
	foo = "somethingElse"
	if err := RetrieveAppSettings(); err != nil {
		t.Fatalf("RetrieveAppSettings failed: %v", err)
	}
	if foo != "persisted" {
		t.Fatalf("expected foo to be 'persisted', got %q", foo)
	}
}

func TestListRunning_GetRunningSettings(t *testing.T) {
	resetGlobals()
	var a, b string
	a = "A"
	b = "B"
	registerStringSetting("a", "A desc", &a)
	registerStringSetting("b", "B desc", &b)

	// No DB needed for GetRunningSettings
	cmd := &SettingsListRunningCommand{}
	var out []models.AppSetting
	if err := cmd.GetRunningSettings(&struct{}{}, &out); err != nil {
		t.Fatalf("GetRunningSettings error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 settings, got %d", len(out))
	}
	// Values should reflect getters
	m := map[string]string{out[0].Key: out[0].Value, out[1].Key: out[1].Value}
	if m["a"] != "A" || m["b"] != "B" {
		t.Fatalf("unexpected values: %#v", out)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestListSavedAndActiveCommands_Print(t *testing.T) {
	resetGlobals()
	var foo string
	foo = "default"
	registerStringSetting("foo", "Foo setting", &foo)
	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Save value
	if err := SetSetting("foo", "savedValue"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Saved list prints row for saved setting
	savedCmd := &SettingsListSavedCommand{}
	out := captureStdout(func() {
		_ = savedCmd.Run()
	})
	if !strings.Contains(out, "foo") || !strings.Contains(out, "savedValue") {
		t.Fatalf("saved list output unexpected: %s", out)
	}

	// Active list should reflect saved value overriding default
	activeCmd := &SettingsListActiveCommand{}
	out = captureStdout(func() {
		_ = activeCmd.Run()
	})
	if !strings.Contains(out, "foo") || !strings.Contains(out, "savedValue") {
		t.Fatalf("active list output unexpected: %s", out)
	}
}

func TestRemoveCommand_RemovesFromDB(t *testing.T) {
	resetGlobals()
	var foo string
	registerStringSetting("foo", "Foo setting", &foo)
	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := SetSetting("foo", "bar"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Remove via command
	rm := &SettingsRemoveCommand{Setting: "foo"}
	if err := rm.Run(); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Ensure DB is empty
	got, err := db.AppSetting.Find()
	if err != nil {
		t.Fatalf("db find failed: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty DB, got: %#v", got)
	}
}
