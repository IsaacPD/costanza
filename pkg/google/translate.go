package google

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/translate/v2"

	"github.com/isaacpd/costanza/pkg/cmd"
)

var (
	translateService *translate.Service
)

func translateHelper(q []string, target string) string {
	response, err := translateService.Translations.List(q, target).Do()
	if err != nil {
		logrus.Errorf("Error translating: %s", err)
		return "Error translating"
	}

	var builder strings.Builder
	for _, t := range response.Translations {
		fmt.Fprintf(&builder, "%s\n", t.TranslatedText)
	}
	return builder.String()
}

func isLanguage(lang string) string {
	languages, err := translateService.Languages.List().Target("en").Do()
	if err != nil {
		logrus.Errorf("Error getting supported languages: %s", err)
		return ""
	}
	for _, l := range languages.Languages {
		if l.Language == lang || strings.EqualFold(l.Name, lang) {
			return l.Language
		}
	}
	return ""
}

func Translate(c cmd.Context) (string, error) {
	var target, lang string
	var params []string

	if c.Message != nil {
		m := c.Arg
		begin := strings.Index(m, "(") + 1
		end := strings.Index(m, ")")
		params = strings.Split(m[begin:end], ";")
		target = m[strings.Index(m[end:], "to ")+end+3:]
	} else {
		target = c.Interaction.ApplicationCommandData().Options[1].StringValue()
		params = strings.Split(c.Interaction.ApplicationCommandData().Options[0].StringValue(), ";")
	}
	lang = isLanguage(target)
	if lang != "" {
		return translateHelper(params, lang), nil
	} else {
		return "", fmt.Errorf("%s is not a supported language", target)
	}
}
