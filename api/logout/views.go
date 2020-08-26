package logout

import (
	"encoding/json"
	"net/http"
	"time"
)

type Res struct {
	Name string `json:"name""`
	Message string `json:"message"`
}


func LogoutView(w http.ResponseWriter, r *http.Request) {
	var out Res
	name := r.Context().Value("user_name")

	http.SetCookie(w, &http.Cookie{
		Name: "jwt",
		MaxAge: -1,
		Expires: time.Now().Add(-100 * time.Hour),// Set expires for older versions of IE
		Path: "/",
	})

	out.Name = name.(string)
	out.Message = "success"
	response, _ := json.Marshal(out)

	w.Write(response)
}

