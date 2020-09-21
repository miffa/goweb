package auth

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	LoginUri  = "/api/v1/login/db" //post
	LogoutUri = "/api/v1/logout"

	NsAuthUri  = "/api/v1/current/authenticate/%s"
	AppAuthUri = "/api/v1/current/authenticate/%s/%s"

	TokenUriLogin   = "/api/v1/current/user"
	TokenUriNoLogin = "/api/v1/current/token"

	ICanAccess    = 1
	ICanNotAccess = 2

	HttpTimeout = 5
)

var (
	authurl []string
)

func Init(authurls []string) error {
	authurl = authurls
	return nil
}

//////////////////
type TpaasRespHeader struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type JwtToken struct {
	Token string `json:"token,omitempty"`
}

type User struct {
	OpsUser

	Password   string    `json:"password"`
	LastLogin  time.Time `json:"lastLogin"`
	LastIP     string    `json:"lastIp"`
	CreateTime time.Time `json:"createTime"`
}

type OpsUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
}

type TpaasPermission struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Enable  int    `json:"enable"`
}

type TpaasTokenSessionResp struct {
	TpaasRespHeader
	Data User `json:"data"`
}

type TpaasTokenResp struct {
	TpaasRespHeader
	Data string `json:"data"`
}

type TpaasPermissionResp struct {
	TpaasRespHeader
	Data map[string][]TpaasPermission `json:"data"`
}

type TpaasLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TpaasLoginResp struct {
	TpaasRespHeader
	Data JwtToken `json:"data"`
}

/////////////////////

func idx() int {
	l := len(authurl)
	return rand.Intn(l) % l
}
func loginUrl() string {
	return authurl[idx()] + LoginUri
}

func tokenUrlWithNoSession() string {
	return authurl[idx()] + TokenUriNoLogin
}
func tokenUrlWithSession() string {
	return authurl[idx()] + TokenUriLogin
}

func authUrl4Ns(ns string) string {
	return fmt.Sprintf(authurl[idx()]+NsAuthUri, ns)
}
func authUrl4App(ns, app string) string {
	return fmt.Sprintf(authurl[idx()]+AppAuthUri, ns, app)
}

func HttpOK(i int) bool {
	return i >= 200 && i < 226 //  from https://golang.org/src/net/http/status.go
}
