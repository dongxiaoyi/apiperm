package auth

import (
	"apiring/configs"
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/vektah/gqlparser/gqlerror"
	"runtime"
	"strings"
)

// 获取jwt token中的username
func forContext(ctx context.Context) string {
	return ctx.Value("user_name").(string)
}


// 获取resolver或者views的资源名称，查库验证权限
func ResourcePermVerify(ctx context.Context) *gqlerror.Error {
	/*
	resolver.go使用示例：
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Datas{}, e
	}
	*/
	isPerm := true

	driver := configs.Neo4j
	logger := configs.Logger
	userName := forContext(ctx)
	pc, _, _, _ := runtime.Caller(1)
	resource := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	appData := resource[len(resource) - 1]
	resourceStyle := strings.Split(appData, ".")
	// 资源属于哪个app
	service := resourceStyle[0]
	// 资源的名称data
	data := resourceStyle[len(resourceStyle) - 1]
	// TODO: 查询用户的角色对资源的权限

	var session neo4j.Session
	var err error
	if session, err = driver.Session(neo4j.AccessModeWrite); err != nil {
		logger.Error(err)
	}
	defer session.Close()
	// 查询用户所属角色(只要登录, roles肯定存在)
	roles, err := session.Run(QueryUserRole(userName), map[string]interface{}{})
	if err != nil {
		logger.Error(err)
	}

	var belongRoles []string
	for roles.Next() {
		roleName, _ := roles.Record().Get("roleName")
		belongRoles = append(belongRoles, roleName.(string))
	}

	// 查询role与data的权限
	if len(belongRoles) > 0 {
		sqlBelongIn := "["
		for i, r := range belongRoles {
			sqlBelongIn += "\"" + r + "\""
			if i > 0 && i < len(belongRoles) - 1 {
				sqlBelongIn += ","
			}
		}
		sqlBelongIn += "]"

		opCountResult, err := session.Run(QueryRoleDataRelationCount(sqlBelongIn, data, service), map[string]interface{}{})
		if err != nil {
			logger.Error(err)
		}
		if opCountResult == nil {
			isPerm = false
		}
		for opCountResult.Next() {
			c, _ := opCountResult.Record().Get("opCount")
			opCount := c.(int64)
			if opCount <= 0 {
				isPerm = false
			}
		}
	} else {
		isPerm =  false
	}
	//fmt.Println(userName, data, isPerm)
	if !isPerm {
		return gqlerror.Errorf("Permission denied!")
	}
	return nil
}

// 查询用户所属角色
func QueryUserRole(userName string) string {
	sql := fmt.Sprintf(`
MATCH (u:USER)-[b:BELONG_TO]->(r:ROLE)
WHERE u.name = "%s"
RETURN r.name as roleName
`, userName)
	return sql
}

// 查询角色与资源的权限关系条数
func QueryRoleDataRelationCount(roleList, dataName, service string) string {
	sql := fmt.Sprintf(`
MATCH (r:ROLE)-[o:OPERATE]->(d:DATA)
WHERE r.name IN %s AND d.name = "%s" AND d.service = "%s"
RETURN COUNT(*) as opCount
`, roleList, dataName, service)
	return sql
}