package google

import (
	"context"
	"os"

	"flag"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/translate/v2"
)

var (
	APIKey string
)

func init() {
	flag.StringVar(&APIKey, "key", "", "Google API Key")
}

func InitializeServices(ctx context.Context) {
	if APIKey == "" {
		APIKey = os.Getenv("GOOGLE_KEY")
	}
	var err error
	translateService, err = translate.NewService(ctx, option.WithAPIKey(APIKey))
	if err != nil {
		logrus.Errorf("Error creating translate service: %s", err)
	}
}
