package email_warning

import (
	"fmt"
	"log"
	"net/smtp"
)

// Отправка email warning в Mailtrap
func SendEmailWarning(userGUID string) {

	from := "john.doe@example.com"

	user := "9c1d45eaf7af5b"
	password := "ad62926fa75d0f"

	to := []string{
		"roger.roe@example.com",
	}

	addr := "smtp.mailtrap.io:2525"
	host := "smtp.mailtrap.io"
	msg := []byte(fmt.Sprintf("Тема: Оповещение о сменившемся IP адресе\n Предупреждение: у пользователя %s сменился IP адрес.", userGUID))

	auth := smtp.PlainAuth("", user, password, host)

	err := smtp.SendMail(addr, auth, from, to, msg)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email warning успешно отправлен")

}
