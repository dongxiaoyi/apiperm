package database

import (
    "fmt"
    "strings"
)

/*

################################profile相关sql###############################################

*/


// 创建UNIQUE约束
func CreateUniqueData() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (data:DATA)
ASSERT data.name IS UNIQUE
`)
    return sql
}

func CreateUniqueUser() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (user:USER)
ASSERT user.name IS UNIQUE
`)
    return sql
}

func CreateUniqueRole() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (role:ROLE)
ASSERT role.name IS UNIQUE
`)
    return sql
}

// 创建USER（用户）标签
func CreateUser(name, password, description, createdtime string) string {
    // 默认用户为admin角色下admin用户；public用户下public用户
    sql := fmt.Sprintf(`
CREATE (
  user:USER
  {
    name:"%s",
    password:"%s",
    description:"%s",
    createdtime:"%s",
    updatedtime:"%s"
  }
)
`, name, password, description, createdtime, createdtime)

    return sql
}

// 更新USER（用户）标签
func SetUser(name, password, description, updatedtime string) string {
    sql := fmt.Sprintf(`
MATCH (user:USER) 
WHERE user.name="%s"
SET user.updatedtime = "%s", user.description = "%s",`, name, updatedtime, description)
    if password != "" {
        sql += fmt.Sprintf(`user.password = "%s"`, password)
    }
    sql = strings.TrimRight(sql, ",")
    sql += fmt.Sprintf(`
RETURN user.name as userName, user.description as userDescription, user.createdtime as userCreatedtime, user.updatedtime as userUpdatedtime`)
    return sql
}

// 创建ROLE（角色）标签
func CreateRole(name, description, createdtime string) string {
    sql := fmt.Sprintf(`
CREATE (
  role:ROLE
  {
    name:"%s",
    description:"%s",
    createdtime:"%s",
    updatedtime:"%s"
  }
)
RETURN role.name as roleName, role.description as roleDescription, role.createdtime as roleCreatedtime, role.updatedtime as roleUpdatedtime
`, name, description, createdtime, createdtime)
    return sql
}

// 更新ROLE（角色）标签
func SetRole(name, description, updatedtime string) string {
    sql := fmt.Sprintf(`
MATCH (role:ROLE) 
WHERE role.name="%s"
SET role.description = "%s", role.updatedtime = "%s"
return role.name as roleName, role.description as roleDescription, role.createdtime as roleCreatedtime, role.updatedtime as roleUpdatedtime
`, name, description, updatedtime)
    return sql
}

func RelationUserRole(userName, roleName, createdtime string) string {
    // 用户 - 属于 -> 角色
    sql := fmt.Sprintf(`
MATCH (user:USER), (role:ROLE)
WHERE user.name = "%s" AND role.name = "%s"
CREATE (user)-[r:BELONG_TO{createdtime:"%s"}]->(role)
RETURN user, r, role
`, userName, roleName, createdtime)
    return sql
}

// 查询用户属于哪些角色
func QueryUserBelongRole(userName string) string {
    sql := fmt.Sprintf(`
MATCH (user:USER)-[r:BELONG_TO]->(role:ROLE)
WHERE user.name = "%s"
RETURN role.name as roleName, role.description as roleDescription, role.createdtime as roleCreatedtime, role.updatedtime as roleUpdatedtime
`, userName)
    return sql
}

// 查询角色拥有哪些资源的操作权限
func QueryRoleoperateData(roleName string) string {
    sql := fmt.Sprintf(`
MATCH (role:ROLE)-[r:OPERATE]->(data:DATA)
WHERE role.name = "%s"
RETURN data.name as dataName, data.description as dataDescription, data.mode as dataMode, data.style as dataStyle, data.service as dataService, data.createdtime as dataCreatedtime, data.updatedtime as dataUpdatedtime
`, roleName)
    return sql
}

// 删除用户
func DeleteUser(userName string) string {
    // 删除 `用户` 节点
    sql := fmt.Sprintf(`
