package app_settings

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dan-sherwin/go-app-settings/db"
	"github.com/dan-sherwin/go-app-settings/db/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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

func TestHiddenSettings_AreNotAvailableThroughCLI(t *testing.T) {
	resetGlobals()
	visible := "visible-default"
	hidden := "hidden-default"
	RegisterStringSetting("visible", "Visible setting", &visible)
	RegisterSetting(&Setting{
		Name:        "hidden",
		Description: "Hidden setting",
		Hidden:      true,
		GetFunc:     func() string { return hidden },
		SetFunc: func(s string) error {
			hidden = s
			return nil
		},
	})

	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := SetSetting("hidden", "hidden-saved"); err != nil {
		t.Fatalf("programmatic SetSetting for hidden setting failed: %v", err)
	}
	if hidden != "hidden-saved" {
		t.Fatalf("hidden setting was not updated, got %q", hidden)
	}

	defaultsOut := captureStdout(func() {
		_ = (&SettingsListDefaultsCommand{}).Run()
	})
	if strings.Contains(defaultsOut, "hidden") || !strings.Contains(defaultsOut, "visible") {
		t.Fatalf("defaults output did not hide hidden setting: %s", defaultsOut)
	}

	running := []models.AppSetting{}
	if err := (&SettingsListRunningCommand{}).GetRunningSettings(&struct{}{}, &running); err != nil {
		t.Fatalf("GetRunningSettings failed: %v", err)
	}
	for _, s := range running {
		if s.Key == "hidden" {
			t.Fatalf("running settings included hidden setting: %#v", running)
		}
	}

	if err := (&SettingsSaveCommand{Setting: "hidden", Value: "cli-value"}).Run(); err == nil {
		t.Fatalf("expected CLI save for hidden setting to fail")
	}
	if err := (&SettingsRemoveCommand{Setting: "hidden"}).Run(); err == nil {
		t.Fatalf("expected CLI remove for hidden setting to fail")
	}
	if got := SettingsVars()["hidden"]; got != "hidden-saved" {
		t.Fatalf("SettingsVars should keep hidden settings available to the app, got %q", got)
	}
}

func TestRegisterJSONSetting_SettingPersistenceAndValidation(t *testing.T) {
	resetGlobals()
	type launchLayout struct {
		ViewMode string `json:"viewMode"`
		Columns  int    `json:"columns"`
	}
	layout := launchLayout{ViewMode: "list", Columns: 1}
	RegisterJSONSettingWithValidator("layout", "Launch layout", &layout, func(value launchLayout) error {
		if value.Columns < 1 {
			return os.ErrInvalid
		}
		return nil
	})

	if err := Setup(tempDBPath(t), SettingsOptions{}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if got := SettingsVars()["layout"]; got != `{"viewMode":"list","columns":1}` {
		t.Fatalf("unexpected default JSON setting: %s", got)
	}
	if err := SetSetting("layout", launchLayout{ViewMode: "grid", Columns: 4}); err != nil {
		t.Fatalf("SetSetting with struct failed: %v", err)
	}
	if layout.ViewMode != "grid" || layout.Columns != 4 {
		t.Fatalf("layout was not updated from struct value: %#v", layout)
	}
	if err := SetSetting("layout", `{"viewMode":"list","columns":2}`); err != nil {
		t.Fatalf("SetSetting with JSON string failed: %v", err)
	}
	if layout.ViewMode != "list" || layout.Columns != 2 {
		t.Fatalf("layout was not updated from JSON string: %#v", layout)
	}
	if err := SetSetting("layout", `{"viewMode":"grid","columns":0}`); err == nil {
		t.Fatalf("expected validation error for invalid JSON setting")
	}
}

func TestSetupWithDB_UsesApplicationDatabase(t *testing.T) {
	resetGlobals()
	type LaunchItem struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	var foo string
	foo = "default"
	RegisterStringSetting("foo", "Foo setting", &foo)

	gormDB, err := gorm.Open(sqlite.Open(tempDBPath(t)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open app db failed: %v", err)
	}
	if err := gormDB.AutoMigrate(&LaunchItem{}); err != nil {
		t.Fatalf("app migration failed: %v", err)
	}
	if err := SetupWithDB(gormDB, SettingsOptions{}); err != nil {
		t.Fatalf("SetupWithDB failed: %v", err)
	}
	if !gormDB.Migrator().HasTable("launch_items") {
		t.Fatalf("expected app table to remain available")
	}
	if !gormDB.Migrator().HasTable(DefaultTableName) {
		t.Fatalf("expected app_settings table in app database")
	}
	if err := SetSetting("foo", "persisted"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	var count int64
	if err := gormDB.Table(DefaultTableName).Where("key = ? AND value = ?", "foo", "persisted").Count(&count).Error; err != nil {
		t.Fatalf("query app_settings failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected setting in shared app database, count=%d", count)
	}
}

func TestSetupWithDB_TableConflictAndCustomTableName(t *testing.T) {
	resetGlobals()
	var foo string
	RegisterStringSetting("foo", "Foo setting", &foo)

	gormDB, err := gorm.Open(sqlite.Open(tempDBPath(t)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open app db failed: %v", err)
	}
	if err := gormDB.Exec("CREATE TABLE app_settings (id integer primary key, bogus text)").Error; err != nil {
		t.Fatalf("create conflicting table failed: %v", err)
	}
	if err := SetupWithDB(gormDB, SettingsOptions{}); err == nil {
		t.Fatalf("expected incompatible app_settings table to fail setup")
	}
	if err := SetupWithDB(gormDB, SettingsOptions{TableName: "runtime_settings"}); err != nil {
		t.Fatalf("SetupWithDB with custom table name failed: %v", err)
	}
	if !gormDB.Migrator().HasTable("runtime_settings") {
		t.Fatalf("expected custom settings table")
	}
}
