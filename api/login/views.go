package login

import (
	"apiring/configs"
	"apiring/database"
	"apiring/utils"
	"encoding/json"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"io/ioutil"
	"net/http"
)

type Req struct {
	Name string `json:"name""`
	Password string `json:"password"`
}

type Res struct {
	Name string `json:"name""`
	Message string `json:"message"`
}

func LoginView(w http.ResponseWriter, r *http.Request) {
		var input Req
		var output Res

		json.NewDecoder(r.Body).Decode(&input)
		var session neo4j.Session
		var err error
		if session, err = configs.Neo4j.Session(neo4j.AccessModeWrite); err != nil {
			http.Error(w, err.Error(), 403)
		}
		defer session.Close()
		// 解析password
		pass := utils.Decrypt(input.Password, configs.Logger)

		userIns, err := session.Run(database.QueryUserPasswd(input.Name, pass), map[string]interface{}{})
		if err != nil {
			http.Error(w, err.Error(), 403)
		}
		// 用户不存在
		hasUser := userIns.Next()
		if !hasUser {
			http.Error(w, "用户不存在或者密码错误!", 403)
		}

		// 用户存在，登录操作
		tokenString := configs.JWT.Encoding(input.Name)
		input.Name = userIns.Record().GetByIndex(0).(neo4j.Node).Props()["name"].(string)

		c := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true}
		http.SetCookie(w, &c)

		output.Name = input.Name
		output.Message = "登录成功！"

		response, _ := json.Marshal(output)
		w.Write(response)
}


type Gohangout struct {
	status int64
	message string
}

// 测试gohangout的视图（可删）
func GohangoutView(w http.ResponseWriter, r *http.Request) {
	var s []map[string]interface{}
	b, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(b))
	fmt.Println("++++++++++++++++++++++++++++")
	json.Unmarshal(b, &s)
	for i, m := range s {
		st, _ := json.Marshal(m)
		fmt.Println(i, string(st), "\n")
	}
	fmt.Println("############################")
	var res Gohangout
	res.status = 200
	res.message = "success"
	response, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Write(response)
}