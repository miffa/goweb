package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockFuc(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/login/db":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
    "code": 200,
    "message": "success",
    "status": true,
    "data": {
        "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJsIiwiaWF0IjoxNTkxMDg3NTIyLCJpc3MiOiJ0cGFhcyJ9.dyx1zJX2ytC3NsWmTawMCuuUJA7R7-DN4dSCyFhwR3Avm5MpTIhMkiDVt1iWdwGg9a_WOjITlgsw2qUUjmhZaA"
    }
	}`))
	case "/api/v1/logout":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
    "code": 200,
    "message": "success",
    "status": true
		  }`))

	case "/api/v1/current/token":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
		    "code": 200,
			"message": "success",
			"status": true ,
			"data": "l"
		  }`))

	case "/api/v1/current/user":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
         "code": 200,
         "message": "success",
         "status": true,
         "data": {
             "id": 5,
             "name": "l",
             "password": "de68641f426e2363684a493cc37a4d6ea14d2bf7ccfb45e813943d24b3fc4d93d54804684b99efd67ad557ed1b71171ce351",
             "email": "l@troila.com",
             "admin": true,
             "lastLogin": "2020-06-02T10:39:37+08:00",
             "lastIp": "172.26.133.163",
             "createTime": "2020-04-23T11:20:26+08:00"
         }

	  }`))

	case "/api/v1/current/authenticate/test1/test11":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
    "code": 200,
    "message": "success",
    "status": true,
    "data": {
        "ALERTRULES": [
            {
                "id": 301,
                "name": "CREATE",
                "comment": "创建",
                "enable": 1
            },
            {
                "id": 302,
                "name": "DELETE",
                "comment": "删除",
                "enable": 1
            },
            {
                "id": 303,
                "name": "UPDATE",
                "comment": "修改",
                "enable": 1
            }
        ],
        "APP": [
            {
                "id": 5,
                "name": "CREATE",
                "comment": "创建",
                "enable": 1
            },
            {
                "id": 6,
                "name": "UPDATE",
                "comment": "修改",
                "enable": 1
            },
            {
                "id": 7,
                "name": "READ",
                "comment": "查看",
                "enable": 1
            },
            {
                "id": 8,
                "name": "DELETE",
                "comment": "删除",
                "enable": 1
            }
        ]
    }

 }`))
	case "/api/v1/current/authenticate/test1":
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{
    "code": 200,
    "message": "success",
    "status": true,
    "data": {
        "ALERTRULES": [
            {
                "id": 301,
                "name": "CREATE",
                "comment": "创建",
                "enable": 1
            },
            {
                "id": 302,
                "name": "DELETE",
                "comment": "删除",
                "enable": 1
            },
            {
                "id": 303,
                "name": "UPDATE",
                "comment": "修改",
                "enable": 1
            }
        ],
        "APP": [
            {
                "id": 5,
                "name": "CREATE",
                "comment": "创建",
                "enable": 1
            },
            {
                "id": 6,
                "name": "UPDATE",
                "comment": "修改",
                "enable": 1
            },
            {
                "id": 7,
                "name": "READ",
                "comment": "查看",
                "enable": 1
            },
            {
                "id": 8,
                "name": "DELETE",
                "comment": "删除",
                "enable": 1
            }
        ]
    }

 }`))

	default:
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{ 
            "code": 404,
            "message": "no data found",
            "status": false
		}`))
	}
}

func TestToken(t *testing.T) {

	ts := httptest.NewServer(
		http.HandlerFunc(MockFuc),
	)
	defer ts.Close()
	assert := assert.New(t)

	Init([]string{ts.URL})

	tokenstr, err := Login("l", "1qaz2wsx")

	assert.NoError(err)
	if err != nil {
		t.Errorf("login err:%v", err)
		return
	}

	t.Logf("token:%s", tokenstr)

	user, err := TokenNoSession(tokenstr)

	assert.NoError(err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("token user:%s", user)

	userinfo, err := TokenWithSession(tokenstr)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	assert.NoError(err)

	t.Logf("token user info:%v", *userinfo)

	qx, err := GetPermissionFromTpaas("test1", "test11", tokenstr)

	assert.NoError(err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("access data:%v", qx)
	canigo, err := AuthorizeFromTpaasByNs("test1", "ALERTRULES", "CREATE", tokenstr)

	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	assert.NoError(err)
	assert.Equal(true, canigo)

	t.Logf("access data: test1 test11 ALERTRULES CREATE :%v", canigo)

	canigo, err = AuthorizeFromTpaasByApp("test1", "test11", "ALERTRULES", "CREATE", tokenstr)

	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	assert.NoError(err)
	assert.Equal(true, canigo)

	t.Logf("access data: test1 test11 ALERTRULES CREATE :%v", canigo)
}
