// Package api is responsible for exporting services
// as an HTTP API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/katcipis/loaner/loan"
)

// CreateLoanPlanRequest is the request body required to create loan plans
type CreateLoanPlanRequest struct {
	LoanAmount  string `json:"loanAmount"`
	NominalRate string `json:"nominalRate"`
	Duration    int    `json:"duration"`
	StartDate   string `json:"startDate"`
}

// CreateLoanPlanResponse is the response of the create loan plan request
type CreateLoanPlanResponse struct {
	FullName string `json:"fullname"`
}

// Error contains error information used in error responses
type Error struct {
	Message string `json:"message"`
}

// ErrorResponse represents the response body
// of all requests that failed.
type ErrorResponse struct {
	Error Error `json:"error"`
}

// LoanPlanCreator is a function that given the loan parameters
// will create a loan plan in the form of a list of payments.
type LoanPlanCreator func(
	totalLoanAmount decimal.Decimal,
	annualInterestRate decimal.Decimal,
	durationInMonths int,
	start time.Time,
) ([]loan.Payment, error)

const (
	CreateLoanPlanPath = "/loan-plan"
)

// New creates a new HTTP handler with all the service routes.
func New(createLoanPlan LoanPlanCreator) http.Handler {

	mux := http.NewServeMux()
	logger := log.WithFields(log.Fields{"path": CreateLoanPlanPath})

	mux.HandleFunc(CreateLoanPlanPath, func(res http.ResponseWriter, req *http.Request) {
		// TODO: test wrong method
		//if req.Method != http.MethodPost {
		//res.WriteHeader(http.StatusMethodNotAllowed)
		//msg := fmt.Sprintf("method %q is not allowed", req.Method)
		//logResponseBodyWrite(logger, res, errorResponse(msg))
		//logger.WithFields(log.Fields{"error": msg}).Warning("method not allowed")
		//return
		//}
		dec := json.NewDecoder(req.Body)
		parsedReq := CreateLoanPlanRequest{}

		err := dec.Decode(&parsedReq)
		if err != nil {
			msg := fmt.Sprintf("error parsing JSON request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			logResponseBodyWrite(logger, res, errorResponse(msg))
			logger.WithFields(log.Fields{"error": msg}).Warning("invalid request body")
			return
		}

		// TODO: check parameters are properly passed
		_, err = createLoanPlan(decimal.Zero, decimal.Zero, 0, time.Time{})
		if err != nil {
			if errors.Is(err, loan.ErrInvalidParameter) {
				res.WriteHeader(http.StatusBadRequest)
				// Invalid params errors are guaranteed
				// to be safe to send to users in this case
				// (not much info added on the error context).
				// If a service is external care must be taken to not leak details
				// that can be a potential security threat.
				// When that is not the case I like the idea of
				// informative error responses as detailed here:
				//
				// - https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
				//
				// I'm specially fond to the idea of a cross service
				// operational trace (instead of stack traces).
				// But I never tried it yet :-).
				logResponseBodyWrite(logger, res, errorResponse(err.Error()))
				logger.WithFields(log.Fields{"error": err.Error()}).Warning("bad request error")
				return
			}
			//// Specially when you can't give much detail on errors for
			//// security reasons it would be a good idea to have
			//// a tracing id for errors to help map the error to the logs.
			//res.WriteHeader(http.StatusInternalServerError)
			//logResponseBodyWrite(logger, res, errorResponse("internal server error"))
			//logger.WithFields(log.Fields{"error": err.Error()}).Error("internal server error")
			//return
			// TODO: test internal unknown errors
		}

		res.WriteHeader(http.StatusCreated)
		logResponseBodyWrite(logger, res, jsonResponse(CreateLoanPlanResponse{}))
	})
	return mux
}

func logResponseBodyWrite(logger *log.Entry, w io.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Warning("writing response body")
	}
}

func errorResponse(message string) []byte {
	return jsonResponse(ErrorResponse{
		Error: Error{Message: message},
	})
}

func jsonResponse(v interface{}) []byte {
	// TODO: handle and log err
	res, _ := json.Marshal(v)
	return res
}
