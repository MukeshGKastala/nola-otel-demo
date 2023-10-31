// Package calculatorv1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package calculatorv1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/oapi-codegen/runtime"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CalculationResponse defines model for CalculationResponse.
type CalculationResponse struct {
	Completed  time.Time          `json:"completed"`
	Created    time.Time          `json:"created"`
	Expression string             `json:"expression"`
	Id         openapi_types.UUID `json:"id"`
	Result     float64            `json:"result"`
	Student    string             `json:"student"`
}

// CreateCalculationRequest defines model for CreateCalculationRequest.
type CreateCalculationRequest struct {
	Expression string `json:"expression"`
	Student    string `json:"student"`
}

// CreateCalculationResponse defines model for CreateCalculationResponse.
type CreateCalculationResponse struct {
	Id openapi_types.UUID `json:"id"`
}

// Error defines model for Error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DefaultError defines model for DefaultError.
type DefaultError = Error

// NotFound defines model for NotFound.
type NotFound = Error

// CreateCalculationJSONRequestBody defines body for CreateCalculation for application/json ContentType.
type CreateCalculationJSONRequestBody = CreateCalculationRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /calculations)
	CreateCalculation(w http.ResponseWriter, r *http.Request)

	// (GET /calculations/{uuid})
	GetCalculation(w http.ResponseWriter, r *http.Request, uuid openapi_types.UUID)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// CreateCalculation operation middleware
func (siw *ServerInterfaceWrapper) CreateCalculation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateCalculation(w, r)
	}))

	for i := len(siw.HandlerMiddlewares) - 1; i >= 0; i-- {
		handler = siw.HandlerMiddlewares[i](handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetCalculation operation middleware
func (siw *ServerInterfaceWrapper) GetCalculation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "uuid" -------------
	var uuid openapi_types.UUID

	err = runtime.BindStyledParameter("simple", false, "uuid", mux.Vars(r)["uuid"], &uuid)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "uuid", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetCalculation(w, r, uuid)
	}))

	for i := len(siw.HandlerMiddlewares) - 1; i >= 0; i-- {
		handler = siw.HandlerMiddlewares[i](handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{})
}

type GorillaServerOptions struct {
	BaseURL          string
	BaseRouter       *mux.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r *mux.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options GorillaServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = mux.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.HandleFunc(options.BaseURL+"/calculations", wrapper.CreateCalculation).Methods("POST")

	r.HandleFunc(options.BaseURL+"/calculations/{uuid}", wrapper.GetCalculation).Methods("GET")

	return r
}

type DefaultErrorJSONResponse Error

type NotFoundJSONResponse Error

type CreateCalculationRequestObject struct {
	Body *CreateCalculationJSONRequestBody
}

type CreateCalculationResponseObject interface {
	VisitCreateCalculationResponse(w http.ResponseWriter) error
}

type CreateCalculation200JSONResponse CreateCalculationResponse

func (response CreateCalculation200JSONResponse) VisitCreateCalculationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type CreateCalculationdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response CreateCalculationdefaultJSONResponse) VisitCreateCalculationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetCalculationRequestObject struct {
	Uuid openapi_types.UUID `json:"uuid"`
}

type GetCalculationResponseObject interface {
	VisitGetCalculationResponse(w http.ResponseWriter) error
}

type GetCalculation200JSONResponse CalculationResponse

func (response GetCalculation200JSONResponse) VisitGetCalculationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetCalculation404JSONResponse struct{ NotFoundJSONResponse }

func (response GetCalculation404JSONResponse) VisitGetCalculationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response)
}

type GetCalculationdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response GetCalculationdefaultJSONResponse) VisitGetCalculationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (POST /calculations)
	CreateCalculation(ctx context.Context, request CreateCalculationRequestObject) (CreateCalculationResponseObject, error)

	// (GET /calculations/{uuid})
	GetCalculation(ctx context.Context, request GetCalculationRequestObject) (GetCalculationResponseObject, error)
}

type StrictHandlerFunc = strictnethttp.StrictHttpHandlerFunc
type StrictMiddlewareFunc = strictnethttp.StrictHttpMiddlewareFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// CreateCalculation operation middleware
func (sh *strictHandler) CreateCalculation(w http.ResponseWriter, r *http.Request) {
	var request CreateCalculationRequestObject

	var body CreateCalculationJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.CreateCalculation(ctx, request.(CreateCalculationRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "CreateCalculation")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(CreateCalculationResponseObject); ok {
		if err := validResponse.VisitCreateCalculationResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// GetCalculation operation middleware
func (sh *strictHandler) GetCalculation(w http.ResponseWriter, r *http.Request, uuid openapi_types.UUID) {
	var request GetCalculationRequestObject

	request.Uuid = uuid

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetCalculation(ctx, request.(GetCalculationRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetCalculation")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetCalculationResponseObject); ok {
		if err := validResponse.VisitGetCalculationResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}
