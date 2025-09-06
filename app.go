package main

import (
	"github.com/webook-project-go/webook-feed/events"
	"github.com/webook-project-go/webook-feed/grpc"
	"github.com/webook-project-go/webook-pkgs/grpcx"
)

type App struct {
	Service  *grpc.Service
	Server   *grpcx.GrpcxServer
	Consumer []events.Consumer
}
