package notifier

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigName("notifier-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/notifier/")
	viper.AddConfigPath("$HOME/.config/notifier")
	viper.AddConfigPath(".")

	viper.SetDefault("http.addr", ":8080")
	viper.SetDefault("general.date_format", "2006-01-02 15:04:05")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error while reading config file: %v", err)
	}

	tgManager := NewTelegramManager()
	sinks, err := sinksFromConfig(tgManager)
	if err != nil {
		log.Fatalf("Fatal error in config file: %v", err)
	}
	users, err := usersFromConfig()
	if err != nil {
		log.Fatalf("Fatal error in config file: %v", err)
	}
	if viper.GetString("http.jwt_secret") == "" {
		log.Fatalf("Fatal error in config file: jwt_secret is not defined")
	}
	hs := NewHttpServer(sinks, users)
	hs.Start(viper.GetString("http.addr"))
}

func sinksFromConfig(tgManager *TelegramManager) ([]NotificationSink, error) {
	sinks := []NotificationSink{}
	sinksRaw := viper.Get("sinks")
	if sinksRaw == nil {
		return nil, fmt.Errorf("no sinks defined in config file")
	}

	if sinksRaw, ok := sinksRaw.([]interface{}); ok {
		for i, sink := range sinksRaw {
			if sink, ok := sink.(map[interface{}]interface{}); ok {
				switch sink["type"] {
				case "telegram":
					s := &TelegramNotificationSink{
						TelegramManager: tgManager,
						BotToken:        sink["bot_token"].(string),
					}
					fmt.Sscanf(fmt.Sprintf("%v", sink["chat_id"]), "%v", &s.ChatID)
					if err := s.Init(); err != nil {
						return nil, fmt.Errorf("error initializing telegram sink #%v: %v", i, err)
					}
					sinks = append(sinks, s)
				case "email":
					to := []string{}
					if sinkTo, ok := sink["to"].([]interface{}); ok {
						for _, t := range sinkTo {
							to = append(to, t.(string))
						}
					} else {
						to = append(to, sink["to"].(string))
					}
					var startTLS bool
					if sink["starttls"] != nil {
						startTLS = sink["starttls"].(bool)
					}
					s := &EmailNotificationSink{
						From:         sink["from"].(string),
						To:           to,
						SMTPAddress:  sink["smtp_address"].(string),
						SMTPUsername: sink["smtp_username"].(string),
						SMTPPassword: sink["smtp_password"].(string),
						StartTLS:     startTLS,
					}
					if err := s.Init(); err != nil {
						return nil, fmt.Errorf("error initializing email sink #%v: %v", i, err)
					}
					sinks = append(sinks, s)
				default:
					return nil, fmt.Errorf("unknown sink type: %s", sink["type"])
				}
			}
		}
	} else {
		return nil, fmt.Errorf("sinks should be an array in config file")
	}
	return sinks, nil
}

func usersFromConfig() ([]*User, error) {
	usersRaw := viper.Get("users")
	if usersRaw == nil {
		return nil, fmt.Errorf("no users defined in config file")
	}
	users := []*User{}
	if usersRaw, ok := usersRaw.([]interface{}); ok {
		for _, user := range usersRaw {
			if user, ok := user.(map[interface{}]interface{}); ok {
				m, err := json.Marshal(fiber.Map{
					"username": user["username"],
					"password": user["password"],
					"token":    user["token"],
				})
				if err != nil {
					return nil, fmt.Errorf("error marshaling user: %v", err)
				}
				u := &User{}
				err = json.Unmarshal(m, u)
				if err != nil {
					return nil, fmt.Errorf("error unmarshaling user: %v", err)
				}
				users = append(users, u)
			}
		}
	} else {
		return nil, fmt.Errorf("users should be an array in config file")
	}

	return users, nil
}
