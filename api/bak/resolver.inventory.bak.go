package inventory

import (
	"apiring/auth"
	"apiring/configs"
	"apiring/utils"
	"context"
	sql "apiring/database"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/pkg/errors"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

// 创建主机分组（已测）
func (r *mutationResolver) CreateGroup(ctx context.Context, input GroupInput) (*Group, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Group{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()
	// 检测分组是否已存在
	groupExists, err := session.Run(sql.QueryGroup(input.Name), map[string]interface{}{})
	if err != nil {
		return &Group{}, err
	}
	var count int64
	for groupExists.Next() {
		count += 1
	}
	if count > 0 {
		return &Group{}, errors.Errorf("分组[%s]已存在", input.Name)
	}

	// 不存在则创建
	rs, err := session.Run(sql.CreateGroup(input.Name, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Group{}, err
	}
	var group *Group
	for rs.Next() {
		record := rs.Record()
		name, _ := record.Get("groupName")
		description, _ := record.Get("groupDescription")
		createdtime, _ := record.Get("groupCreatedtime")
		updatedtime, _ := record.Get("groupUpdatedtime")
		group = &Group{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
	}
	return group, nil
}

// 更新主机分组信息（已测）
// 目前只能更新 description
func (r *mutationResolver) UpdateGroup(ctx context.Context, input GroupInput) (*Group, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Group{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	rs, err := session.Run(sql.SetGroup(input.Name, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Group{}, err
	}

	var group *Group
	for rs.Next() {
		record := rs.Record()
		name, _ := record.Get("groupName")
		description, _ := record.Get("groupDescription")
		createdtime, _ := record.Get("groupCreatedtime")
		updatedtime, _ := record.Get("groupUpdatedtime")
		group = &Group{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
	}
	return group, nil
}

// 删除主机分组(已测)
func (r *mutationResolver) DeleteGroup(ctx context.Context, name string) (bool, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return false, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	// 查询分组是否存在
	groupExists, err := session.Run(sql.QueryGroup(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	var count int64
	for groupExists.Next() {
		count += 1
	}
	if count <= 0 {
		return false, errors.Errorf("分组[%s]不存在", name)
	}

	// 删除分组与其下所有主机的关联
	_, err = session.Run(sql.DeleteGroupAllHost(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	// 移除分组
	_, err = session.Run(sql.DeleteGroup(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// 创建主机(已测)
func (r *mutationResolver) CreateHost(ctx context.Context, input HostInput) (*Host, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Host{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	// 查询用户是否存在
	hostExists, err := session.Run(sql.QueryHost(input.Name), map[string]interface{}{})
	if err != nil {
		return &Host{}, err
	}
	var count int64
	for hostExists.Next() {
		count += 1
	}
	if count > 0 {
		return &Host{}, errors.Errorf("主机[%s]已存在", input.Name)
	}

	// 主机不存在则创建主机
	// pass := utils.Decrypt(input.RemotePass, configs.Logger) // 密码加密
	_, err = session.Run(sql.CreateHost(input.Name, input.RemotePythonInterpreter, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Host{}, err
	}

	host := &Host{Name:input.Name, Description:input.Description, RemotePythonInterpreter:input.RemotePythonInterpreter, CreateTime:utils.CurrentTime(), UpdateTime:utils.CurrentTime()}
	return host, nil
}

// 更新主机信息（已测）
func (r *mutationResolver) UpdateHost(ctx context.Context, input SetHostInput) (*Host, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Host{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()


	var hostName string
	if input.Name == nil {
		return &Host{}, errors.Errorf("主机名不可为空")
	} else {
		hostName = *input.Name
	}


	var hostDes string
	if input.Description == nil {
		hostDes = ""
	} else {
		hostDes = *input.Description
	}

	var hostPy string
	if input.RemotePythonInterpreter == nil {
		hostPy = ""
	} else {
		hostPy = *input.RemotePythonInterpreter
	}

	result, err := session.Run(sql.SetHost(hostName, hostPy, hostDes, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Host{}, err
	}

	for result.Next() {
		record := result.Record()
		name, _ := record.Get("hostName")
		py, _ := record.Get("hostRemotePythonInterpreter")
		description, _ := record.Get("hostDescription")
		createdtime, _ := record.Get("hostCreatedtime")
		updatedtime, _ := record.Get("hostUpdatedtime")
		h := &Host{Name:name.(string), Description:description.(string), RemotePythonInterpreter:py.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		return h, nil
	}

	return &Host{}, nil
}

// 删除主机（已测）
func (r *mutationResolver) DeleteHost(ctx context.Context, name string) (bool, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return false, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()
	// 查询主机是否存在
	hostExists, err := session.Run(sql.QueryHost(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	var count int64
	for hostExists.Next() {
		count += 1
	}
	if count <= 0 {
		return false, errors.Errorf("主机[%s]不存在", name)
	}

	// 存在，首先解除所有服务关联
	_, err = session.Run(sql.DeleteRelationHostGroup(name, ""), map[string]interface{}{})

	// 删除主机
	_, err = session.Run(sql.DeleteHost(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	configs.Logger.Warnf("主机[%s]已删除！", name)
	return true, nil
}

// 创建host对应的远程操作用户与密码（用户名规则：user@host），同时创建用户名与host的对应关系
func (r *mutationResolver) CreateRemoteUserPass(ctx context.Context, input UserPassInput) (*RemoteUserPass, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &RemoteUserPass{}, e
	}

	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	username := input.RemoteUser + "@" + input.Host
	// 查询用户名是否存在
	userExists, err := session.Run(sql.QueryRemoteUser(username), map[string]interface{}{})

	if err != nil {
		return &RemoteUserPass{}, err
	}
	var count int64
	for userExists.Next() {
		count += 1
	}
	if count > 0 {
		return &RemoteUserPass{}, errors.Errorf("主机[%s]的用户[%s]已存在", input.Host, input.RemoteUser)
	}

	// 不存在则创建用户
	_, err = session.Run(sql.CreateHost(input.Name, input.RemotePythonInterpreter, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Host{}, err
	}
}

// 只能更新密码
func (r *mutationResolver) UpdateRemoteUserPass(ctx context.Context, input UserPassInput) (*RemoteUserPass, error) {
	panic("not implemented")
}

// 删除用户（同时会删除用户名与host的对应关系）
func (r *mutationResolver) DeleteRemoteUserPass(ctx context.Context, name DeleteUserPassInput) (bool, error) {
	panic("not implemented")
}

// 创建主机与分组的关系(已测)
func (r *mutationResolver) CreateHostBelong(ctx context.Context, input HostBelongInput) (bool, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return false, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	// 查询主机与分组是否关联
	ghCount, err := session.Run(sql.QueryCountRelationHostGroup(input.GroupName, input.HostName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for ghCount.Next() {
		record := ghCount.Record()
		c, _ := record.Get("GroupHostCount")
		if c.(int64) > 0 {
			return false, errors.Errorf("主机[%s]与分组[%s]关系已存在", input.HostName, input.GroupName)
		}
	}

	// 关系不存在则创建关系
	_, err = session.Run(sql.CreateGroupUserRelation(input.HostName, input.GroupName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

// 删除主机与分组关系（已测）
func (r *mutationResolver) DeleteHostBelong(ctx context.Context, input HostBelongInput) (bool, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return false, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	// 查询主机与分组是否关联
	ghCount, err := session.Run(sql.QueryCountRelationHostGroup(input.GroupName, input.HostName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for ghCount.Next() {
		record := ghCount.Record()
		c, _ := record.Get("GroupHostCount")
		if c.(int64) <= 0 {
			return false, errors.Errorf("主机[%s]与分组[%s]关系不存在，无法删除!", input.HostName, input.GroupName)
		}
	}

	// 关联存在，则移除
	_, err = session.Run(sql.DeleteRelationHostGroup(input.HostName, input.GroupName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

type queryResolver struct{ *Resolver }

// 查询分组（已测）
func (r *queryResolver) Groups(ctx context.Context, name *string) ([]*Group, error) {
	e := auth.ResourcePermVerify(ctx)
	var rs []*Group
	if e != nil {
		return rs, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	var hostName string
	if name == nil {
		hostName = ""
	} else {
		hostName = *name
	}

	results, err := session.Run(sql.QueryGroup(hostName), map[string]interface{}{})
	if err != nil {
		return rs, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("groupName")
		description, _ := record.Get("groupDescription")
		createdtime, _ := record.Get("groupCreatedtime")
		updatedtime, _ := record.Get("groupUpdatedtime")
		r := &Group{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		rs = append(rs, r)
	}
	return rs, nil
}

// 用户查询（已测）
func (r *queryResolver) Hosts(ctx context.Context, name *string) ([]*Host, error) {
	e := auth.ResourcePermVerify(ctx)
	var hs []*Host
	if e != nil {
		return hs, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	var hostName string
	if name == nil {
		hostName = ""
	} else {
		hostName = *name
	}

	results, err := session.Run(sql.QueryHost(hostName), map[string]interface{}{})
	if err != nil {
		return hs, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("hostName")
		py, _ := record.Get("hostRemotePythonInterpreter")
		description, _ := record.Get("hostDescription")
		createdtime, _ := record.Get("hostCreatedtime")
		updatedtime, _ := record.Get("hostUpdatedtime")
		h := &Host{Name:name.(string), Description:description.(string), RemotePythonInterpreter:py.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		hs = append(hs, h)
	}
	return hs, nil
}

// 分组关联的主机(已测)
func (r *queryResolver) GroupHas(ctx context.Context, groupName string) ([]*Host, error) {
	e := auth.ResourcePermVerify(ctx)
	var hs []*Host
	if e != nil {
		return hs, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()


	results, err := session.Run(sql.GroupHasHost(groupName), map[string]interface{}{})
	if err != nil {
		return hs, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("hostName")
		py, _ := record.Get("hostRemotePythonInterpreter")
		description, _ := record.Get("hostDescription")
		createdtime, _ := record.Get("hostCreatedtime")
		updatedtime, _ := record.Get("hostUpdatedtime")
		h := &Host{Name:name.(string), Description:description.(string), RemotePythonInterpreter:py.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		hs = append(hs, h)
	}
	return hs, nil
}
