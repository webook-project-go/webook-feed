package main

import (
	"context"
	v1 "github.com/webook-project-go/webook-apis/gen/go/apis/feed/v1"
	_ "github.com/webook-project-go/webook-feed/config"
	"github.com/webook-project-go/webook-feed/ioc"
)

func main() {
	app := InitApp()
	for _, c := range app.Consumer {
		c.Start()
	}
	shutdwon := ioc.InitOTEL()
	defer shutdwon(context.Background())
	v1.RegisterFeedServiceServer(app.Server, app.Service)
	err := app.Server.Serve()
	if err != nil {
		panic(err)
	}
}
