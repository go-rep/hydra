package hydra

import (
	"fmt"
	"testing"

	"github.com/micro-plat/hydra/conf/server/limiter"
	"github.com/micro-plat/hydra/global"
	"github.com/micro-plat/hydra/hydra/servers/http"
	"github.com/micro-plat/hydra/hydra/servers/pkg/middleware"
	"github.com/micro-plat/hydra/test/assert"
	"github.com/micro-plat/hydra/test/mocks"
)

//author:liujinyin
//time:2020-10-21 10:00
//desc:测试限流中间件逻辑
func TestLimit(t *testing.T) {

	type testCase struct {
		name        string
		requestPath string
		opts        []limiter.Option
		wantStatus  int
		wantContent string
		wantSpecial string
	}
	/*

	 */

	tests := []*testCase{
		{
			name:        "限流-未启用-未配置",
			requestPath: "/limiter",
			opts:        []limiter.Option{},
			wantStatus:  200,
			wantContent: "",
			wantSpecial: "",
		},
		{
			name:        "限流-未启用-Disable=true",
			requestPath: "/limiter",
			opts: []limiter.Option{
				limiter.WithDisable(),
			},
			wantStatus:  200,
			wantContent: "",
			wantSpecial: "",
		},
		{
			name:        "限流-启用-不在限流配置内",
			requestPath: "/limiter-notin",
			opts: []limiter.Option{
				limiter.WithEnable(),
				limiter.WithRuleList(&limiter.Rule{
					Path:     "/limiter",
					Action:   []string{"GET"},
					MaxAllow: 1,
					MaxWait:  0,
					Fallback: false,
					Resp: &limiter.Resp{
						Status:  510,
						Content: "fallback",
					},
				}),
			},
			wantStatus:  200,
			wantContent: "",
			wantSpecial: "",
		},
		{
			name:        "限流-启用-在限流配置内-不延迟",
			requestPath: "/limiter",
			opts: []limiter.Option{
				limiter.WithEnable(),
				limiter.WithRuleList(&limiter.Rule{
					Path:     "/limiter",
					Action:   []string{"GET"},
					MaxAllow: 1,
					MaxWait:  0,
					Fallback: false,
					Resp: &limiter.Resp{
						Status:  510,
						Content: "fallback",
					},
				}),
			},
			wantStatus:  200,
			wantContent: "",
			wantSpecial: "",
		},
		{
			name:        "限流-启用-在限流配置内-延迟",
			requestPath: "/limiter",
			opts: []limiter.Option{
				limiter.WithEnable(),
				limiter.WithRuleList(&limiter.Rule{
					Path:     "/limiter",
					Action:   []string{"GET"},
					MaxAllow: 0,
					MaxWait:  10,
					Fallback: false,
					Resp: &limiter.Resp{
						Status:  510,
						Content: "fallback",
					},
				}),
			},
			wantStatus:  510,
			wantContent: "fallback",
			wantSpecial: "limit",
		},
	}
	for _, tt := range tests {
		global.Def.ServerTypes = []string{http.API}
		fmt.Println("---------------------------", tt.name)

		mockConf := mocks.NewConf()
		mockConf.API(":51001")
		//mockConf.Service.API.Add()
		//初始化测试用例参数
		mockConf.GetAPI().Limit(tt.opts...)
		serverConf := mockConf.GetAPIConf()

		ctx := &mocks.MiddleContext{
			MockTFuncs: map[string]interface{}{},
			MockRequest: &mocks.MockRequest{
				MockPath: &mocks.MockPath{
					MockRequestPath: tt.requestPath,
				},
			},
			MockResponse:   &mocks.MockResponse{MockStatus: 200},
			MockServerConf: serverConf,
		}

		//获取中间件
		handler := middleware.Limit()

		//调用中间件
		handler(ctx)

		//断言结果
		gotStatus, gotContent := ctx.Response().GetFinalResponse()
		gotSpecial := ctx.Response().GetSpecials()

		assert.Equalf(t, tt.wantStatus, gotStatus, tt.name, tt.wantStatus, gotStatus)
		assert.Equalf(t, tt.wantContent, gotContent, tt.name, tt.wantContent, gotContent)
		assert.Equalf(t, tt.wantSpecial, gotSpecial, tt.name, tt.wantSpecial, gotSpecial)

	}
}
