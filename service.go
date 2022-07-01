package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func telegramBotNotice(program, gitReversion, buildTime, buildGoVersion, flag string) {

	var localIp string
	ts := time.Now()
	bot, err := tgbotapi.NewBotAPI("5249320515:AAHibqLVtW69J6_OyJi1amDwXO1HfVTr3iw")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	str := "\U00002733	\U00002733	\U00002733\r\ndatetime: \t%s\r\nprogram: \t%s\r\nGitReversion: \t%s\r\nBuildTime: \t%s\r\nBuildGoVersion: \t%s\r\nhostname: \t%s\r\nIP: \t%s\r\nflag: \t%s\n\U00002728 \U00002728 \U00002728\r\n"
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		return
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIp = ipnet.IP.String()
				break
			}
		}
	}

	msg := tgbotapi.NewMessage(-738052985, "")
	msg.Text = fmt.Sprintf(str, ts.Format("2006-01-02 15:04:05"), program, gitReversion, buildTime, buildGoVersion, hostname, localIp, flag)
	if _, err := bot.Send(msg); err != nil {
		fmt.Println("tgbot error : ", err.Error())
	}
}
