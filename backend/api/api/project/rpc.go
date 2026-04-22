package project

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"test.com/project-api/config"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	"test.com/project-grpc/project"
)

var ProjectServiceClient project.ProjectServiceClient

func InitRpcProjectClient() {
	target := "etcd:///project"
	if config.C.RpcConfig != nil && config.C.RpcConfig.ProjectAddr != "" {
		target = config.C.RpcConfig.ProjectAddr
	} else {
		etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
		resolver.Register(etcdRegister)
	}

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	ProjectServiceClient = project.NewProjectServiceClient(conn)
}
