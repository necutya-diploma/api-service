package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	core "necutya/faker/internal/domain"
	"necutya/faker/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

const (
	contentTypeJSON = "application/json;charset=utf-8"

	accessTokenHeader = "Authorization"
	bearerSchema      = "Bearer"

	userIDHeaderName = "user-id"

	intType    = "int"
	stringType = "string"
)

// SendHTTPError sends empty HTTP response
func SendEmptyResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
}

// SendHTTPError marshals and sends HTTP response
func SendResponse(w http.ResponseWriter, statusCode int, respBody interface{}) {
	w.Header().Set("Content-Type", contentTypeJSON)

	binRespBody, err := json.Marshal(respBody)
	if err != nil {
		statusCode = http.StatusInternalServerError

		logger.Error(err)
	}

	w.WriteHeader(statusCode)

	_, err = w.Write(binRespBody)
	if err != nil {
		logger.Error(err)
	}
}

// SendHTTPError sends HTTP error with code.
func SendHTTPError(w http.ResponseWriter, err error) {
	type HttError struct {
		StatusCode int             `json:"-"`
		Code       string          `json:"code,omitempty"`
		Message    string          `json:"message,omitempty"`
		Errors     []core.ApiError `json:"validation_errors,omitempty"`
	}

	httpErr := HttError{}

	switch err {
	case core.ErrNotFound:
		httpErr.StatusCode = http.StatusNotFound
		httpErr.Code = "not_found"

	case core.ErrExpiredSession, core.ErrInvalidSession:
		httpErr.StatusCode = http.StatusUnauthorized
		httpErr.Code = "unauthorized"

	case core.ErrUnconfirmedEmail, core.ErrInvalidLoginOrPassword:
		httpErr.StatusCode = http.StatusUnauthorized
		httpErr.Code = "unauthorized"
		httpErr.Message = err.Error()

	case core.ErrRequestLimit:
		httpErr.StatusCode = http.StatusForbidden
		httpErr.Code = "requests limit"
		httpErr.Message = err.Error()

	case core.ErrAlreadyExist, core.ErrThisPlanAlreadySet:
		httpErr.StatusCode = http.StatusConflict
		httpErr.Code = "conflict"
		httpErr.Message = err.Error()

	case core.ErrUnknownCallbackType,
		core.ErrTransactionInvalid,
		core.ErrInvalidCode,
		core.ErrExpiredCode,
		core.ErrInvalidCurrentPassword:
		httpErr.StatusCode = http.StatusBadRequest
		httpErr.Code = "bad_request"
		httpErr.Message = err.Error()

	default:
		switch v := err.(type) {
		case validator.ValidationErrors:
			apiErrors := make([]core.ApiError, len(v))
			for i, fe := range v {
				apiErrors[i] = core.ApiError{
					Field: fe.Field(),
					Msg:   msgForTag(fe.Tag(), fe.Param()),
				}
			}
			httpErr.StatusCode = http.StatusBadRequest
			httpErr.Errors = apiErrors
			httpErr.Code = "bad_request"
		default:
			httpErr.StatusCode = http.StatusInternalServerError
			httpErr.Code = "internal"
		}
	}

	logger.Error(err)

	SendResponse(w, httpErr.StatusCode, httpErr)
}

// unmarshalRequest - unmarshal http request to provided structure.
func UnmarshalRequest(r *http.Request, body interface{}) error {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// todo custom error
		return err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(reqBody, body); err != nil {
		if e, ok := err.(*json.UnmarshalTypeError); ok {
			return e
			// todo custom error
			// return models.BadRequest{
			// 	Msg: err.Error(),
			// 	Errors: []models.FieldError{{
			// 		Field: strings.ToLower(e.Field),
			// 		Code:  "validation_is_" + rxNotAlphanumeric.ReplaceAllString(e.Type.String(), ""),
			// 	}},
			// }
		}
		// todo custom error
		return err
	}

	return nil
}

func GetUserClient(r *http.Request) string {
	return r.UserAgent()
}

func GetUserIPAddress(r *http.Request) string {
	ipAddress := r.Header.Get("X-Real-Ip")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	}

	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	if ipAddress != "" {
		return strings.Split(ipAddress, ":")[0]
	}
	return ""
}

func GetAuthorizationHeader(r *http.Request) string {
	return r.Header.Get(accessTokenHeader)
}

func SetUserIDHeader(r *http.Request, userID string) {
	r.Header.Set(userIDHeaderName, userID)
}

func ParseAuthorizationHeader(header, schema string) (string, error) {
	authSlice := strings.Split(strings.TrimSpace(header), " ")
	if len(authSlice) != 2 || !strings.EqualFold(authSlice[0], schema) {
		return "", errors.New("invalid authorization header") // nolint: goerr113
	}

	return authSlice[1], nil
}

func ParseBearerAuthorizationHeader(header string) (string, error) {
	return ParseAuthorizationHeader(header, bearerSchema)
}

func GetPathVar(r *http.Request, varName string, varType string) (interface{}, error) {
	if valueStr, ok := mux.Vars(r)[varName]; ok {
		switch varType {
		case intType:
			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return nil, err
			}

			return value, nil

		case stringType:
			return valueStr, nil

		default:
			return valueStr, nil
		}
	}

	return nil, errors.New("can`t get path variable")
}

func msgForTag(tag, tagParam string) string {
	switch tag {
	case "required":
		return "field is required"
	case "email":
		return "invalid email"
	case "max":
		return "value is too bigger"
	case "min":
		return "value is too lower"
	case "nefield":
		return fmt.Sprintf("value must not be equal to value of %s", tagParam)
	case "eqfield":
		return fmt.Sprintf("value must be equal to value of %s", tagParam)
	default:
		return "invalid value"
	}
}
