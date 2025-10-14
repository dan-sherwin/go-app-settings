

# App Settings Package

The `app_settings` package provides a standardized way to register, manage, and persist application settings. It supports integration with a Kong-based CLI interface and optional RPC access for listing live in-memory settings.

## Features

- Register settings dynamically from any package
- Save and remove settings via CLI
- List default, saved, and running settings
- Retrieve current values programmatically
- Integrate with Kong CLI commands
- Expose settings via RPC socket (optional)

---

## Installation

Import the module in your Go project:

```go
import "github.com/dan-sherwin/go-app-settings"
```

---

## Setup

Call `Setup()` early in your application initialization:

```go
vars := kong.Vars{}
err := app_settings.Setup("myapp.db", app_settings.SettingsOptions{
    RpcSocketPathToListRunningSettings: "/tmp/myapp.sock",
    KongVars:                           &vars,
})
if err != nil {
    log.Fatalf("setup failed: %v", err)
}
```

---

## Registering Settings

You can register settings either by providing a `Setting` directly or by implementing the `SettingReceiver` interface.

### Option 1: Direct Registration

```go
var foobar string

app_settings.RegisterSetting(&app_settings.Setting{
    Name:        "foobar",
    Description: "The setting of foobar",
    GetFunc: func() string { return foobar },
    SetFunc: func(s string) error {
        foobar = s
        return nil
    },
})
```

### Option 2: Struct-Based Receiver

```go
type Feebar struct{}

func (Feebar) SettingName() string        { return "feebar" }
func (Feebar) SettingDescription() string { return "The setting of feebar" }
func (Feebar) SettingSet(s string) error  { feebar = s; return nil }
func (Feebar) SettingGet() string         { return feebar }

func init() {
    app_settings.RegisterSettingReceiver(&Feebar{})
}
```

---

## Kong CLI Integration

Add `SettingsDef` to your Kong configuration struct:

```go
var CLIConfig struct {
    app_settings.SettingsDef
    // your other commands
}
```

Kong will automatically wire in the following CLI structure:

```bash
myapp settings list defaults
myapp settings list saved
myapp settings list running
myapp settings save <setting> <value>
myapp settings remove <setting>
```

---

## Retrieving Settings in Code

To retrieve the current value of a setting programmatically:

```go
val := app_settings.SettingsVars()["foobar"]
```

---

## RPC Access

If `RpcSocketPathToListRunningSettings` is specified in `Setup()`, the running settings can be accessed via RPC. This enables external tooling or dashboards to introspect current config state at runtime.

```go
client, _ := rpc.Dial("unix", "/tmp/myapp.sock")
var running []models.AppSetting
client.Call("SettingsListRunningCommand.GetRunningSettings", &struct{}{}, &running)
```

---

## Registration Helper Functions

**RegisterBoolSetting**  
```go
func RegisterBoolSetting(name, description string, prop *bool)
```  
Registers a boolean setting with a specified name, description, and pointer to the bool property.

**RegisterDurationSetting**  
```go
func RegisterDurationSetting(name, description string, prop *time.Duration)
```  
Registers a `time.Duration` setting with a specified name, description, and pointer to the `time.Duration` property.

**RegisterFloat32Setting**  
```go
func RegisterFloat32Setting(name, description string, prop *float32)
```  
Registers a 32-bit float setting with a specified name, description, and pointer to the `float32` property.

**RegisterFloatSetting**  
```go
func RegisterFloatSetting(name, description string, prop *float64)
```  
Registers a `float64` setting with a specified name, description, and pointer to the `float64` property.

**RegisterInt32Setting**  
```go
func RegisterInt32Setting(name, description string, prop *int32)
```  
Registers a 32-bit integer setting with a specified name, description, and pointer to the `int32` property.

**RegisterInt64Setting**  
```go
func RegisterInt64Setting(name, description string, prop *int64)
```  
Registers a 64-bit integer setting with a specified name, description, and pointer to the `int64` property.

**RegisterInt8Setting**  
```go
func RegisterInt8Setting(name, description string, prop *int8)
```  
Registers an 8-bit integer setting with a specified name, description, and pointer to the `int8` property.

**RegisterIntSetting**  
```go
func RegisterIntSetting(name, description string, prop *int)
```  
Registers an integer setting with a specified name, description, and pointer to the `int` property.

**RegisterIPNetSetting**  
```go
func RegisterIPNetSetting(name, description string, prop *net.IPNet)
```  
Registers a `net.IPNet` setting (CIDR format) with a specified name, description, and pointer to the `net.IPNet` property.

**RegisterIPSetting**  
```go
func RegisterIPSetting(name, description string, prop *net.IP)
```  
Registers a `net.IP` setting with a specified name, description, and pointer to the `net.IP` property.

**RegisterStringSetting**  
```go
func RegisterStringSetting(name, description string, prop *string)
```  
Registers a string setting with a specified name, description, and pointer to the `string` property.

**RegisterStringSliceSetting**  
```go
func RegisterStringSliceSetting(name, description string, prop *[]string)
```  
Registers a string slice setting (comma-separated) with a specified name, description, and pointer to the `[]string` property.

**RegisterTimeSetting**  
```go
func RegisterTimeSetting(name, description string, prop *time.Time)
```  
Registers a `time.Time` setting (RFC3339 format) with a specified name, description, and pointer to the `time.Time` property.

**RegisterUint16Setting**  
```go
func RegisterUint16Setting(name, description string, prop *uint16)
```  
Registers a 16-bit unsigned integer setting with a specified name, description, and pointer to the `uint16` property.

**RegisterUint32Setting**  
```go
func RegisterUint32Setting(name, description string, prop *uint32)
```  
Registers a 32-bit unsigned integer setting with a specified name, description, and pointer to the `uint32` property.

**RegisterUint64Setting**  
```go
func RegisterUint64Setting(name, description string, prop *uint64)
```  
Registers a 64-bit unsigned integer setting with a specified name, description, and pointer to the `uint64` property.

**RegisterUint8Setting**  
```go
func RegisterUint8Setting(name, description string, prop *uint8)
```  
Registers an 8-bit unsigned integer setting with a specified name, description, and pointer to the `uint8` property.

**RegisterUintSetting**  
```go
func RegisterUintSetting(name, description string, prop *uint)
```  
Registers an unsigned integer setting with a specified name, description, and pointer to the `uint` property.

**RegisterURLSetting**  
```go
func RegisterURLSetting(name, description string, prop *url.URL)
```  
Registers a `url.URL` setting with a specified name, description, and pointer to the `url.URL` property.


---

## Testing

This repository includes unit tests that can be executed locally and in any CI environment that has Go installed.

- Minimum Go version: 1.22
- Run all tests:

```bash
go test ./...
```

The tests use a temporary on-disk SQLite database via GORM; no external services are required. They also capture stdout for commands that print tables.

### Example: GitHub Actions CI

If you want to run the tests in GitHub Actions, you can use a minimal workflow like this:

```yaml
name: go-test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22.x'
      - run: go test ./...
```

---

## License

ISC