MATCH (user:USER)
WHERE user.name = "%s"
DELETE user
`, userName)
    return sql
}

// 删除角色
func DeleteRole(roleName string) string {
    // 删除 `角色` 节点
    sql := fmt.Sprintf(`
MATCH (role:ROLE)
WHERE role.name = "%s"
DELETE role
`, roleName)
    return sql
}

// 创建资源
func CreateData(name, description, mode, service, style, cratedtime string) string {
    // 资源的mode分为：w、r,分别代表写、读
    // 资源的style分为：rest、graph，分别代表restful、graphql标准的
    sql := fmt.Sprintf(`
CREATE (data:DATA
{
  name:"%s",
  description: "%s",
  mode: "%s",
  service: "%s",
  style: "%s",
  createdtime:"%s",
  updatedtime:"%s"
})
`, name, description, mode, service, style, cratedtime, cratedtime)
    return sql
}

// 更新资源
func SetData(name, description, mode, service, style, updatedtime string) string {
    sql := fmt.Sprintf(`
MATCH (data:Data) 
WHERE data.name="%s"
SET data.description = "%s", data.mode = "%s", data.service = "%s", data.style = "%s", data.updatedtime = "%s"
return role
`, name, description, mode, service, style, updatedtime)
    return sql
}

// 用户查询
func QueryUser(userName string) string {
    var sql string
    if userName == "" {
        sql = fmt.Sprintf(`
MATCH (user:USER)
RETURN user.name as userName, user.description as userDescription, user.createdtime as userCreatedtime, user.updatedtime as userUpdatedtime
`)
    } else {
        sql = fmt.Sprintf(`
MATCH (user:USER)
WHERE user.name = "%s"
RETURN user.name as userName, user.description as userDescription, user.createdtime as userCreatedtime, user.updatedtime as userUpdatedtime
`, userName)
    }
    return sql
}

// 用户(密码)查询
func QueryUserPasswd(userName, passwd string) string {
    sql := fmt.Sprintf(`
MATCH (user:USER)
WHERE user.name = "%s" AND user.password = "%s"
RETURN user
`, userName, passwd)
    return sql
}

// 角色查询
func QueryRole(roleName string) string {
    var sql string
    if roleName == "" {
        sql = fmt.Sprintf(`
MATCH (role:ROLE)
RETURN role.name as roleName, role.description as roleDescription, role.createdtime as roleCreatedtime, role.updatedtime as roleUpdatedtime
`)
    } else {
        sql = fmt.Sprintf(`
MATCH (role:ROLE)
WHERE role.name = "%s"
RETURN role.name as roleName, role.description as roleDescription, role.createdtime as roleCreatedtime, role.updatedtime as roleUpdatedtime
`, roleName)
    }

    return sql
}

// 资源查询
func QueryData(dataName, dataMode, dataStyle, dataService string) string {
    var sql string
    if dataName == "" && dataMode == "" && dataStyle == "" && dataService == "" {
        sql = fmt.Sprintf(`
MATCH (data:DATA)
RETURN data.name as dataName, data.description as dataDescription, data.mode as dataMode, data.style as dataStyle, data.service as dataService, data.createdtime as dataCreatedtime, data.updatedtime as dataUpdatedtime
`)
    } else {
        sql = fmt.Sprintf(`
MATCH (data:DATA)
WHERE `)
        if dataName != "" {
            sql += fmt.Sprintf(`data.name = "%s",`, dataName)
        }
        if dataMode != "" {
            sql += fmt.Sprintf(`data.mode = "%s",`, dataMode)
        }
        if dataStyle != "" {
            sql += fmt.Sprintf(`data.style = "%s",`, dataStyle)
        }
        if dataService != "" {
            sql += fmt.Sprintf(`data.service = "%s",`, dataService)
        }
        sql = strings.TrimRight(sql, ",")
        sql += fmt.Sprintf(`
