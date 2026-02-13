package settings

import (
	"log"
	"os"

	variables "github.com/HuseyinAsik/Notifications/pkg/settings"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var AppSettings = &variables.App{}
var ServerSettings = &variables.Server{}
var DatabaseSettings = &variables.Database{}

func Setup() {
	_ = godotenv.Load()

	validate := validator.New()
	validate.SetTagName("notification_api_validate")
	AppSettings.Encoding = os.Getenv("LOG_ENCODING")
	AppSettings.LogLevel = os.Getenv("LOG_LEVEL")
	AppSettings.AppName = os.Getenv("APP_NAME")
	AppSettings.RunMode = os.Getenv("RUN_MODE")
	AppSettings.ContextTimeout = os.Getenv("APP_CONTEXT_TIMEOUT")
	AppSettings.Hostname, _ = os.Hostname()

	AppSettingsErr := validate.Struct(AppSettings)
	if AppSettingsErr != nil {
		log.Fatalf("app settings missing err: %v", AppSettingsErr)
	}
	AppSettings.Load()

	ServerSettings.HttpPort = os.Getenv("HTTP_PORT")

	serverSettingsErr := validate.Struct(ServerSettings)
	if serverSettingsErr != nil {
		log.Fatalf("server settings missing err: %v", serverSettingsErr)
	}

	DatabaseSettings.ReadUrl = os.Getenv("POSTGRESQL_READ_DSN")
	DatabaseSettings.WriteUrl = os.Getenv("POSTGRESQL_WRITE_DSN")

	databaseSettingsErr := validate.Struct(DatabaseSettings)
	if databaseSettingsErr != nil {
		log.Fatalf("database settings missing err: %v", databaseSettingsErr)
	}
}
