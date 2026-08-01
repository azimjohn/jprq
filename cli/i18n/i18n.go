package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed locales/*.json
var localeFS embed.FS

var translations map[string]string
var currentLang = "en" // Default language

// Init Language
func Init(lang string) {
	if lang == "" {
		lang = "en"
	}
	currentLang = lang

	data, err := localeFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
	if err != nil {
		// Agar til topilmasa, ingliz tiliga (fallback) o'tadi
		data, _ = localeFS.ReadFile("locales/en.json")
	}

	json.Unmarshal(data, &translations)
}

// T funksiyasi kalit bo'yicha tarjimani qaytaradi
func T(key string, args ...interface{}) string {
	msg, ok := translations[key]
	if !ok {
		return key // Kalit topilmasa, kalitning o'zi qaytadi
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}