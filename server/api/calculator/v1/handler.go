package calculatorv1

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func MakeHTTPHandler(si StrictServerInterface) http.Handler {
	mux := mux.NewRouter()
	mux.Use(otelmux.Middleware("otel-test"))
	return HandlerFromMux(NewStrictHandler(si, nil), mux)
}
