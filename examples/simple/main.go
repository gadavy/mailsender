package main

import (
	"log"

	"github.com/gadavy/mailsender"
)

func main() {
	// Создаем клиент
	emailClient, err := mailsender.NewSMTPClient().
		Host("smtp.example.ru:465").
		Login("example@mail.com").
		Password("password").
		SSL(true).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Отправляем текст
	err = emailClient.Send(mailsender.New().Email().
		From("from@mail.com").
		To("to@mail.com").
		Subject("subject").
		Text("email text").
		Build())
	if err != nil {
		log.Fatal(err)
	}
}