RETURN data.name as dataName, data.description as dataDescription, data.mode as dataMode, data.style as dataStyle, data.service as dataService, data.createdtime as dataCreatedtime, data.updatedtime as dataUpdatedtime
`)
    }
    return sql
}

// 角色与用户所属关系条数
func QueryCountRelationRoleUser(roleName, userName string) string {
    sql := fmt.Sprintf(`
MATCH (n:USER{name:"%s"}),(n1:ROLE{name:"%s"}),p=(n)-[r:BELONG_TO]->(n1)
return count(p) as RoleUserCount
`, userName, roleName)
    return sql
}

// 角色与资源所属关系条数
func QueryCountRelationRoleData(roleName, dataName string) string {

    sql := fmt.Sprintf(`
MATCH (n:ROLE{name:"%s"}),(n1:DATA{name:"%s"}),p=(n)-[r:OPERATE]->(n1)
return count(p) as RoleDataCount
`, roleName, dataName)
    return sql
}

// 创建角色和资源关系
func MergeRelationRoleData(roleName, dataName string) string {
    sql := fmt.Sprintf(`
MERGE (data:DATA {name:"%s"})
MERGE (role:ROLE {name:"%s"})
MERGE (role)-[o:OPERATE]->(data)
`, dataName, roleName)
    return sql
}


// 创建资源和角色的关系
func CreateRoleDataRelation(roleName, dataName string) string {
    sql := fmt.Sprintf(`
MATCH (r:ROLE),(d:DATA)
WHERE r.name = "%s" AND d.name = "%s"
CREATE (r)-[o:OPERATE]->(d)
RETURN r, o, d
`, roleName, dataName)
    return sql
}

// 创建用户和角色的关系
func CreateRoleUserRelation(userName, roleName string) string {
    sql := fmt.Sprintf(`
MATCH (u:USER),(r:ROLE)
WHERE u.name = "%s" AND r.name = "%s"
CREATE (u)-[b:BELONG_TO]->(r)
RETURN u, b, r
`, userName, roleName)
    return sql
}

func DeleteRelationUserRole(userName, roleName string) string {
    // 删除 `用户 - 属于 -> 角色` 关系
    sql := fmt.Sprintf(`
MATCH (user:USER)-[r:BELONG_TO]->(role:ROLE)
WHERE user.name = "%s"`, userName)
    if roleName != "" {
        sql += fmt.Sprintf(` AND role.name = "%s"`, roleName)
    }
    sql += fmt.Sprintf(`
DELETE r
`)
    return sql
}

// 删除某一角色与所有用户的关联
func DeleteRoleAllUser(roleName string) string {
    // 删除 `用户 - 属于 -> 角色` 关系
    sql := fmt.Sprintf(`
MATCH (user:USER)-[r:BELONG_TO]->(role:ROLE)
WHERE role.name = "%s"
DELETE r`, roleName)
    return sql
}

// 删除某一角色与所有资源的关联
func DeleteRoleAllData(roleName string) string {
    // 删除 `角色 - 操作 -> 资源` 关系
    sql := fmt.Sprintf(`
MATCH (role:ROLE)-[o:OPERATE]->(data:DATA)
WHERE role.name = "%s"
DELETE o`, roleName)
    return sql
}

func DeleteRelationRoleData(roleName, dataName string) string {
    // 删除 `角色 - 操作 -> 资源` 关系
    sql := fmt.Sprintf(`
MATCH (role:ROLE)-[o:OPERATE]->(data:DATA)
WHERE role.name = "%s"`, roleName)
    if roleName != "" {
        sql += fmt.Sprintf(` AND data.name = "%s"`, dataName)
    }
    sql += fmt.Sprintf(`
