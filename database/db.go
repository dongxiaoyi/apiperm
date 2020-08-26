package database

import (
	"apiring/models"
	"apiring/utils"
	"github.com/Unknwon/goconfig"
	_ "github.com/lib/pq"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
)


func InitNeo4j(driver neo4j.Driver, logger *zap.SugaredLogger, pass *models.DefaultPassword) {
	// 创建默认用户admin & public
	var session neo4j.Session
	var err error
	if session, err = driver.Session(neo4j.AccessModeWrite); err != nil {
		logger.Error(err)
	}
	defer session.Close()

	// 检查、创建唯一性约束
	session.Run(CreateUniqueData(), map[string]interface{}{})  // 资源
	session.Run(CreateUniqueUser(), map[string]interface{}{})  // 用户
	session.Run(CreateUniqueRole(), map[string]interface{}{})  // 角色
	session.Run(CreateUniqueGroup(), map[string]interface{}{}) // 主机分组
	session.Run(CreateUniqueHost(), map[string]interface{}{})  // 主机
	session.Run(CreateUniqueRemoteUser(), map[string]interface{}{})  // 主机远程操作用户

	// 用户密码加密
	passAdmin := utils.Decrypt(pass.Admin, logger)
	passPublic := utils.Decrypt(pass.Public, logger)

	// 初始化默认角色、用户、资源、权限检查
	// TODO：首次导入初始角色、资源权限
	checkRoleUserDataPermDefault(driver, "admin", "默认角色：admin", "admin", passAdmin,"默认用户：admin", logger)
	checkRoleUserDataPermDefault(driver, "public", "默认角色：public", "public", passPublic,"默认用户：public", logger)

	logger.Info("数据库初始化，完毕！")
}


func checkRoleUserDataPermDefault(driver neo4j.Driver, roleName, descriptionRole, userName, pass, descriptionUser string, logger *zap.SugaredLogger) {
	var session neo4j.Session
	var err error
	if session, err = driver.Session(neo4j.AccessModeWrite); err != nil {
		logger.Error(err)
	}
	defer session.Close()
	// 检查是否存在分组
	role, err := session.Run(QueryRole(roleName), map[string]interface{}{})
	if !role.Next() {
		_, err = session.Run(CreateRole(roleName, descriptionRole, utils.CurrentTime()), map[string]interface{}{})
		if err != nil {
			logger.Error(err)
		} else {
			logger.Infof("初始化创建角色 -> 角色: %s", roleName)
		}
	}

	// 检查角色下用户存在
	user, err := session.Run(QueryUser(userName), map[string]interface{}{})

	if !user.Next() {
		_, err = session.Run(CreateUser(userName, pass, descriptionUser, utils.CurrentTime()), map[string]interface{}{})
		if err != nil {
			logger.Error(err)
		} else {
			logger.Infof("初始化创建用户 -> 账号: %s", userName)
		}
	}

	// 判断是否存在所属关系
	countRoleUser, err := session.Run(QueryCountRelationRoleUser(roleName, userName), map[string]interface{}{})

	for countRoleUser.Next() {
		if countRoleUser.Record().GetByIndex(0).(int64) == 0 {
			// 创建初始化用户与角色的所属关系
			_, relationErr := session.Run(RelationUserRole(roleName, userName, utils.CurrentTime()), map[string]interface{}{})
			if relationErr != nil {
				logger.Error(err)
			} else {
				logger.Infof("初始化创建关系: (%s:用户)-[属于]->(%s:角色)", userName, roleName)
			}
		}
	}
}


// 生成各gql下的初始权限配置
func GenPermsDefault(driver neo4j.Driver, logger *zap.SugaredLogger) {
	absDir := utils.AbsPath()
	gqlPath := filepath.Join(absDir, "api")
	appList , _ := filepath.Glob(gqlPath+"/*")
	for _, pathApp := range appList {
		s, _ := os.Stat(pathApp)  // 文件状态
		isDir := s.IsDir()
		if isDir {
			dirBase := path.Base(pathApp)
			// 只有路径是目录的情况下才是app
			// 判断是否存在权限配置文件(resolver.perm/views.perm)
			permViewsPath := filepath.Join(pathApp, "views.perm")
			permResolverPath := filepath.Join(pathApp, "resolver.perm")
			style := switchPermStyle(permViewsPath)
			if style == "" {
				style = switchPermStyle(permResolverPath)
			}

			// 根据不同的视图类型（restfule/graphql生成资源数据）
			switch style {
			case "rest":
				setDataStyleDefault(driver, logger, permViewsPath, "rest", dirBase)
			case "graph":
				setDataStyleDefault(driver, logger, permResolverPath, "graph", dirBase)
			case "":
				logger.Warnf("`%s` 或者 `%s` 权限文件未定义！", permViewsPath, permResolverPath)
			default:
				logger.Warnf("`%s` 或者 `%s` 权限文件未定义！", permViewsPath, permResolverPath)
			}
		}

	}
}

// 根据权限文件生成资源数据
func setDataStyleDefault(driver neo4j.Driver, logger *zap.SugaredLogger, permFile, style, app string) {
	var session neo4j.Session
	var err error
	if session, err = driver.Session(neo4j.AccessModeWrite); err != nil {
		logger.Error(err)
	}
	defer session.Close()

	permConfig, err := goconfig.LoadConfigFile(permFile)
	if err != nil {
		logger.Error("Get config file error")
		os.Exit(-1)
	}
	nodeList := permConfig.GetSectionList()
	if len(nodeList) >= 0 {
		for _, node := range nodeList {
			roles := permConfig.MustValueArray(node, "roles", ",")
			description := permConfig.MustValue(node, "description", ",")
			mode := permConfig.MustValue(node, "mode", ",")
			// 不存在资源即创建
			mergeData := CreateData(node, description, mode, app, style, utils.CurrentTime())
			_, err := session.Run(mergeData, map[string]interface{}{})
			if err != nil {
				logger.Warnf("资源 %s 已存在！", node)
			}
			// 资源应用于默认roles
			for _, r := range roles {
				session.Run(MergeRelationRoleData(r, node), map[string]interface{}{})
			}
		}
	}
}

func switchPermStyle(file string) string {
	// 根据文件名称，返回权限类型：rest、graph、nothing
	style := ""
	_, err := os.Stat(file)
	if err == nil || os.IsExist(err) {
		fileSuffix := path.Base(file)
		if fileSuffix == "views.perm" {
			return "rest"
		} else if fileSuffix == "resolver.perm" {
			return "graph"
		}
	}
	return style
}