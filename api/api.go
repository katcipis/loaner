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

// BorrowerPayment is part of the CreateLoanPlanResponse
type BorrowerPayment struct {
	Date                          string `json:"date"`
	PaymentAmount                 string `json:"borrowerPaymentAmount"`
	Interest                      string `json:"interest"`
	Principal                     string `json:"principal"`
	InitialOutstandingPrincipal   string `json:"initialOutstandingPrincipal"`
	RemainingOutstandingPrincipal string `json:"remainingOutstandingPrincipal"`
}

// CreateLoanPlanResponse is the response of the create loan plan request
type CreateLoanPlanResponse struct {
	BorrowerPayments []BorrowerPayment `json:"borrowerPayments"`
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
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			msg := fmt.Sprintf("method %q is not allowed", req.Method)
			logResponseBodyWrite(logger, res, newErrorResponse(logger, msg))
			logger.WithFields(log.Fields{"error": msg}).Warning("method not allowed")
			return
		}
		dec := json.NewDecoder(req.Body)
		parsedReq := CreateLoanPlanRequest{}

		err := dec.Decode(&parsedReq)
		if err != nil {
			msg := fmt.Sprintf("cant parse request body as JSON:%v", err)
			res.WriteHeader(http.StatusBadRequest)
			logResponseBodyWrite(logger, res, newErrorResponse(logger, msg))
			logger.WithFields(log.Fields{"error": msg}).Warning("invalid request body")
			return
		}

		loanAmount, err := decimal.NewFromString(parsedReq.LoanAmount)
		if err != nil {
			handleFieldParsingError(logger, res, "loanAmount", err)
			return
		}

		annualInterestRate, err := decimal.NewFromString(parsedReq.NominalRate)
		if err != nil {
			handleFieldParsingError(logger, res, "nominalRate", err)
			return
		}

		startDate, err := time.Parse(dateLayout, parsedReq.StartDate)
		if err != nil {
			handleFieldParsingError(logger, res, "startDate", err)
			return
		}

		payments, err := createLoanPlan(loanAmount, annualInterestRate, parsedReq.Duration, startDate)
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
				logResponseBodyWrite(logger, res, newErrorResponse(logger, err.Error()))
				logger.WithError(err).Warning("bad request error")
				return
			}
			// Specially when you can't give much detail on errors for
			// security reasons it would be a good idea to have
			// a tracing id for errors to help map the error to the logs.
			res.WriteHeader(http.StatusInternalServerError)
			logResponseBodyWrite(logger, res, newErrorResponse(logger, "internal server error"))
			logger.WithError(err).Error("internal server error")
			return
		}

		resp := CreateLoanPlanResponse{
			BorrowerPayments: toBorrowerPayments(payments),
		}
		res.WriteHeader(http.StatusOK)
		logResponseBodyWrite(logger, res, toJSON(logger, resp))
	})
	return mux
}

const (
	dateLayout = time.RFC3339
)

func toBorrowerPayments(payments []loan.Payment) []BorrowerPayment {
	res := make([]BorrowerPayment, len(payments))
	for i, p := range payments {
		res[i] = BorrowerPayment{
			Date:                          p.Date.Format(dateLayout),
			PaymentAmount:                 p.PaymentAmount.String(),
			Interest:                      p.Interest.String(),
			Principal:                     p.Principal.String(),
			InitialOutstandingPrincipal:   p.InitialOutstandingPrincipal.String(),
			RemainingOutstandingPrincipal: p.RemainingOutstandingPrincipal.String(),
		}
	}
	return res
}

func logResponseBodyWrite(logger *log.Entry, w io.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Warning("writing response body")
	}
}

func newErrorResponse(logger *log.Entry, message string) []byte {
	return toJSON(logger, ErrorResponse{
		Error: Error{Message: message},
	})
}

func toJSON(logger *log.Entry, v interface{}) []byte {
	res, err := json.Marshal(v)
	if err != nil {
		logger.WithError(err).Warning("unable to marshal as JSON")
	}
	return res
}

func handleFieldParsingError(logger *log.Entry, res http.ResponseWriter, fieldName string, err error) {
	msg := fmt.Sprintf("can't parse %q from request:%v", fieldName, err)
	res.WriteHeader(http.StatusBadRequest)
	logResponseBodyWrite(logger, res, newErrorResponse(logger, msg))
	logger.WithError(err).WithFields(log.Fields{"field": fieldName}).Warning("invalid field on request")
}