DELETE o
`)
    return sql
}

func DeleteDatabase() string {
    return "match (n) detach delete n"
}


/*

################################inventory相关sql###############################################

*/
// 唯一性索引（分组名称）
func CreateUniqueGroup() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (group:GROUP)
ASSERT group.name IS UNIQUE
`)
    return sql
}

func CreateUniqueHost() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (host:HOST)
ASSERT host.name IS UNIQUE
`)
    return sql
}

// 创建GROUP（分组）标签
func CreateGroup(name, description, createdtime string) string {
    sql := fmt.Sprintf(`
CREATE (
  group:GROUP
  {
    name:"%s",
    description:"%s",
    createdtime:"%s",
    updatedtime:"%s"
  }
)
RETURN group.name as groupName, group.description as groupDescription, group.createdtime as groupCreatedtime, group.updatedtime as groupUpdatedtime
`, name, description, createdtime, createdtime)
    return sql
}

// 分组查询
func QueryGroup(groupName string) string {
    var sql string
    if groupName == "" {
        sql = fmt.Sprintf(`
MATCH (group:GROUP)
RETURN group.name as groupName, group.description as groupDescription, group.createdtime as groupCreatedtime, group.updatedtime as groupUpdatedtime
`)
    } else {
        sql = fmt.Sprintf(`
MATCH (group:GROUP)
WHERE group.name = "%s"
RETURN group.name as groupName, group.description as groupDescription, group.createdtime as groupCreatedtime, group.updatedtime as groupUpdatedtime
`, groupName)
    }

    return sql
}

// 更新GROUP（分组）标签
func SetGroup(name, description, updatedtime string) string {
    sql := fmt.Sprintf(`
MATCH (group:GROUP)
WHERE group.name="%s"
SET group.description = "%s", group.updatedtime = "%s"
RETURN group.name as groupName, group.description as groupDescription, group.createdtime as groupCreatedtime, group.updatedtime as groupUpdatedtime
`, name, description, updatedtime)
    return sql
}

// 主机查询
func QueryHost(hostName string) string {
    var sql string
    if hostName == "" {
        sql = fmt.Sprintf(`
MATCH (host:HOST)
RETURN host.name as hostName, host.remote_python_interpreter as hostRemotePythonInterpreter, host.description as hostDescription, host.createdtime as hostCreatedtime, host.updatedtime as hostUpdatedtime`)
    } else {
        sql = fmt.Sprintf(`
MATCH (host:HOST)
WHERE host.name = "%s"
RETURN host.name as hostName, host.remote_python_interpreter as hostRemotePythonInterpreter, host.description as hostDescription, host.createdtime as hostCreatedtime, host.updatedtime as hostUpdatedtime`, hostName)
    }
    return sql
}

// 创建HOST（主机）标签
func CreateHost(name, interpreter, description, createdtime string) string {
    // 默认用户为admin角色下admin用户；public用户下public用户
    sql := fmt.Sprintf(`
CREATE (
  host:HOST
  {
    name:"%s",
    remote_python_interpreter:"%s",
    description:"%s",
    createdtime:"%s",
    updatedtime:"%s"
  }
)
`, name, interpreter, description, createdtime, createdtime)

    return sql
}

// 更新HOST（主机）标签
func SetHost(name,  py, description, updatedtime string) string {
    sql := fmt.Sprintf(`
MATCH (host:HOST) 
WHERE host.name="%s"
SET host.updatedtime = "%s", host.description = "%s",`, name, updatedtime, description)
    if py != "" {
        sql += fmt.Sprintf(`host.remote_python_interpreter = "%s",`, py)
    }

    sql = strings.TrimRight(sql, ",")
    sql += fmt.Sprintf(`
RETURN host.name as hostName, host.remote_python_interpreter as hostRemotePythonInterpreter, host.description as hostDescription, host.createdtime as hostCreatedtime, host.updatedtime as hostUpdatedtime`)
    return sql
}

// 主机与分组所属关系条数
func QueryCountRelationHostGroup(groupName, hostName string) string {

    sql := fmt.Sprintf(`
