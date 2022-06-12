package notification

import (
	"log"

	"web-health-check/pkg/config"
)

func Notify(message string, check *config.Check) {
	conf := config.Conf.Get()
	notifications := conf.Notifications
	if len(check.Notifications) > 0 {
		notifications = check.Notifications
	}

	for _, notification := range notifications {
		if method, ok := conf.NotificationMethods[notification]; ok {
			switch method["type"] {
			case "slack":
				slack(message, method)
			case "email":
				email()
			default:
				log.Printf("Not implemented")
			}
		}
	}
}
