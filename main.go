package main

import (
	"fmt"
	"log"
	"merchant/contrib/apollo"
	"merchant/contrib/conn"
	"merchant/contrib/session"
	"merchant/middleware"
	"merchant/model"
	"merchant/router"
	"os"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/beanstalkd/go-beanstalk"
	"github.com/valyala/fasthttp"
	_ "go.uber.org/automaxprocs"
)

var (
	gitReversion   = ""
	buildTime      = ""
	buildGoVersion = ""
)

func main() {

	var err error

	argc := len(os.Args)
	if argc != 4 {
		fmt.Printf("%s <etcds> <cfgPath> <web|load>\r\n", os.Args[0])
		return
	}

	cfg := conf{}

	endpoints := strings.Split(os.Args[1], ",")
	mt := new(model.MetaTable)
	apollo.New(endpoints)
	apollo.Parse(os.Args[2], &cfg)
	mt.PromoteConfig, _ = apollo.ParseToml("/promote.toml", true)
	apollo.Close()

	mt.Lang = cfg.Lang
	mt.Prefix = cfg.Prefix
	mt.EsPrefix = cfg.EsPrefix
	mt.PullPrefix = cfg.PullPrefix
	mt.AutoCommission = cfg.AutoCommission

	mt.MerchantTD = conn.InitTD(cfg.Td.Addr, cfg.Td.MaxIdleConn, cfg.Td.MaxOpenConn)
	mt.MerchantDB = conn.InitDB(cfg.Db.Master.Addr, cfg.Db.Master.MaxIdleConn, cfg.Db.Master.MaxOpenConn)
	mt.ReportDB = conn.InitDB(cfg.Db.Report.Addr, cfg.Db.Report.MaxIdleConn, cfg.Db.Report.MaxOpenConn)
	mt.BetDB = conn.InitDB(cfg.Db.Bet.Addr, cfg.Db.Bet.MaxIdleConn, cfg.Db.Bet.MaxOpenConn)
	mt.MerchantRedis = conn.InitRedisCluster(cfg.Redis.Addr, cfg.Redis.Password)

	mt.ES = conn.InitES(cfg.Es.Host, cfg.Es.Username, cfg.Es.Password)
	mt.AccessEs = conn.InitES(cfg.AccessEs.Host, cfg.AccessEs.Username, cfg.AccessEs.Password)

	bin := strings.Split(os.Args[0], "/")
	mt.Program = bin[len(bin)-1]
	mt.GcsDoamin = cfg.GcsDoamin

	mt.MerchantBean, err = beanstalk.Dial("tcp", cfg.Beanstalkd.Addr)
	if err != nil {
		fmt.Printf("beanstalk error: %s", err.Error())
		os.Exit(1)
	}

	mt.MerchantMQ, err = rocketmq.NewProducer(
		producer.WithNameServer(cfg.Rocketmq),
		producer.WithRetry(2),
		producer.WithGroupName("merchant"),
	)
	if err != nil {
		fmt.Printf("start NewProducer error: %s", err.Error())
		os.Exit(1)
	}
	err = mt.MerchantMQ.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}

	model.Constructor(mt, cfg.RPC)
	session.New(mt.MerchantRedis, cfg.Prefix)

	if os.Args[3] == "load" {
		model.Load()
		return
	}

	defer func() {
		model.Close()
		mt = nil
	}()

	b := router.BuildInfo{
		GitReversion:   gitReversion,
		BuildTime:      buildTime,
		BuildGoVersion: buildGoVersion,
	}
	app := router.SetupRouter(b)
	srv := &fasthttp.Server{
		Handler:            middleware.Use(app.Handler),
		ReadTimeout:        router.ApiTimeout,
		WriteTimeout:       router.ApiTimeout,
		Name:               "merchant2",
		MaxRequestBodySize: 51 * 1024 * 1024,
	}
	//fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)
	//fmt.Println("Merchant2 running", cfg.Port.Merchant)

	service := model.NewService(gitReversion, buildTime, buildGoVersion, 1)
	go service.Start()

	if err := srv.ListenAndServe(cfg.Port.Merchant); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
