## 1. 默认权限

resolver.perm: 针对graphql接口
views.perm: 针对restful接口

```vim
[write]
CreateRole = admin
CreateUser = amdin
CreateDate
Logout

[read]
CreateRole = admin
Users = amdin,public
Datas = amdin,public
```

- `write`: 表示SET方式（增删改资源）的rest接口、graphql节点。
- `read`: 表示QUERY方式（获取资源）的rest接口、graphql节点。
- `CreateRole = admin`: 含义为： `节点/视图 = 默认角色（多个角色以逗号隔开）`

## 2. 初始化graphql
```shell
$ go run github.com/99designs/gqlgen init
```