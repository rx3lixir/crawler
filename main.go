package main

import (
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	telegramToken := "7430205004:AAGqv2y68ISkfEFQMeN_foCIVgmhdCKbAG8"
	telegram.StartBot(telegramToken)
}
