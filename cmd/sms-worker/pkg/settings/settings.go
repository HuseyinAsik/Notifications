package settings

import (
	"log"
	"os"
	"strings"

	variables "github.com/HuseyinAsik/Notifications/pkg/settings"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var AppSettings = &variables.App{}
var DatabaseSettings = &variables.Database{}
var KafkaSettings = &variables.Kafka{}

func Setup() {
	_ = godotenv.Load()

	validate := validator.New()
	validate.SetTagName("sms_worker_validate")
	AppSettings.Encoding = os.Getenv("LOG_ENCODING")
	AppSettings.LogLevel = os.Getenv("LOG_LEVEL")
	AppSettings.AppName = os.Getenv("APP_NAME")
	AppSettings.Hostname, _ = os.Hostname()

	AppSettingsErr := validate.Struct(AppSettings)
	if AppSettingsErr != nil {
		log.Fatalf("app settings missing err: %v", AppSettingsErr)
	}
	AppSettings.Load()

	KafkaSettings.BrokersStr = os.Getenv("KAFKA_BROKERS")

	kafkaSettingsErr := validate.Struct(KafkaSettings)
	if kafkaSettingsErr != nil {
		log.Fatalf("kafka settings missing err: %v", kafkaSettingsErr)
	}

	KafkaSettings.Brokers = strings.Split(KafkaSettings.BrokersStr, ",")

	DatabaseSettings.ReadUrl = os.Getenv("POSTGRESQL_READ_DSN")
	DatabaseSettings.WriteUrl = os.Getenv("POSTGRESQL_WRITE_DSN")

	databaseSettingsErr := validate.Struct(DatabaseSettings)
	if databaseSettingsErr != nil {
		log.Fatalf("database settings missing err: %v", databaseSettingsErr)
	}
}
