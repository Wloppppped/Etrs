package main

import (
	client "Etrs/etrs_cli"
	server "Etrs/etrs_ser"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func main() {
	c := zap.NewDevelopmentConfig()
	l, err := c.Build()
	if err != nil {
		return
	}
	zap.ReplaceGlobals(l)

	go ser()
	time.Sleep(1 * time.Second)
	go cli()
	time.Sleep(60 * 60 * time.Second)
}

func ser() {
	etrsRegistry := server.ServcerInit()
	err := etrsRegistry.ServerRegistry("/ser", "localhost1")
	etrsRegistry.ServerRegistry("/ser", "localhost2")
	etrsRegistry.ServerRegistry("/ser", "localhost3")
	if err != nil {
		zap.L().Error(err.Error())
	}
	time.Sleep(10 * time.Second)
	err2 := etrsRegistry.ServerRegistry("/ser", "127.0.0.1:80")
	if err2 != nil {
		zap.L().Error(err2.Error())
	}
	time.Sleep(10 * time.Second)
	etrsRegistry.ServerCancle()
	// time.Sleep(60 * 60 * time.Second)
}

func cli() {
	k := "/ser"
	etrsResolver := client.ClientInit()
	s, err := etrsResolver.GetServicePrefix(k)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	zap.L().Debug(fmt.Sprintf("FIRST resolver key: %s, value: %s", k, s))
	for {
		time.Sleep(2 * time.Second)
		s, b := etrsResolver.Resolver(k)
		if b {
			zap.L().Debug(fmt.Sprintf("resolver key: %s, value: %s", k, s))
		} else {
			zap.L().Debug(fmt.Sprintf("resolver key: %s, value: %s", k, "[not found]"))
		}
	}
}
