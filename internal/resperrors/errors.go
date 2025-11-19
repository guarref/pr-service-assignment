package resperrors

import (
	"fmt"
	"net/http"
)

type RespError struct {
	Code       string 
	Message    string 
	StatusCode int   
}

func NewRespError(code, message string, status int) *RespError {
	return &RespError{
		Code:       code,
		Message:    message,
		StatusCode: status,
	}
}

func (e *RespError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

var (
	//409
	ErrTeamExists = &RespError{						
		Code:       "TEAM_EXISTS",
		Message:    "team_name already exists",
		StatusCode: http.StatusConflict, 
	}

	ErrPRExists = &RespError{
		Code:       "PR_EXISTS",
		Message:    "PR id already exists",
		StatusCode: http.StatusConflict,
	}

	ErrPRMerged = &RespError{
		Code:       "PR_MERGED",
		Message:    "cannot reassign on merged PR",
		StatusCode: http.StatusConflict,
	}

	ErrNotAssigned = &RespError{
		Code:       "NOT_ASSIGNED",
		Message:    "reviewer is not assigned to this PR",
		StatusCode: http.StatusConflict,
	}

	ErrNoCandidate = &RespError{
		Code:       "NO_CANDIDATE",
		Message:    "no active replacement candidate in team",
		StatusCode: http.StatusConflict,
	}

	//404
	ErrNotFound = &RespError{					
		Code:       "NOT_FOUND",
		Message:    "resource not found",
		StatusCode: http.StatusNotFound, 
	}

	ErrTeamNotFound = &RespError{
		Code:       "NOT_FOUND",
		Message:    "team not found",
		StatusCode: http.StatusNotFound,
	}

	ErrUserNotFound = &RespError{
		Code:       "NOT_FOUND",
		Message:    "user not found",
		StatusCode: http.StatusNotFound,
	}

	ErrPRNotFound = &RespError{
		Code:       "NOT_FOUND",
		Message:    "pull request not found",
		StatusCode: http.StatusNotFound,
	}

	//400
	ErrBadRequest = &RespError{						
		Code:       "BAD_REQUEST",
		Message:    "invalid request body",
		StatusCode: http.StatusBadRequest, 
	}

	ErrInvalidJSON = &RespError{
		Code:       "BAD_REQUEST",
		Message:    "invalid JSON format",
		StatusCode: http.StatusBadRequest,
	}
)