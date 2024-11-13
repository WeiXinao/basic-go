package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/WeiXinao/basic-go/webook/internal/integration/startup"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	"github.com/WeiXinao/basic-go/webook/ioc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string

		// 你要考虑准备数据。
		before func(t *testing.T)
		// 以及验证数据 数据库的数据对不对，你 Redis 的数据对不对
		after   func(t *testing.T)
		reqBody string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 不需要，也就是 Redis 什么数据也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
{
	"phone": "15212345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				// 这个手机号码，已经有一个验证码了
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_code:login:15212345678", "123456",
					time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位,没有被覆盖，还是123456
				assert.Equal(t, "123456", val)
			},
			reqBody: `
{
	"phone": "15212345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁，请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 这个手机号码，已经有一个验证码了，但是没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_code:login:15212345678", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 你要清理数据
				// "phone_code:%s:%s"
				val, err := rdb.GetDel(ctx, "phone_code:login:15212345678").Result()
				cancel()
				assert.NoError(t, err)
				// 你的验证码是 6 位,没有被覆盖，还是123456
				assert.Equal(t, "123456", val)
			},
			reqBody: `
{
	"phone": "15212345678"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},

		{
			name: "手机号码为空",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
{
	"phone": ""
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "输入有误",
			},
		},
		{
			name: "数据格式错误",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
{
	"phone": ,
}
`,
			wantCode: 400,
			//wantBody: web.Result{
			//	Code: 4,
			//	Msg:  "输入有误",
			//},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tc.after(t)
		})
	}
}

func TestUserHandler_e2e_LoginJWT(t *testing.T) {
	server := startup.InitWebServer()
	db := startup.InitDB()
	testCases := []struct {
		name         string
		before       func(t *testing.T, email string, password string)
		after        func(t *testing.T, email string, password string)
		email        string
		password     string
		readPassword string
		wantCode     int
		wantBody     web.Result
	}{
		{
			name: "登录成功",
			before: func(t *testing.T, email string, password string) {
				bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				require.NoError(t, err)
				err = db.Create(&dao.User{
					Email: sql.NullString{
						String: email,
						Valid:  true,
					},
					Password: string(bcryptPassword),
				}).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T, email string, password string) {
				err := db.Where("email = ?", email).Delete(&dao.User{}).Error
				require.NoError(t, err)
			},
			email:        "1632967698@qq.com",
			password:     "123456",
			readPassword: "123456",
			wantCode:     http.StatusOK,
			wantBody: web.Result{
				Code: 5,
				Msg:  "登录成功",
			},
		},
		{
			name: "用户名或密码不对",
			before: func(t *testing.T, email string, password string) {
				bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				require.NoError(t, err)
				err = db.Create(&dao.User{
					Email: sql.NullString{
						String: email,
						Valid:  true,
					},
					Password: string(bcryptPassword),
				}).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T, email string, password string) {
				err := db.Where("email = ?", email).Delete(&dao.User{}).Error
				require.NoError(t, err)
			},
			email:        "1632967698@qq.com",
			password:     "123456",
			readPassword: "123",
			wantCode:     http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "用户名或密码不对",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t, tc.email, tc.password)
			reqBody := web.LoginReq{
				Email:    tc.email,
				Password: tc.readPassword,
			}
			reqBytes, err := json.Marshal(reqBody)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/users/login",
				bytes.NewReader(reqBytes))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			assert.NoError(t, err)
			assert.Equal(t, webRes, tc.wantBody)
			tc.after(t, tc.email, tc.password)
		})
	}
}
