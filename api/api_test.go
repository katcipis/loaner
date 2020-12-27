package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/katcipis/loaner/api"
	"github.com/katcipis/loaner/loan"
	"github.com/shopspring/decimal"
)

func TestLoanPlanCreation(t *testing.T) {
	type Test struct {
		name           string
		requestBody    []byte
		method         string
		injectResponse []loan.Payment
		injectErr      error
		wantStatusCode int
		want           api.CreateLoanPlanResponse
	}

	tests := []Test{
		{
			name:           "MethodNotAllowedForGet",
			method:         "GET",
			requestBody:    validCreateLoanRequestBody(t),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "BadRequestIfParametersAreConsideredInvalidByLoanPlanCreator",
			requestBody:    validCreateLoanRequestBody(t),
			injectErr:      loan.ErrInvalidParameter,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "BadRequestIfRequestBodyIsEmpty",
			requestBody:    []byte{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "BadRequestIfRequestLoanAmountIsNotDecimal",
			requestBody: toJSON(t, api.CreateLoanPlanRequest{
				LoanAmount:  "notADecimal",
				NominalRate: "5.0",
				Duration:    1,
				StartDate:   "2020-12-01T00:00:00Z",
			}),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "BadRequestIfRequestNominalRateIsNotDecimal",
			requestBody: toJSON(t, api.CreateLoanPlanRequest{
				LoanAmount:  "1.00",
				NominalRate: "wrongValue",
				Duration:    1,
				StartDate:   "2020-12-01T00:00:00Z",
			}),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "BadRequestIfRequestStartDateIsNotValidDate",
			requestBody: toJSON(t, api.CreateLoanPlanRequest{
				LoanAmount:  "1.00",
				NominalRate: "1.0",
				Duration:    1,
				StartDate:   "notDate",
			}),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "BadRequestIfRequestBodyIsNotValidJSON",
			requestBody:    []byte("{notvalidjson]"),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "InternalServerErrorOnLoanCalculationError",
			requestBody:    validCreateLoanRequestBody(t),
			injectErr:      errors.New("injected generic error"),
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "SuccessBuildingLoanPlan",
			requestBody: validCreateLoanRequestBody(t),
			injectResponse: []loan.Payment{
				{
					Date:                          parseTime(t, "2018-01-01T00:00:00Z"),
					PaymentAmount:                 parseDecimal(t, "1001.25"),
					Interest:                      parseDecimal(t, "1.67"),
					Principal:                     parseDecimal(t, "999.58"),
					InitialOutstandingPrincipal:   parseDecimal(t, "2000"),
					RemainingOutstandingPrincipal: parseDecimal(t, "1000.42"),
				},
				{
					Date:                          parseTime(t, "2018-02-01T00:00:00Z"),
					PaymentAmount:                 parseDecimal(t, "1001.25"),
					Interest:                      parseDecimal(t, "0.83"),
					Principal:                     parseDecimal(t, "1000.42"),
					InitialOutstandingPrincipal:   parseDecimal(t, "1000.42"),
					RemainingOutstandingPrincipal: parseDecimal(t, "0"),
				},
			},
			want: api.CreateLoanPlanResponse{
				BorrowerPayments: []api.BorrowerPayment{
					{
						Date:                          "2018-01-01T00:00:00Z",
						PaymentAmount:                 "1001.25",
						Interest:                      "1.67",
						Principal:                     "999.58",
						InitialOutstandingPrincipal:   "2000",
						RemainingOutstandingPrincipal: "1000.42",
					},
					{
						Date:                          "2018-02-01T00:00:00Z",
						PaymentAmount:                 "1001.25",
						Interest:                      "0.83",
						Principal:                     "1000.42",
						InitialOutstandingPrincipal:   "1000.42",
						RemainingOutstandingPrincipal: "0",
					},
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SuccessBuildingLoanPlanWithOnePayment",
			requestBody: validCreateLoanRequestBody(t),
			injectResponse: []loan.Payment{
				{
					Date:                          parseTime(t, "2018-01-01T00:00:00Z"),
					PaymentAmount:                 parseDecimal(t, "1001.25"),
					Interest:                      parseDecimal(t, "1.67"),
					Principal:                     parseDecimal(t, "999.58"),
					InitialOutstandingPrincipal:   parseDecimal(t, "2000"),
					RemainingOutstandingPrincipal: parseDecimal(t, "1000.42"),
				},
			},
			want: api.CreateLoanPlanResponse{
				BorrowerPayments: []api.BorrowerPayment{
					{
						Date:                          "2018-01-01T00:00:00Z",
						PaymentAmount:                 "1001.25",
						Interest:                      "1.67",
						Principal:                     "999.58",
						InitialOutstandingPrincipal:   "2000",
						RemainingOutstandingPrincipal: "1000.42",
					},
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			// On the core logic this doesn't seem possible.
			// But it is important to document/prove what happens
			// if the plan is empty.
			name:           "SuccessBuildingEmptyLoanPlan",
			requestBody:    validCreateLoanRequestBody(t),
			injectResponse: []loan.Payment{},
			want: api.CreateLoanPlanResponse{
				BorrowerPayments: []api.BorrowerPayment{},
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// I'm not extremely against mocking frameworks
			// Used them on the past, like testify mocks
			//
			// My overall feeling is that they made the tests
			// more bloated and it was easier to end up with
			// odd error messages that were hard to debug on
			// failures. Also depending on how you use mocks
			// you can end up coupling too much on the structure
			// of the code instead of the behavior.
			//
			// So I tend to prefer lightweight handwritten
			// mocks, preferably fakes (like in-memory storages).
			// And when possible to integrated tests instead of validating
			// individual parameters.
			//
			// There is a good post from Kent Beck that relates to this:
			// https://medium.com/@kentbeck_7670/programmer-test-principles-d01c064d7934
			service := api.New(func(
				totalLoanAmount decimal.Decimal,
				annualInterestRate decimal.Decimal,
				durationInMonths int,
				start time.Time,
			) ([]loan.Payment, error) {
				return test.injectResponse, test.injectErr
			})
			server := httptest.NewServer(service)
			defer server.Close()

			method := http.MethodPost
			if test.method != "" {
				method = test.method
			}

			createLoanPlanURL := server.URL + api.CreateLoanPlanPath
			request := newRequest(t, method, createLoanPlanURL, test.requestBody)
			client := server.Client()

			res, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != test.wantStatusCode {
				t.Fatalf("got response %d want %d", res.StatusCode, test.wantStatusCode)
			}

			if test.wantStatusCode != http.StatusOK {
				wantErr := api.ErrorResponse{}
				fromJSON(t, res.Body, &wantErr)

				// Validate that a message is sent, but not its contents
				// since the message is for human inspection only and
				// should be handled opaquely by code.
				// If necessary we can introduce error codes (strings or ints),
				// but it does not seem necessary for now.
				// If we add some tracing ID for errors this would also
				// be the place to check for them.
				if wantErr.Error.Message == "" {
					t.Fatalf("expected an error message on status code %d", test.wantStatusCode)
				}
				return
			}

			got := api.CreateLoanPlanResponse{}
			fromJSON(t, res.Body, &got)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("api: POST %s mismatch (-want +got):\n%s", api.CreateLoanPlanPath, diff)
			}

		})
	}
}

func fromJSON(t *testing.T, data io.Reader, v interface{}) {
	t.Helper()

	dec := json.NewDecoder(data)
	err := dec.Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
}

func toJSON(t *testing.T, v interface{}) []byte {
	t.Helper()

	j, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func newRequest(t *testing.T, method string, url string, body []byte) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func validCreateLoanRequestBody(t *testing.T) []byte {
	return toJSON(t, api.CreateLoanPlanRequest{
		LoanAmount:  "1000.00",
		NominalRate: "5.0",
		Duration:    1,
		StartDate:   "2020-12-01T00:00:00Z",
	})
}

func parseDecimal(t *testing.T, v string) decimal.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(v)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
