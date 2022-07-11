package emails

import "net/smtp"

var (
	from        = "luca.sarubbi1@gmail.com"
	appPassword = "nlollervymiauiiw"

	host    = "smtp.gmail.com"
	address = host + ":587"
)

func SendEmail(to []string, subject string, body string) error {
	return smtp.SendMail(address, smtp.PlainAuth("MineOs", from, appPassword, host), from, to, []byte(subject+body))
}
