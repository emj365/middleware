package i18n

import (
	"strings"

	"github.com/kataras/iris/context"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// I18n the struct
type I18n struct {
	// Default set it if you want a default language
	//
	// Checked: Configuration state, not at runtime
	Default string
	// URLParameter is the name of the url parameter which the language can be indentified
	//
	// Checked: Serving state, runtime
	URLParameter string
	// Bundle of i18n
	//
	// Checked: Configuration state, not at runtime
	Bundle *i18n.Bundle
}

//New returns a new i18n middleware and load locale files by given args
func New(locales ...string) *I18n {
	b := i18n.NewBundle(language.English)
	for _, loc := range locales {
		b.MustLoadMessageFile(loc)
	}

	return &I18n{
		Default:      "en-US",
		URLParameter: "lang",
		Bundle:       b,
	}
}

// Serve implemented iris handler
func (i *I18n) Serve(ctx context.Context) {
	wasByCookie := false

	language := i.Default

	langKey := ctx.Application().ConfigurationReadOnly().GetTranslateLanguageContextKey()
	if ctx.Values().GetString(langKey) == "" {
		// try to get by url parameter
		language = ctx.URLParam(i.URLParameter)
		if language == "" {
			// then try to take the lang field from the cookie
			language = ctx.GetCookie(langKey)

			if len(language) > 0 {
				wasByCookie = true
			} else {
				// try to get by the request headers.
				langHeader := ctx.GetHeader("Accept-Language")
				if len(langHeader) > 0 {
					for _, langEntry := range strings.Split(langHeader, ",") {
						lc := strings.Split(langEntry, ";")[0]
						for _, tag := range i.Bundle.LanguageTags() {
							code := strings.Split(lc, "-")[0]
							if strings.Contains(tag.String(), code) {
								language = lc
								break
							}
						}
					}
				}
			}
		}
		// if it was not taken by the cookie, then set the cookie in order to have it
		if !wasByCookie {
			ctx.SetCookieKV(langKey, language)
		}
	}

	localizer := i18n.NewLocalizer(i.Bundle, language)

	ctx.Values().Set(langKey, language)
	translateFuncKey := ctx.Application().ConfigurationReadOnly().GetTranslateFunctionContextKey()
	//wrap tr to raw func for ctx.Translate usage
	ctx.Values().Set(translateFuncKey, func(translationID string, args ...interface{}) string {
		desc := ""
		if len(args) > 0 {
			if v, ok := args[0].(string); ok {
				desc = v
			}
		}

		translated, err := localizer.LocalizeMessage(&i18n.Message{ID: translationID, Description: desc})
		if err != nil {
			return err.Error()
		}
		return translated
	})

	ctx.Next()
}
