package app_settings

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/dan-sherwin/go-app-settings/db"
	"github.com/dan-sherwin/go-app-settings/db/models"
	"github.com/dan-sherwin/go-utilities"
	"github.com/olekukonko/tablewriter"
	"gorm.io/gorm"
)

// SettingsOptions defines configuration options for listing running settings.
type (
	SettingsOptions struct {
		RpcSocketPathToListRunningSettings string
		KongVars                           *kong.Vars
	}
	SettingsDef struct {
		Logging struct {
			Level string `enum:"debug,info,warn,error" default:"${logging_level}" help:"debug, info, warn, error" group:"logging"`
		} `embed:"" prefix:"logging."`
		Settings SettingsCommand `cmd:"" help:"Settings" group:"App Settings"`
	}
	SettingsCommand struct {
		List   SettingsListCommand   `cmd:"" help:"List settings"`
		Save   SettingsSaveCommand   `cmd:"" help:"Save settings"`
		Set    SettingsSaveCommand   `cmd:"" help:"Alias for save"`
		Remove SettingsRemoveCommand `cmd:"" help:"Remove settings"`
		Unset  SettingsRemoveCommand `cmd:"" help:"Alias for remove"`
	}

	SettingsListDefaultsCommand struct{}
	SettingsListSavedCommand    struct{}
	SettingsListRunningCommand  struct{}
	SettingsListActiveCommand   struct{}
	SettingsListCommand         struct {
		Defaults SettingsListDefaultsCommand `cmd:"" help:"List default settings"`
		Saved    SettingsListSavedCommand    `cmd:"" help:"List saved settings"`
		Running  SettingsListRunningCommand  `cmd:"" help:"List running settings"`
		Active   SettingsListActiveCommand   `cmd:"" help:"List active settings"`
	}

	SettingsSaveCommand struct {
		Setting string `arg:"" help:"Setting to set" required:""`
		Value   string `arg:"" help:"Value to set" required:""`
	}
	SettingsRemoveCommand struct {
		Setting string `arg:"" help:"Setting to remove" required:""`
	}
	Setting struct {
		SetFunc     func(string) error
		GetFunc     func() string
		Name        string
		Description string
	}

	SettingReceiver interface {
		SettingName() string
		SettingDescription() string
		SettingSet(string) error
		SettingGet() string
	}
)

// defaultSettings holds a list of default application settings of type AppSetting from the models package.
// settings contains a list of customizable settings defined by the Setting type.
// socketPath represents the file path for the application socket, initialized as an empty string.
var (
	defaultSettings = []*models.AppSetting{}
	settings        = []*Setting{}
	settingsMu      sync.RWMutex
	socketPath      = ""
)

// Setup initializes the application with the provided settings file and options.
// It configures the database, sets up the RPC socket if specified, merges Kong variables, and retrieves application settings.
// Returns an error if any initialization step fails.
func Setup(settingsFileName string, options SettingsOptions) error {
	if err := db.DBInit(settingsFileName); err != nil {
		return err
	}
	if options.RpcSocketPathToListRunningSettings != "" {
		socketPath = options.RpcSocketPathToListRunningSettings
		rpc.Register(&SettingsListRunningCommand{})
	}
	if options.KongVars != nil {
		utilities.MergeInto(*options.KongVars, SettingsVars())
	}
	return RetrieveAppSettings()
}

// getSetting retrieves a `Setting` by its name from the global list `settings`.
// If no match is found, it returns `nil`.
func getSetting(name string) (*Setting, error) {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	if !slices.ContainsFunc(settings, func(s *Setting) bool {
		return s.Name == name
	}) {
		return nil, fmt.Errorf("Setting %s not found", name)
	}
	if setting := settings[slices.IndexFunc(settings, func(s *Setting) bool {
		return s.Name == name
	})]; setting == nil {
		return nil, fmt.Errorf("Setting %s not found", name)
	} else {
		return setting, nil
	}
}

// Run removes the specified application setting if it exists, otherwise returns an error.
// It identifies the setting by name, deletes it from the database, and handles any errors encountered during the operation.
// On success, it prints a confirmation message.
func (c *SettingsRemoveCommand) Run() error {
	setting, err := getSetting(c.Setting)
	if err != nil {
		return printAndReturnErr(err)
	}
	_, err = db.AppSetting.Where(db.AppSetting.Key.Eq(setting.Name)).Delete()
	if err != nil {
		return printAndReturnErr(fmt.Errorf("Error deleting setting %s: %w", c.Setting, err))
	}
	fmt.Printf("Setting %s removed\n", c.Setting)
	return nil
}

// Run executes the command to update a specific application setting with a provided value and persists it in the database.
func (c *SettingsSaveCommand) Run() error {
	setting, err := getSetting(c.Setting)
	if err != nil {
		return printAndReturnErr(err)
	}

	if err := setting.SetFunc(c.Value); err != nil {
		return printAndReturnErr(err)
	}
	if err := db.AppSetting.Save(&models.AppSetting{
		Key:   setting.Name,
		Value: c.Value,
	}); err != nil {
		return printAndReturnErr(fmt.Errorf("Error saving setting %s: %w", c.Setting, err))
	}
	fmt.Printf("Setting %s saved to %s\n", c.Setting, c.Value)
	return nil
}

