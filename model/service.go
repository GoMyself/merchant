package model

import (
	"fmt"
	"log"
	"merchant/contrib/helper"
	"net"
	"os"
	"time"

	g "github.com/doug-martin/goqu/v9"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var flags = map[int]string{
	1: "api",
	2: "rpc",
	3: "script",
}

type BuildInfo struct {
	id             int64
	Name           string
	GitReversion   string
	BuildTime      string
	BuildGoVersion string
	Flag           int
	IP             string
	Hostname       string
}

func telegramBotNotice(program, gitReversion, buildTime, buildGoVersion, hostname, localIp string, flag int) {

	ts := time.Now()
	bot, err := tgbotapi.NewBotAPI("5249320515:AAHibqLVtW69J6_OyJi1amDwXO1HfVTr3iw")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	str := "\U00002733	\U00002733	\U00002733\r\ndatetime: \t%s\r\nprogram: \t%s\r\nGitReversion: \t%s\r\nBuildTime: \t%s\r\nBuildGoVersion: \t%s\r\nhostname: \t%s\r\nIP: \t%s\r\nflag: \t%s\U00002728 \U00002728 \U00002728\r\n"

	msg := tgbotapi.NewMessage(-738052985, "")
	msg.Text = fmt.Sprintf(str, ts.Format("2006-01-02 15:04:05"), program, gitReversion, buildTime, buildGoVersion, hostname, localIp, flags[flag])
	if _, err := bot.Send(msg); err != nil {
		_ = pushLog(err, helper.ServerErr)
	}
}

func NewService(gitReversion, buildTime, buildGoVersion string, flag int) BuildInfo {

	ts := time.Now().UnixMicro()
	b := BuildInfo{
		id:             ts,
		Flag:           flag,
		Name:           meta.Program,
		GitReversion:   gitReversion,
		BuildTime:      buildTime,
		BuildGoVersion: buildGoVersion,
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		return b
	}

	b.Hostname = hostname
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
		return b
	}

	var localIp string
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIp = ipnet.IP.String()
				break
			}
		}
	}
	b.IP = localIp

	// 正式环境服务启动，小飞机推送消息
	if !meta.IsDev {
		telegramBotNotice(meta.Program, gitReversion, buildTime, buildGoVersion, hostname, localIp, flag)
	}

	return b
}

func (s BuildInfo) keepAlive() error {

	now := time.Now()
	recs := g.Record{
		"ts":             s.id,
		"name":           s.Name,
		"flag":           s.Flag,
		"ip":             s.IP,
		"hostname":       s.Hostname,
		"buildTime":      s.BuildTime,
		"gitReversion":   s.GitReversion,
		"buildGoVersion": s.BuildGoVersion,
		"created_at":     now.Unix(),
		"prefix":         meta.Prefix,
	}

	query, _, _ := dialect.Insert("services").Rows(recs).ToSQL()
	_, err := meta.MerchantTD.Exec(query)

	if err != nil {
		fmt.Println("insert service failed query ", query)
		fmt.Println("insert service failed error ", err.Error())
		return err
	}

	return nil
}

func (s BuildInfo) Start() error {

	s.keepAlive()

	for {

		time.Sleep(10 * time.Second)
		s.keepAlive()
	}
}
