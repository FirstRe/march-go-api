package translation

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle
var localT *i18n.Localizer

func InitI18n() {
	dir, _ := os.Getwd()
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	var filePathEn string
	var filePathTh string
	if dir == "/app/services/march-inventory/cmd/app" {
		filePathEn = filepath.Join(dir, "i18n", "en", "en.json")
		filePathTh = filepath.Join(dir, "i18n", "th", "th.json")
	} else {
		filePathEn = filepath.Join(dir, "cmd", "app", "i18n", "en", "en.json")
		filePathTh = filepath.Join(dir, "cmd", "app", "i18n", "th", "th.json")
	}

	// Load message files
	bundle.MustLoadMessageFile(filePathEn)
	bundle.MustLoadMessageFile(filePathTh)
}

func InitLocalizer(langCode string) {
	localT = i18n.NewLocalizer(bundle, langCode)
}

func LocalizeMessage(messageID string) string {
	var translation string
	defer func() {
		if r := recover(); r != nil {
			translation = ""
		}
	}()

	translation = localT.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	return translation
}
