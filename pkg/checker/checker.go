package checker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"web-health-check/pkg/config"
	"web-health-check/pkg/notification"
)

const (
	defaultStatusCode      = 200
	notificationFilePrefix = "notification-"
	defaultContentType     = "application/json"
	failedMessageTemplate  = "[DOWN] CHECK (NAME=%v URL=%v) WRONG '%v' GOT '%v' EXPECTED '%v'"
	okMessageTemplate      = "[UP] CHECK (NAME=%v URL=%v)"
)

func init() {
	cleanNotificationFiles()
}

// delete notification files which have no coresponding web check
func cleanNotificationFiles() {
	conf := config.Conf.Get()
	files, err := filepath.Glob(conf.DataDir + "/" + notificationFilePrefix + "*")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		return
	}

	check_files := make(map[string]string)
	for _, check := range conf.Checks {
		check_files[filepath.Base(GetNotificationFilePath(&check))] = ""
	}

	for _, file := range files {
		if _, ok := check_files[filepath.Base(file)]; !ok {
			err_rm := os.Remove(file)
			if err_rm != nil {
				log.Println(err_rm)
			}
		}
	}
}

func GetNotificationFilePath(check *config.Check) string {
	conf := config.Conf.Get()
	return fmt.Sprintf("%v/%v%v", conf.DataDir, notificationFilePrefix, AsSha256(check))
}

func notifyOrNot(isFailed bool, failedMessage string, check config.Check) {
	conf := config.Conf.Get()
	nFile := GetNotificationFilePath(&check)
	if isFailed {
		log.Printf(failedMessage)
		nFileStat, err_stat := os.Stat(nFile)
		if os.IsNotExist(err_stat) {
			notification.Notify(failedMessage, &check)
			f, err_create := os.Create(nFile)
			if err_create != nil {
				log.Println(err_create)
			} else {
				_, err_write := fmt.Fprintf(f, "%+v", check)
				if err_write != nil {
					log.Println(err_write)
				}
			}
			defer f.Close()
		} else {
			currentTime := time.Now().Local()

			notificationInterval := conf.NotificationInterval
			if check.NotificationInterval != 0 {
				notificationInterval = check.NotificationInterval
			}

			if uint(currentTime.Unix()-nFileStat.ModTime().Unix()) > notificationInterval {
				notification.Notify(failedMessage, &check)
				err_touch := os.Chtimes(nFile, currentTime, currentTime)
				if err_touch != nil {
					log.Println(err_touch)
				}
			}
		}
	} else {
		if _, err := os.Stat(nFile); err == nil {
			okMessage := fmt.Sprintf(okMessageTemplate, check.Name, check.Url)
			log.Printf(okMessage)
			err_rm := os.Remove(nFile)
			if err_rm != nil {
				log.Println(err_rm)
			}
			notification.Notify(okMessage, &check)
		}
	}
}

func DoCheck(check config.Check, conf config.Config) {
	// fmt.Println(check.Name, GetNotificationFilePath(&check))
	log.Println("Started ", check)
	var statusCode int
	var isFailed bool
	var failedMessage string

	client := http.Client{
		Timeout: time.Duration(conf.Timeout) * time.Second,
	}

	for i := 1; i <= conf.FailureThreshold; i++ {
		log.Printf("Starting attempt %v %v", i, check)

		isFailed = false
		resp, err := client.Get(check.Url)

		if err != nil {
			failedMessage = fmt.Sprintf("FAILED: CHECK (NAME=%v URL=%v) %v", check.Name, check.Url, err)
			isFailed = true
		}

		// check status code (always even it's not set in config)
		if !isFailed {
			statusCode = defaultStatusCode
			if check.StatusCode != 0 {
				statusCode = check.StatusCode
			}

			if resp.StatusCode != statusCode {
				failedMessage = fmt.Sprintf(failedMessageTemplate, check.Name, check.Url,
					"status code", resp.StatusCode, statusCode)
				isFailed = true
			}
		}

		// check content-type
		if !isFailed && check.ContentType != "" {
			if !strings.Contains(resp.Header.Get("Content-Type"), check.ContentType) {
				failedMessage = fmt.Sprintf(failedMessageTemplate, check.Name, check.Url,
					"content-type", resp.Header.Get("Content-Type"), check.ContentType)
				isFailed = true
			}
		}

		// check response
		if !isFailed && check.Response != "" {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				failedMessage = fmt.Sprintf("CHECK (NAME=%v URL=%v) FAILED: %v", check.Name, check.Url, err)
				isFailed = true
			} else {
				// TODO: improve response comparison
				body_str := strings.TrimSpace(string(body))
				if body_str != check.Response {
					failedMessage = fmt.Sprintf(failedMessageTemplate, check.Name, check.Url,
						"response", body_str, check.Response)
					isFailed = true
				}
			}
		}

		if resp != nil {
			resp.Body.Close()
		}

		if isFailed {
			log.Printf("Failed attempt %v %v, sleep %v", i, check, conf.FailureInterval)
			if i < conf.FailureThreshold {
				time.Sleep(time.Duration(conf.FailureInterval) * time.Second)
			}
		} else {
			log.Printf("OK attempt %v %v", i, check)
			break
		}

	}

	notifyOrNot(isFailed, failedMessage, check)
	log.Println("Finished ", check)
}
