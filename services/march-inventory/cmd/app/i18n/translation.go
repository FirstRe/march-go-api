package translation

import (
	"encoding/json"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

func InitI18n() {
	dir, _ := os.Getwd()
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load message files
	bundle.MustLoadMessageFile(dir + "/cmd/app/i18n/en/en.json")
	bundle.MustLoadMessageFile(dir + "/cmd/app/i18n/th/th.json")
}

func LocalizeMessage(localizer *i18n.Localizer, messageID string) string {
	translation := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	return translation
}

func InitLocalizer(langCode string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, langCode)
}
