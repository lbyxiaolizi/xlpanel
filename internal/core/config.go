package core

type AppConfig struct {
	DefaultTheme       string
	AllowThemeOverride bool
}

var DefaultConfig = AppConfig{
	DefaultTheme:       "classic",
	AllowThemeOverride: true,
}
