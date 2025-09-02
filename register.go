package app_settings

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// RegisterSettingReceiver registers a SettingReceiver by wrapping its methods in a Setting struct and appending it to the global settings list.
func RegisterSettingReceiver(r SettingReceiver) {
	RegisterSetting(&Setting{
		SetFunc:     r.SettingSet,
		GetFunc:     r.SettingGet,
		Name:        r.SettingName(),
		Description: r.SettingDescription(),
	})
}

// RegisterSetting adds a given Setting to the global settings list.
func RegisterSetting(s *Setting) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	settings = append(settings, s)
}

// RegisterStringSetting registers a string setting with a specified name, description, and a pointer to the property to manage its value.
func RegisterStringSetting(name string, description string, prop *string) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return *prop
		},
		SetFunc: func(s string) error {
			*prop = s
			return nil
		},
	})
}

// RegisterIntSetting registers an integer setting with a name, description, and a pointer to the integer property.
func RegisterIntSetting(name string, description string, prop *int) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.Itoa(*prop)
		},
		SetFunc: func(s string) error {
			i, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			*prop = i
			return nil
		},
	})
}

// RegisterBoolSetting registers a boolean setting with a specified name, description, and pointer to the bool property.
func RegisterBoolSetting(name string, description string, prop *bool) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatBool(*prop)
		},
		SetFunc: func(s string) error {
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			*prop = b
			return nil
		},
	})
}

// RegisterUintSetting registers an unsigned integer setting with a specified name, description, and pointer to the uint property.
func RegisterUintSetting(name string, description string, prop *uint) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatUint(uint64(*prop), 10)
		},
		SetFunc: func(s string) error {
			u, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			*prop = uint(u)
			return nil
		},
	})
}

// RegisterFloatSetting registers a float64 setting with a specified name, description, and pointer to the float64 property.
func RegisterFloatSetting(name string, description string, prop *float64) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatFloat(*prop, 'f', -1, 64)
		},
		SetFunc: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			*prop = f
			return nil
		},
	})
}

// RegisterDurationSetting registers a time.Duration setting with a specified name, description, and pointer to the time.Duration property.
func RegisterDurationSetting(name string, description string, prop *time.Duration) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return prop.String()
		},
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			*prop = d
			return nil
		},
	})
}

// RegisterInt32Setting registers a 32-bit integer setting with a specified name, description, and pointer to the int32 property.
func RegisterInt32Setting(name string, description string, prop *int32) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatInt(int64(*prop), 10)
		},
		SetFunc: func(s string) error {
			i, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			*prop = int32(i)
			return nil
		},
	})
}

// RegisterInt64Setting registers a 64-bit integer setting with a specified name, description, and pointer to the int64 property.
func RegisterInt64Setting(name string, description string, prop *int64) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatInt(*prop, 10)
		},
		SetFunc: func(s string) error {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			*prop = i
			return nil
		},
	})
}

// RegisterFloat32Setting registers a 32-bit float setting with a specified name, description, and pointer to the float32 property.
func RegisterFloat32Setting(name string, description string, prop *float32) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc: func() string {
			return strconv.FormatFloat(float64(*prop), 'f', -1, 32)
		},
		SetFunc: func(s string) error {
			f, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return err
			}
			*prop = float32(f)
			return nil
		},
	})
}

// RegisterInt8Setting registers an 8-bit integer setting.
func RegisterInt8Setting(name, description string, prop *int8) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatInt(int64(*prop), 10) },
		SetFunc: func(s string) error {
			i, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return err
			}
			*prop = int8(i)
			return nil
		},
	})
}

// RegisterInt16Setting registers a 16-bit integer setting.
func RegisterInt16Setting(name, description string, prop *int16) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatInt(int64(*prop), 10) },
		SetFunc: func(s string) error {
			i, err := strconv.ParseInt(s, 10, 16)
			if err != nil {
				return err
			}
			*prop = int16(i)
			return nil
		},
	})
}

// RegisterUint8Setting registers an 8-bit unsigned integer setting.
func RegisterUint8Setting(name, description string, prop *uint8) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatUint(uint64(*prop), 10) },
		SetFunc: func(s string) error {
			u, err := strconv.ParseUint(s, 10, 8)
			if err != nil {
				return err
			}
			*prop = uint8(u)
			return nil
		},
	})
}

// RegisterUint16Setting registers a 16-bit unsigned integer setting.
func RegisterUint16Setting(name, description string, prop *uint16) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatUint(uint64(*prop), 10) },
		SetFunc: func(s string) error {
			u, err := strconv.ParseUint(s, 10, 16)
			if err != nil {
				return err
			}
			*prop = uint16(u)
			return nil
		},
	})
}

// RegisterUint32Setting registers a 32-bit unsigned integer setting.
func RegisterUint32Setting(name, description string, prop *uint32) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatUint(uint64(*prop), 10) },
		SetFunc: func(s string) error {
			u, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return err
			}
			*prop = uint32(u)
			return nil
		},
	})
}

// RegisterUint64Setting registers a 64-bit unsigned integer setting.
func RegisterUint64Setting(name, description string, prop *uint64) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strconv.FormatUint(*prop, 10) },
		SetFunc: func(s string) error {
			u, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			*prop = u
			return nil
		},
	})
}

// RegisterTimeSetting registers a time.Time setting using RFC3339 format.
func RegisterTimeSetting(name, description string, prop *time.Time) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return prop.Format(time.RFC3339) },
		SetFunc: func(s string) error {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return err
			}
			*prop = t
			return nil
		},
	})
}

// RegisterStringSliceSetting registers a string slice setting (comma-separated).
func RegisterStringSliceSetting(name, description string, prop *[]string) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return strings.Join(*prop, ",") },
		SetFunc: func(s string) error {
			if s == "" {
				*prop = nil
			} else {
				*prop = strings.Split(s, ",")
			}
			return nil
		},
	})
}

// RegisterIPSetting registers a net.IP setting.
func RegisterIPSetting(name, description string, prop *net.IP) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return prop.String() },
		SetFunc: func(s string) error {
			ip := net.ParseIP(s)
			if ip == nil {
				return fmt.Errorf("invalid IP: %q", s)
			}
			*prop = ip
			return nil
		},
	})
}

// RegisterIPNetSetting registers a net.IPNet setting (CIDR format).
func RegisterIPNetSetting(name, description string, prop *net.IPNet) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return prop.String() },
		SetFunc: func(s string) error {
			_, ipnet, err := net.ParseCIDR(s)
			if err != nil {
				return err
			}
			*prop = *ipnet
			return nil
		},
	})
}

// RegisterURLSetting registers a url.URL setting.
func RegisterURLSetting(name, description string, prop *url.URL) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return prop.String() },
		SetFunc: func(s string) error {
			u, err := url.Parse(s)
			if err != nil {
				return err
			}
			*prop = *u
			return nil
		},
	})
}

func RegisterCronSetting(name, description string, cronString *string) {
	RegisterSetting(&Setting{
		Name:        name,
		Description: description,
		GetFunc:     func() string { return *cronString },
		SetFunc: func(s string) error {
			parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
			if _, err := parser.Parse(s); err != nil {
				return fmt.Errorf("invalid cron expression: %w", err)
			}
			*cronString = s
			return nil
		},
	})
}
