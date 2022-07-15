package emails

import (
	"net/smtp"
	"regexp"
)

var (
	from        = "luca.sarubbi1@gmail.com"
	appPassword = "nlollervymiauiiw"

	host    = "smtp.gmail.com"
	address = host + ":587"

	emailRegex = regexp.MustCompile(`^(?:[a-z]|[0-9]|\.[^\.])+@[a-z]+.[a-z]+$`)
)

func SendEmail(to []string, subject string, body string) error {
	return smtp.SendMail(address, smtp.PlainAuth("MineOs", from, appPassword, host), from, to, []byte(subject+body))
}

func IsValidEmail(email string) bool {
	return emailRegex.Match([]byte(email))
}

func AreValidEmails(emails []string) bool {
	for _, email := range emails {
		if !IsValidEmail(email) {
			return false
		}
	}
	return true
}
