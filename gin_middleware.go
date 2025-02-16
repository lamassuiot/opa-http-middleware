package opamiddleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lamassuiot/opa-http-middleware/config"
	"github.com/lamassuiot/opa-http-middleware/internal"
)

type GinInputCreationMethod func(c *gin.Context) (map[string]interface{}, error)

type GinMiddleware struct {
	Config *config.Config
	// InputCreationMethod is a function that returns the value to be sent to the OPA server.
	InputCreationMethod GinInputCreationMethod `json:"binding_method,omitempty"`
}

// NewGinMiddleware is the constructor for the opa gin middleware.
func NewGinMiddleware(cfg *config.Config, input GinInputCreationMethod) (*GinMiddleware, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	if input == nil {
		if cfg.InputCreationMethod == nil {
			return nil, errors.New("[opa-middleware-gin] InputCreationMethod must be provided")
		}
		input = func(c *gin.Context) (map[string]interface{}, error) {
			bind, err := cfg.InputCreationMethod(c.Request)
			if err != nil {
				return nil, err
			}
			return bind, nil
		}
	}
	return &GinMiddleware{
		Config:              cfg,
		InputCreationMethod: input,
	}, nil
}

// Use returns the handler for the middleware that is used by gin to evaluate the request against the policy.
func (g *GinMiddleware) Use() func(c *gin.Context) {
	return func(c *gin.Context) {
		if g.Config.Debug {
			g.Config.Logger.Printf("[opa-middleware-gin] Request received")
		}
		result, err := g.query(c)
		if err != nil {
			if g.Config.Debug {
				g.Config.Logger.Printf("[opa-middleware-gin] Error: %s", err.Error())
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if g.Config.Debug {
			g.Config.Logger.Printf("[opa-middleware-gin] Result: %t", result)
		}
		if result != g.Config.ExceptedResult {
			c.JSON(g.Config.DeniedStatusCode, gin.H{"error": g.Config.DeniedMessage})
			c.AbortWithStatus(g.Config.DeniedStatusCode)
			return
		}
		c.Next()
	}
}

func (g *GinMiddleware) query(c *gin.Context) (bool, error) {
	bind, err := g.InputCreationMethod(c)
	if err != nil {
		return !g.Config.ExceptedResult, err
	}
	if g.Config.URL != "" {
		input := make(map[string]interface{})
		input["input"] = bind
		return internal.QueryURL(c.Request, g.Config, input)
	}
	return internal.QueryPolicy(c.Request, g.Config, bind)
}
