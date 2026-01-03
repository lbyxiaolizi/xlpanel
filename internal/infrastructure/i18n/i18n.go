// Package i18n provides internationalization support for OpenHost.
// It supports multiple languages with JSON-based translation files and
// provides a clean API for use in templates and handlers.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Language represents a supported language
type Language struct {
	Code       string `json:"code"`        // ISO 639-1 code (e.g., "en", "zh")
	Name       string `json:"name"`        // English name (e.g., "English", "Chinese")
	NativeName string `json:"native_name"` // Native name (e.g., "English", "中文")
	Direction  string `json:"direction"`   // Text direction ("ltr" or "rtl")
	Flag       string `json:"flag"`        // Flag emoji
}

// Translator handles translations for a specific language
type Translator struct {
	lang         string
	translations map[string]string
	fallback     *Translator
}

// Manager manages all translations and languages
type Manager struct {
	mu           sync.RWMutex
	languages    map[string]*Language
	translators  map[string]*Translator
	defaultLang  string
	fallbackLang string
	basePath     string
}

var (
	defaultManager *Manager
	once           sync.Once
)

// Default returns the default i18n manager
func Default() *Manager {
	once.Do(func() {
		defaultManager = NewManager("./locales")
	})
	return defaultManager
}

// SetDefault sets the default i18n manager
func SetDefault(m *Manager) {
	defaultManager = m
}

// NewManager creates a new i18n manager
func NewManager(basePath string) *Manager {
	m := &Manager{
		languages:    make(map[string]*Language),
		translators:  make(map[string]*Translator),
		defaultLang:  "en",
		fallbackLang: "en",
		basePath:     basePath,
	}
	return m
}

// SetDefaultLanguage sets the default language
func (m *Manager) SetDefaultLanguage(lang string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultLang = lang
}

// SetFallbackLanguage sets the fallback language
func (m *Manager) SetFallbackLanguage(lang string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackLang = lang
}

// LoadLanguage loads a language from a JSON file
func (m *Manager) LoadLanguage(lang string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load language metadata
	metaPath := filepath.Join(m.basePath, lang, "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		// Create default metadata if not found
		m.languages[lang] = &Language{
			Code:       lang,
			Name:       lang,
			NativeName: lang,
			Direction:  "ltr",
		}
	} else {
		var langMeta Language
		if err := json.Unmarshal(metaData, &langMeta); err != nil {
			return fmt.Errorf("parse language meta %s: %w", lang, err)
		}
		m.languages[lang] = &langMeta
	}

	// Load translations
	translations := make(map[string]string)
	langDir := filepath.Join(m.basePath, lang)

	err = filepath.Walk(langDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") || info.Name() == "meta.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read translation file %s: %w", path, err)
		}

		var fileTranslations map[string]any
		if err := json.Unmarshal(data, &fileTranslations); err != nil {
			return fmt.Errorf("parse translation file %s: %w", path, err)
		}

		// Get the namespace from the filename (without .json)
		namespace := strings.TrimSuffix(info.Name(), ".json")
		flattenTranslations(fileTranslations, namespace, translations)

		return nil
	})

	if err != nil {
		return fmt.Errorf("load translations for %s: %w", lang, err)
	}

	translator := &Translator{
		lang:         lang,
		translations: translations,
	}

	// Set fallback if available
	if lang != m.fallbackLang {
		if fb, ok := m.translators[m.fallbackLang]; ok {
			translator.fallback = fb
		}
	}

	m.translators[lang] = translator

	// Update fallback references for other translators
	if lang == m.fallbackLang {
		for code, t := range m.translators {
			if code != lang {
				t.fallback = translator
			}
		}
	}

	return nil
}

// flattenTranslations converts nested JSON to flat key-value pairs
func flattenTranslations(data map[string]any, prefix string, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]any:
			flattenTranslations(v, fullKey, result)
		}
	}
}

// GetLanguages returns all loaded languages
func (m *Manager) GetLanguages() []*Language {
	m.mu.RLock()
	defer m.mu.RUnlock()

	languages := make([]*Language, 0, len(m.languages))
	for _, lang := range m.languages {
		languages = append(languages, lang)
	}
	return languages
}

// GetTranslator returns a translator for the given language
func (m *Manager) GetTranslator(lang string) *Translator {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if t, ok := m.translators[lang]; ok {
		return t
	}

	// Fall back to default language
	if t, ok := m.translators[m.defaultLang]; ok {
		return t
	}

	// Return empty translator if nothing found
	return &Translator{
		lang:         lang,
		translations: make(map[string]string),
	}
}

// T is a shortcut for translating a key using the default manager
func T(lang, key string, args ...any) string {
	return Default().GetTranslator(lang).T(key, args...)
}

// T translates a key, with optional arguments for formatting
func (t *Translator) T(key string, args ...any) string {
	if t == nil {
		return key
	}

	value, ok := t.translations[key]
	if !ok && t.fallback != nil {
		return t.fallback.T(key, args...)
	}
	if !ok {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(value, args...)
	}
	return value
}

// Has checks if a translation key exists
func (t *Translator) Has(key string) bool {
	if t == nil {
		return false
	}
	_, ok := t.translations[key]
	return ok
}

// Lang returns the language code of this translator
func (t *Translator) Lang() string {
	if t == nil {
		return ""
	}
	return t.lang
}

// All returns all translations for this translator
func (t *Translator) All() map[string]string {
	if t == nil {
		return make(map[string]string)
	}
	return t.translations
}

// TranslatorFunc is a function type for use in templates
type TranslatorFunc func(key string, args ...any) string

// Func returns a translator function for use in templates
func (t *Translator) Func() TranslatorFunc {
	return t.T
}

// LoadBuiltinLanguages loads the built-in English and Chinese translations
func (m *Manager) LoadBuiltinLanguages() error {
	// Ensure the base path exists
	if err := os.MkdirAll(m.basePath, 0755); err != nil {
		return fmt.Errorf("create locales directory: %w", err)
	}

	// Load embedded translations
	if err := m.loadEmbeddedLanguage("en", englishTranslations()); err != nil {
		return err
	}
	if err := m.loadEmbeddedLanguage("zh", chineseTranslations()); err != nil {
		return err
	}

	return nil
}

func (m *Manager) loadEmbeddedLanguage(lang string, translations map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Extract metadata
	if meta, ok := translations["meta"].(map[string]any); ok {
		m.languages[lang] = &Language{
			Code:       getString(meta, "code", lang),
			Name:       getString(meta, "name", lang),
			NativeName: getString(meta, "native_name", lang),
			Direction:  getString(meta, "direction", "ltr"),
			Flag:       getString(meta, "flag", ""),
		}
	} else {
		m.languages[lang] = &Language{
			Code:       lang,
			Name:       lang,
			NativeName: lang,
			Direction:  "ltr",
		}
	}

	// Flatten translations
	flatTranslations := make(map[string]string)
	for key, value := range translations {
		if key == "meta" {
			continue
		}
		if nested, ok := value.(map[string]any); ok {
			flattenTranslations(nested, key, flatTranslations)
		} else if str, ok := value.(string); ok {
			flatTranslations[key] = str
		}
	}

	translator := &Translator{
		lang:         lang,
		translations: flatTranslations,
	}

	// Set fallback
	if lang != m.fallbackLang {
		if fb, ok := m.translators[m.fallbackLang]; ok {
			translator.fallback = fb
		}
	}

	m.translators[lang] = translator

	// Update fallback references
	if lang == m.fallbackLang {
		for code, t := range m.translators {
			if code != lang {
				t.fallback = translator
			}
		}
	}

	return nil
}

func getString(m map[string]any, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}
