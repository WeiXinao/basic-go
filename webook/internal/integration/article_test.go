package integration

import (
	"bytes"
	"encoding/json"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/integration/startup"
	dao "github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	ijwt "github.com/WeiXinao/basic-go/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	// 在所有测试执行之前，初始化一些内容
	//s.server = startup.InitWebServer()
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	s.db = startup.InitDB()
	artHdl := startup.InitArticleHandler(dao.NewAuthorDAO(startup.InitDB()))
	// 注册好了路由
	artHdl.RegisterRoutes(s.server)
}

func (s *ArticleTestSuite) TearDownTest() {
	// 请空所有数据并且ziz
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestPublish() {

}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)

		// 预期中的输入
		art Article

		// HTTP 相应码
		wantCode int
		// 我希望 HTTP 响应，带上帖子的 ID
		wantRes Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				//	验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func(t *testing.T) {
				// 提前准备数据
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					// 跟时间有关的测试，不是逼不得已，不要用 time.Now()
					// 因为 time.Now() 每次运行都不同，你很难断言
					Ctime: 123,
					Utime: 234,
					// 假设这个是一个已经发表的，然后你去修改，改成了没发表
					Status: domain.ArticleStatusUnpublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				//	验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				// 是为了确保我更新了 Utime
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				// 提前准备数据
				err := s.db.Create(dao.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					// 测试模拟的用户是 123，这里是 789
					// 意味着你在修改别人的数据
					AuthorId: 789,
					// 跟时间有关的测试，不是逼不得已，不要用 time.Now()
					// 因为 time.Now() 每次运行都不同，你很难断言
					Ctime: 123,
					Utime: 234,
					// 为了验证状态没有变
					Status: domain.ArticleStatusPublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				//	验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer([]byte(reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			s.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello，这是测试套件")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