MATCH (h:HOST{name:"%s"}),(g:GROUP{name:"%s"}),p=(h)-[r:RELATED]->(g)
return count(p) as GroupHostCount
`, hostName, groupName)
    return sql
}

// 创建主机和分组的关系
func CreateGroupUserRelation(hostName, groupName string) string {
    sql := fmt.Sprintf(`
MATCH (h:HOST),(g:GROUP)
WHERE h.name = "%s" AND g.name = "%s"
CREATE (h)-[b:RELATED]->(g)
RETURN h, b, g
`, hostName, groupName)
    return sql
}


func DeleteRelationHostGroup(hostName, groupName string) string {
    // 删除 `主机 - 关联 -> 分组` 关系
    sql := fmt.Sprintf(`
MATCH (h:HOST)-[r:RELATED]->(g:GROUP)
WHERE h.name = "%s"`, hostName)
    if groupName != "" {
        sql += fmt.Sprintf(` AND g.name = "%s"`, groupName)
    }
    sql += fmt.Sprintf(`
DELETE r
`)
    return sql
}

func GroupHasHost(groupName string) string {
    // 删除 `主机 - 关联 -> 分组` 关系
    sql := fmt.Sprintf(`
MATCH p=(host:HOST)-[r:RELATED]->(group:GROUP) 
RETURN host.name as hostName, host.remote_python_interpreter as hostRemotePythonInterpreter, host.description as hostDescription, host.createdtime as hostCreatedtime, host.updatedtime as hostUpdatedtime
`)
    return sql
}

// 删除主机
func DeleteHost(hostName string) string {
    // 删除 `主机` 节点
    sql := fmt.Sprintf(`
MATCH (host:HOST)
WHERE host.name = "%s"
DELETE host
`, hostName)
    return sql
}

// 删除某一分组与所有主机的关联
func DeleteGroupAllHost(groupName string) string {
    // 删除 `用户 - 属于 -> 角色` 关系
    sql := fmt.Sprintf(`
MATCH (host:HOST)-[r:RELATED]->(group:GROUP)
WHERE group.name = "%s"
DELETE r`, groupName)
    return sql
}

// 删除分组
func DeleteGroup(groupName string) string {
    // 删除 `角色` 节点
    sql := fmt.Sprintf(`
MATCH (g:GROUP)
WHERE g.name = "%s"
DELETE g
`, groupName)
    return sql
}

/*######################远程操作用户相关#####################*/


func CreateUniqueRemoteUser() string {
    sql := fmt.Sprintf(`
CREATE CONSTRAINT ON (r:REMOTEUSER)
ASSERT r.name IS UNIQUE
`)
    return sql
}

// 远程操作用户查询
func QueryRemoteUser(remoteUser string) string {
    // 用户名格式其实是  host@username
    var sql string
    if remoteUser == "" {
        sql = fmt.Sprintf(`
MATCH (r:REMOTEUSER)
RETURN r.name as remoteUser`)
    } else {
        sql = fmt.Sprintf(`
MATCH (r:REMOTEUSER)
WHERE r.name = "%s"
RETURN r.name as remoteUser`, remoteUser)
    }
    return sql
}


// 创建REMOTEUSER（远程主机的操作用户）标签
func CreateRemnoteUser(name, pass string) string {
    sql := fmt.Sprintf(`
CREATE (
  r:REMOTEUSER
  {
    name:"%s",
    password:"%s"
  }
)
`, name, pass)

    return sql
}


// 创建主机和远程操作用户的关系
func CreateHostRemoteUserRelation(hostName, RemoteUser string) string {
    sql := fmt.Sprintf(`
MATCH (h:HOST),(ru:REMOTEUSER)
WHERE h.name = "%s" AND r.name = "%s"
CREATE (h)-[have:HAVE]->ru)
RETURN h, have, ru
`, hostName, RemoteUser)
    return sql
}