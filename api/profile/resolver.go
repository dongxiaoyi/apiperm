package profile

import (
	"apiring/auth"
	"apiring/configs"
	sql "apiring/database"
	"apiring/utils"
	"context"
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

// 创建角色(已测)
func (r *mutationResolver) CreateRole(ctx context.Context, input RoleInput) (*Role, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Role{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()
	// 检测角色是否已存在
	roleExists, err := session.Run(sql.QueryRole(input.Name), map[string]interface{}{})
	if err != nil {
		return &Role{}, err
	}
	var count int64
	for roleExists.Next() {
		count += 1
	}
	if count > 0 {
		return &Role{}, errors.Errorf("角色[%s]已存在", input.Name)
	}

	// 不存在则创建
	rs, err := session.Run(sql.CreateRole(input.Name, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Role{}, err
	}
	var role *Role
	for rs.Next() {
		record := rs.Record()
		name, _ := record.Get("roleName")
		description, _ := record.Get("roleDescription")
		createdtime, _ := record.Get("roleCreatedtime")
		updatedtime, _ := record.Get("roleUpdatedtime")
		role = &Role{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
	}
	return role, nil
}

// 更新角色（已测）
func (r *mutationResolver) UpdateRole(ctx context.Context, input RoleInput) (*Role, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &Role{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	rs, err := session.Run(sql.SetRole(input.Name, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &Role{}, err
	}

	var role *Role
	for rs.Next() {
		record := rs.Record()
		name, _ := record.Get("roleName")
		description, _ := record.Get("roleDescription")
		createdtime, _ := record.Get("roleCreatedtime")
		updatedtime, _ := record.Get("roleUpdatedtime")
		role = &Role{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
	}
	return role, nil
}

// 删除角色（已测试）
// 删除角色，将移出其下所有用户的关联、移出其所有资源操作的关联
func (r *mutationResolver) DeleteRole(ctx context.Context, name string) (bool, error) {
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

	// 查询角色是否存在
	roleExists, err := session.Run(sql.QueryRole(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	var count int64
	for roleExists.Next() {
		count += 1
	}
	if count <= 0 {
		return false, errors.Errorf("角色[%s]不存在", name)
	}

	// 删除角色与其下所有用户的关联
	_, err = session.Run(sql.DeleteRoleAllUser(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	// 删除角色与所有资源的关联
	_, err = session.Run(sql.DeleteRoleAllData(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	// 移除角色
	_, err = session.Run(sql.DeleteRole(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// 创建用户(已测)
// 新创建的用户，默认属于public分组
func (r *mutationResolver) CreateUser(ctx context.Context, input UserInput) (*User, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &User{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	// 查询用户是否存在
	userExists, err := session.Run(sql.QueryUser(input.Name), map[string]interface{}{})
	if err != nil {
		return &User{}, err
	}
	var count int64
	for userExists.Next() {
		count += 1
	}
	if count > 0 {
		return &User{}, errors.Errorf("用户[%s]已存在", input.Name)
	}

	// 用户不存在则创建用户
	pass := utils.Decrypt(input.Passwd, configs.Logger) // 密码加密
	_, err = session.Run(sql.CreateUser(input.Name, pass, input.Description, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &User{}, err
	}

	// 赋予用户 角色
	_, err = session.Run(sql.RelationUserRole(input.Name, "public", utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		// 角色赋予失败则删除用户
		session.Run(sql.DeleteUser(input.Name), map[string]interface{}{})
		return &User{}, err
	}

	user := &User{Name:input.Name, Description:input.Description, Password:"加密不可见", CreateTime:utils.CurrentTime(), UpdateTime:utils.CurrentTime()}
	return user, nil
}

// 更新用户信息（已测）
// 当前版本用户名不可修改
func (r *mutationResolver) UpdateUser(ctx context.Context, input SetUserInput) (*User, error) {
	e := auth.ResourcePermVerify(ctx)
	if e != nil {
		return &User{}, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()


	var userName string
	if input.Name == nil {
		return &User{}, errors.Errorf("用户名不可为空")
	} else {
		userName = *input.Name
	}

	var userPass string
	if input.Passwd == nil {
		userPass = ""
	} else {
		userPass = utils.Decrypt(*input.Passwd, configs.Logger)
	}

	var userDes string
	if input.Description == nil {
		userDes = ""
	} else {
		userDes = *input.Description
	}
	result, err := session.Run(sql.SetUser(userName, userPass, userDes, utils.CurrentTime()), map[string]interface{}{})
	if err != nil {
		return &User{}, err
	}

	for result.Next() {
		record := result.Record()
		name, _ := record.Get("userName")
		description, _ := record.Get("userDescription")
		createdtime, _ := record.Get("userCreatedtime")
		updatedtime, _ := record.Get("userUpdatedtime")
		u := &User{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		return u, nil
	}

	return &User{}, nil
}

// 删除用户（已测）
// 删除某个用户，将会移出其与所有角色的关联，且彻底删除
func (r *mutationResolver) DeleteUser(ctx context.Context, name string) (bool, error) {
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
	// 查询用户是否存在
	userExists, err := session.Run(sql.QueryUser(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	var count int64
	for userExists.Next() {
		count += 1
	}
	if count <= 0 {
		return false, errors.Errorf("用户[%s]不存在", name)
	}

	// 存在，首先解除所有角色关联
	_, err = session.Run(sql.DeleteRelationUserRole(name, ""), map[string]interface{}{})

	// 删除用户
	_, err = session.Run(sql.DeleteUser(name), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	configs.Logger.Warnf("用户[%s]已删除！", name)
	return true, nil
}

// 资源存放于.perm文件，不可创建、更新、删除

// 创建用户与角色的关联（已测）
func (r *mutationResolver) CreateUserBelong(ctx context.Context, input UserBelongInput) (bool, error) {
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

	// 查询用户与角色是否关联
	ruCount, err := session.Run(sql.QueryCountRelationRoleUser(input.RoleName, input.UserName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for ruCount.Next() {
		record := ruCount.Record()
		c, _ := record.Get("RoleUserCount")
		if c.(int64) > 0 {
			return false, errors.Errorf("用户[%s]与角色[%s]关系已存在", input.UserName, input.RoleName)
		}
	}

	// 关系不存在则创建关系
	_, err = session.Run(sql.CreateRoleUserRelation(input.UserName, input.RoleName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

// 删除用户与角色的关联（已测）
func (r *mutationResolver) DeleteUserBelong(ctx context.Context, input UserBelongInput) (bool, error) {
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

	// 查询用户与角色是否关联
	ruCount, err := session.Run(sql.QueryCountRelationRoleUser(input.RoleName, input.UserName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for ruCount.Next() {
		record := ruCount.Record()
		c, _ := record.Get("RoleUserCount")
		if c.(int64) <= 0 {
			return false, errors.Errorf("用户[%s]与角色[%s]关系不存在，无法删除!", input.UserName, input.RoleName)
		}
	}

	// 关联存在，则移除
	_, err = session.Run(sql.DeleteRelationUserRole(input.UserName, input.RoleName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

// 创建角色与资源的操作权限（已测）
func (r *mutationResolver) CreateRoleOperate(ctx context.Context, input RoleOperateInput) (bool, error) {
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

	// 查询角色与资源是否已有关联
	rdCount, err := session.Run(sql.QueryCountRelationRoleData(input.RoleName, input.DataName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for rdCount.Next() {
		record := rdCount.Record()
		c, _ := record.Get("RoleDataCount")
		if c.(int64) > 0 {
			return false, errors.Errorf("角色[%s]与资源[%s]关系已存在，无法创建!", input.RoleName, input.DataName)
		}
	}

	// 不存在则创建
	_, err = session.Run(sql.CreateRoleDataRelation(input.RoleName, input.DataName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

// 删除角色操作资源的权限（已测）
func (r *mutationResolver) DeleteRoleOperate(ctx context.Context, input RoleOperateInput) (bool, error) {
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

	// 查询角色与资源是否已有关联
	rdCount, err := session.Run(sql.QueryCountRelationRoleData(input.RoleName, input.DataName), map[string]interface{}{})
	if err != nil {
		return false, err
	}
	for rdCount.Next() {
		record := rdCount.Record()
		c, _ := record.Get("RoleDataCount")
		if c.(int64) <= 0 {
			return false, errors.Errorf("角色[%s]与资源[%s]关系不存在，无需删除!", input.RoleName, input.DataName)
		}
	}

	// 关系存在，则删除
	_, err = session.Run(sql.DeleteRelationRoleData(input.RoleName, input.DataName), map[string]interface{}{})
	if err != nil {
		return false, err
	}

	return true, nil
}

type queryResolver struct{ *Resolver }

// 查询角色（已测）
func (r *queryResolver) Roles(ctx context.Context, name *string) ([]*Role, error) {
	e := auth.ResourcePermVerify(ctx)
	var rs []*Role
	if e != nil {
		return rs, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	var roleName string
	if name == nil {
		roleName = ""
	} else {
		roleName = *name
	}

	results, err := session.Run(sql.QueryRole(roleName), map[string]interface{}{})
	if err != nil {
		return rs, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("roleName")
		description, _ := record.Get("roleDescription")
		createdtime, _ := record.Get("roleCreatedtime")
		updatedtime, _ := record.Get("roleUpdatedtime")
		r := &Role{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		rs = append(rs, r)
	}
	return rs, nil
}

// 查询用户（已测）
func (r *queryResolver) Users(ctx context.Context, name *string) ([]*User, error) {
	e := auth.ResourcePermVerify(ctx)
	var us []*User
	if e != nil {
		return us, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	var userName string
	if name == nil {
		userName = ""
	} else {
		userName = *name
	}

	results, err := session.Run(sql.QueryUser(userName), map[string]interface{}{})
	if err != nil {
		return us, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("userName")
		description, _ := record.Get("userDescription")
		createdtime, _ := record.Get("userCreatedtime")
		updatedtime, _ := record.Get("userUpdatedtime")
		u := &User{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		us = append(us, u)
	}
	return us, nil
}

// 查询资源（已测）
func (r *queryResolver) Datas(ctx context.Context, name *string, mode *string, style *string, service *string) ([]*Data, error) {
	e := auth.ResourcePermVerify(ctx)
	var ds []*Data
	if e != nil {
		return ds, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	var dataName string
	if name == nil {
		dataName = ""
	} else {
		dataName = *name
	}

	var dataMode string
	if mode == nil {
		dataMode = ""
	} else {
		dataMode = *mode
	}

	var dataStyle string
	if style == nil {
		dataStyle = ""
	} else {
		dataStyle = *style
	}

	var dataService string
	if service == nil {
		dataService = ""
	} else {
		dataService = *service
	}

	results, err := session.Run(sql.QueryData(dataName, dataMode, dataStyle, dataService), map[string]interface{}{})
	if err != nil {
		return ds, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("dataName")
		description, _ := record.Get("dataDescription")
		modeS, _ := record.Get("dataMode")
		styleS, _ := record.Get("dataStyle")
		serviceS, _ := record.Get("dataService")
		createdtime, _ := record.Get("dataCreatedtime")
		updatedtime, _ := record.Get("dataUpdatedtime")
		u := &Data{Name:name.(string), Description:description.(string), Mode:modeS.(string), Style:styleS.(string), Service:serviceS.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		ds = append(ds, u)
	}
	return ds, nil
}

// 查询用户属于哪些角色（已测）
func (r *queryResolver) UserBelong(ctx context.Context, userName string) ([]*Role, error) {
	e := auth.ResourcePermVerify(ctx)
	var rs []*Role
	if e != nil {
		return rs, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	results, err := session.Run(sql.QueryUserBelongRole(userName), map[string]interface{}{})
	if err != nil {
		return rs, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("roleName")
		description, _ := record.Get("roleDescription")
		createdtime, _ := record.Get("roleCreatedtime")
		updatedtime, _ := record.Get("roleUpdatedtime")
		r := &Role{Name:name.(string), Description:description.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		rs = append(rs, r)
	}
	return rs, nil
}

// 查询角色拥有操作哪些资源的权限（已测）
func (r *queryResolver) RoleOperate(ctx context.Context, roleName string) ([]*Data, error) {
	e := auth.ResourcePermVerify(ctx)
	var ds []*Data
	if e != nil {
		return ds, e
	}
	var session neo4j.Session
	var err error
	if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
		configs.Logger.Error(err)
	}
	defer session.Close()

	results, err := session.Run(sql.QueryRoleoperateData(roleName), map[string]interface{}{})
	if err != nil {
		return ds, err
	}
	for results.Next() {
		record := results.Record()
		name, _ := record.Get("dataName")
		description, _ := record.Get("dataDescription")
		modeS, _ := record.Get("dataMode")
		styleS, _ := record.Get("dataStyle")
		serviceS, _ := record.Get("dataService")
		createdtime, _ := record.Get("dataCreatedtime")
		updatedtime, _ := record.Get("dataUpdatedtime")
		u := &Data{Name:name.(string), Description:description.(string), Mode:modeS.(string), Style:styleS.(string), Service:serviceS.(string), CreateTime:createdtime.(string), UpdateTime:updatedtime.(string)}
		ds = append(ds, u)
	}
	return ds, nil
}