// Run connects to a Unix socket, retrieves running application settings via RPC, processes them, and displays them. It returns an error if the connection fails or if settings retrieval is unsuccessful.
func (c *SettingsListRunningCommand) Run() error {
	var runningSettings []models.AppSetting

	client, err := rpc.Dial("unix", socketPath)
	if err != nil {
		return printAndReturnErr(fmt.Errorf("Error connecting to socket: %w", err))
	}

	err = client.Call("SettingsListRunningCommand.GetRunningSettings", &struct{}{}, &runningSettings)
	if err != nil {
		return printAndReturnErr(fmt.Errorf("Error getting running settings: %w", err))
	}

	var buf []*models.AppSetting
	for _, s := range runningSettings {
		buf = append(buf, &s)
	}
	printSettings(buf)
	return nil
}

// GetRunningSettings retrieves the current running application settings and maps them into a slice of AppSetting.
// The result is assigned to the provided data pointer. Returns an error if the operation fails.
func (c *SettingsListRunningCommand) GetRunningSettings(_ *struct{}, data *[]models.AppSetting) error {
	runningSettings := []models.AppSetting{}
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	for _, s := range settings {
		runningSettings = append(runningSettings, models.AppSetting{
			Key:         s.Name,
			Value:       s.GetFunc(),
			Description: s.Description,
		})
	}
	*data = runningSettings
	return nil
}

// Run executes the command to display default application settings in a sorted table format. It uses the printSettings function to handle the output of pre-defined settings. Returns nil upon successful execution.
func (c *SettingsListDefaultsCommand) Run() error {
	printSettings(defaultSettings)
	return nil
}

// Run executes the command to retrieve and display saved settings. It fetches settings from the database, handles errors, and prints the retrieved settings to the console. Returns an error if there is an issue during the retrieval process.
func (c *SettingsListSavedCommand) Run() error {
	s, err := db.AppSetting.Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("Error getting saved settings: %w", err)
	}
	savedSettings := []models.AppSetting{}
	for _, as := range s {
		setting, err := getSetting(as.Key)
		if err != nil {
			continue
		}
		as.Description = setting.Description
		savedSettings = append(savedSettings, *as)
	}
	printSettings(s)
	return nil
}

func (c *SettingsListActiveCommand) Run() error {
	s, err := db.AppSetting.Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("Error getting saved settings: %w", err)
	}
	activeSettings := defaultSettings
	for _, ss := range s {
		setting, err := getSetting(ss.Key)
		if err != nil {
			continue
		}
		for _, as := range activeSettings {
			if as.Key == ss.Key {
				as.Value = ss.Value
				as.Description = setting.Description
			}
		}
	}
	printSettings(activeSettings)
	return nil
}

// printSettings displays a sorted list of application settings in a tabular format showing their keys and values.
func printSettings(settings []*models.AppSetting) {
	slices.SortFunc(settings, func(a, b *models.AppSetting) int {
		return strings.Compare(a.Key, b.Key)
	})
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Setting", "Value", "Description"})
	for _, s := range settings {
		table.Append([]string{s.Key, s.Value, s.Description})
	}
	table.Render()
}

// SettingsVars constructs a kong.Vars map by iterating through all settings, retrieving their values using associated getters, and populating the map with setting names as keys and their retrieved values as values.
func SettingsVars() kong.Vars {
	vars := kong.Vars{}
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	for _, s := range settings {
		vars[s.Name] = s.GetFunc()
	}
	return vars
}

// RetrieveAppSettings fetches application settings from the database and initializes default settings.
// If database retrieval fails, the application exits with an error.
// It also updates in-memory settings based on the retrieved values from the database.
func RetrieveAppSettings() error {
	defaultSettings = []*models.AppSetting{}
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	for _, s := range settings {
		defaultSettings = append(defaultSettings, &models.AppSetting{
			Key:         s.Name,
			Value:       s.GetFunc(),
			Description: s.Description,
		})
	}
	appSettings, err := db.AppSetting.Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("Error getting app settings: %w", err)
	}
	errs := []error{}
	if appSettings != nil {
		for _, as := range appSettings {
			if s, err := getSetting(as.Key); err == nil {
				err := s.SetFunc(as.Value)
				if err != nil {
					errs = append(errs, fmt.Errorf("Error setting setting %s: %w", as.Key, err))
				}
			}
		}
	}
	if len(errs) > 0 {
		errorText := ""
		for _, err := range errs {
			errorText += err.Error() + "\n"
		}
		return fmt.Errorf("errors setting settings: %s", errorText)
	}
	return nil
}

func printAndReturnErr(err error) error {
	fmt.Println(err)
	return err
}
