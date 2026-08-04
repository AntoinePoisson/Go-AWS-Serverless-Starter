package health_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/handlertest"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/health"
)

func TestHandler(t *testing.T) {
	newHandler := func(*gomock.Controller) httpx.Handler {
		return health.New(&config.Config{Stage: "prod", Version: "1.2.3"})
	}

	handlertest.Run(t, []handlertest.Case{
		{
			Name:       "health reports the running stage and version",
			Path:       "/health",
			NewHandler: newHandler,
			WantStatus: http.StatusOK,
			Check: func(t *testing.T, res *http.Response) {
				var got struct {
					Status  string `json:"status"`
					Stage   string `json:"stage"`
					Version string `json:"version"`
				}
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, "ok", got.Status)
				assert.Equal(t, "prod", got.Stage)
				assert.Equal(t, "1.2.3", got.Version)
			},
		},
		{
			Name:       "health does not answer to writes",
			Method:     http.MethodPost,
			Path:       "/health",
			NewHandler: newHandler,
			WantStatus: http.StatusMethodNotAllowed,
		},
	})
}
