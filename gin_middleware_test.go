package opamiddleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lamassuiot/opa-http-middleware/config"
)

var Test_Policy = `
package policy

default allow = false

allow {
	input.path = "/api/v1/users"
	input.method = "GET"
}`

func TestGinMiddleware_Query(t *testing.T) {
	type fields struct {
		Config              *config.Config
		InputCreationMethod GinInputCreationMethod
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test GinMiddleware_Query",
			fields: fields{
				Config: &config.Config{
					Policy:           Test_Policy,
					Query:            "data.policy.allow",
					ExceptedResult:   true,
					DeniedStatusCode: 403,
					DeniedMessage:    "Forbidden",
				},
				InputCreationMethod: func(c *gin.Context) (map[string]interface{}, error) {
					return map[string]interface{}{
						"path":   c.Request.URL.Path,
						"method": c.Request.Method,
					}, nil
				},
			},
			args: args{
				req: &http.Request{
					URL: &url.URL{
						Path: "/api/v1/users",
					},
					Method: "GET",
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := gin.New()
			h := &GinMiddleware{
				Config:              tt.fields.Config,
				InputCreationMethod: tt.fields.InputCreationMethod,
			}

			c := gin.CreateTestContextOnly(httptest.NewRecorder(), e)
			c.Request = tt.args.req

			got, err := h.query(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Query() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGinMiddleware_Use(t *testing.T) {
	type fields struct {
		Config              *config.Config
		InputCreationMethod GinInputCreationMethod
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Test GinMiddleware_Use",
			fields: fields{
				Config: &config.Config{
					Policy:           Test_Policy,
					Query:            "data.policy.allow",
					ExceptedResult:   true,
					DeniedStatusCode: 403,
					DeniedMessage:    "Forbidden",
				},
				InputCreationMethod: func(c *gin.Context) (map[string]interface{}, error) {
					return map[string]interface{}{
						"path":   c.Request.URL.Path,
						"method": c.Request.Method,
					}, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &GinMiddleware{
				Config:              tt.fields.Config,
				InputCreationMethod: tt.fields.InputCreationMethod,
			}
			h.Use()
		})
	}
}

func TestNewGinMiddleware(t *testing.T) {
	type args struct {
		cfg                 *config.Config
		inputCreationMethod GinInputCreationMethod
	}
	tests := []struct {
		name    string
		args    args
		want    *GinMiddleware
		wantErr bool
	}{
		{
			name: "Test NewGinMiddleware",
			args: args{
				cfg: &config.Config{
					Policy:           "policy",
					Query:            "data.query",
					ExceptedResult:   true,
					DeniedStatusCode: 403,
					DeniedMessage:    "Forbidden",
				},
				inputCreationMethod: func(c *gin.Context) (map[string]interface{}, error) {
					return map[string]interface{}{
						"path":   c.Request.URL.Path,
						"method": c.Request.Method,
					}, nil
				},
			},
			want: &GinMiddleware{
				Config: &config.Config{
					Policy:           "policy",
					Query:            "data.query",
					ExceptedResult:   true,
					DeniedStatusCode: 403,
					DeniedMessage:    "Forbidden",
				},
				InputCreationMethod: func(c *gin.Context) (map[string]interface{}, error) {
					return map[string]interface{}{
						"path":   c.Request.URL.Path,
						"method": c.Request.Method,
					}, nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGinMiddleware(tt.args.cfg, tt.args.inputCreationMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGinMiddleware() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
