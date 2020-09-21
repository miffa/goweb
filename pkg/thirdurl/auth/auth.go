package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

//// login
// return : jwtoken, error
func Login(user, passd string) (string, error) {
	t := TpaasLogin{Username: user, Password: passd}

	client := gorequest.New().Post(loginUrl()).
		Timeout(HttpTimeout * time.Second).
		Type("json").
		Send(t)

	resp, body, ierrors := client.End()
	if len(ierrors) != 0 {
		return "", ierrors[0]
	}

	if !HttpOK(resp.StatusCode) {
		return "", errors.Errorf("http code:%d body:%s", resp.StatusCode, body)
	}

	var lg TpaasLoginResp
	err := json.Unmarshal([]byte(body), &lg)
	if err != nil {
		return "", errors.WithMessage(err, "logining  response from tpaas is not json")
	}
	if !HttpOK(lg.Code) {
		return "", errors.Errorf("tpaas code:%d body:%s", lg.Code, body)
	}
	return lg.Data.Token, nil
}

//// logout
func Logout(token string) error {
	client := gorequest.New().Get(loginUrl()).
		AppendHeader("Authorization", "Bearer "+token).
		Timeout(HttpTimeout * time.Second).
		Query("")

	resp, body, ierrors := client.End()
	if len(ierrors) != 0 {
		return ierrors[0]
	}

	if !HttpOK(resp.StatusCode) {
		return errors.Errorf("http code:%d body:%s", resp.StatusCode, body)
	}
	return nil
}

//// token verification for no logined user
// return username
func TokenNoSession(token string) (string, error) {
	client := gorequest.New().Get(tokenUrlWithNoSession()).
		AppendHeader("Authorization", "Bearer "+token).
		Timeout(HttpTimeout * time.Second). //.
		Query("from=opsportal")

	resp, body, ierrors := client.End()
	if len(ierrors) != 0 {
		return "", ierrors[0]
	}

	if !HttpOK(resp.StatusCode) {
		return "", errors.Errorf("http code:%d body:%s", resp.StatusCode, body)
	}

	var lg TpaasTokenResp
	err := json.Unmarshal([]byte(body), &lg)
	if err != nil {
		fmt.Printf("jsnbody:%s", body)
		return "", errors.WithMessage(err, "token verification  response from tpaas is not json")
	}

	if !HttpOK(lg.Code) {
		return "", errors.Errorf("tpaas code:%d body:%s", lg.Code, body)
	}

	return lg.Data, nil
}

//// token verification for logined user
// return user info
func TokenWithSession(token string) (*OpsUser, error) {
	client := gorequest.New().Get(tokenUrlWithSession()).
		AppendHeader("Authorization", "Bearer "+token).
		Timeout(HttpTimeout * time.Second). //.
		Query("")

	resp, body, ierrors := client.End()
	if len(ierrors) != 0 {
		return nil, ierrors[0]
	}

	if !HttpOK(resp.StatusCode) {
		return nil, errors.Errorf("http code:%d body:%s", resp.StatusCode, body)
	}

	var lg TpaasTokenSessionResp
	err := json.Unmarshal([]byte(body), &lg)
	if err != nil {
		return nil, errors.WithMessage(err, "token verification  response from tpaas is not json")
	}

	if !HttpOK(lg.Code) {
		return nil, errors.Errorf("tpaas code:%d body:%s", lg.Code, body)
	}

	return &lg.Data.OpsUser, nil
}

func AuthorizeFromTpaasByApp(ns, app string, reource, action, token string) (bool, error) {

	if ns == "" || app == "" {
		return false, errors.Errorf("部门项目参数为空")
	}
	authdta, err := GetPermissionFromTpaas(ns, app, token)
	if err != nil {
		return false, err
	}

	myauthdata, ok := authdta[reource]
	if !ok {
		return false, nil
	}

	for _, a := range myauthdata {
		if a.Name == action && a.Enable == ICanAccess {
			return true, nil
		}
		if a.Name == action && a.Enable != ICanAccess {
			return false, nil
		}
	}

	return false, nil
}

func AuthorizeFromTpaasByNs(ns string, reource, action, token string) (bool, error) {

	if ns == "" {
		return false, errors.Errorf("部门参数为空")
	}
	authdta, err := GetPermissionFromTpaas(ns, "", token)
	if err != nil {
		return false, err
	}

	myauthdata, ok := authdta[reource]
	if !ok {
		return false, nil
	}

	for _, a := range myauthdata {
		if a.Name == action && a.Enable == ICanAccess {
			return true, nil
		}
		if a.Name == action && a.Enable != ICanAccess {
			return false, nil
		}
	}

	return false, nil
}

//
func GetPermissionFromTpaas(ns, app, token string) (map[string][]TpaasPermission, error) {

	if ns == "" {
		return nil, errors.New("部门参数为空")
	}
	var client *gorequest.SuperAgent
	if app == "" {
		client = gorequest.New().Get(authUrl4Ns(ns)).
			AppendHeader("Authorization", "Bearer "+token).
			Timeout(HttpTimeout * time.Second). //.
			Query("")

	} else {
		client = gorequest.New().Get(authUrl4App(ns, app)).
			AppendHeader("Authorization", "Bearer "+token).
			Timeout(HttpTimeout * time.Second). //.
			Query("from=opsportal")

	}

	resp, body, ierrors := client.End()
	if len(ierrors) != 0 {
		return nil, ierrors[0]
	}

	if !HttpOK(resp.StatusCode) {
		return nil, errors.Errorf("http code:%d body:%s", resp.StatusCode, body)
	}

	var lg TpaasPermissionResp
	err := json.Unmarshal([]byte(body), &lg)
	if err != nil {
		return nil, errors.WithMessage(err, "token verification  response from tpaas is not json")
	}

	if !HttpOK(lg.Code) {
		return nil, errors.Errorf("tpaas code:%d body:%s", lg.Code, body)
	}

	return lg.Data, nil
}
