package main

import (
	"apiring/configs"
	"apiring/database"
	"apiring/api"
	"net/http"
)


func main() {
	// 初始化配置
	configs.InitConfig()

	// 关闭数据库驱动
	defer configs.Neo4j.Close()

	// 初始化数据库
	database.InitNeo4j(configs.Neo4j, configs.Logger, configs.DefaultPasswd)

	// 初始化默认资源及其权限 --> neo4j db
	database.GenPermsDefault(configs.Neo4j, configs.Logger)

	// 启动http服务
	configs.Logger.Infof("监听服务: %s", configs.HTTP.Addr)
	http.ListenAndServe(configs.HTTP.Addr, api.MainRouter())
}
