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
			requestBody:    toJSON(t, api.CreateLoanPlanRequest{}),
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "BadRequestIfParametersAreInvalid",
			requestBody:    toJSON(t, api.CreateLoanPlanRequest{}),
			injectErr:      loan.ErrInvalidParameter,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "InternalServerErrorOnLoanCalculationError",
			requestBody:    toJSON(t, api.CreateLoanPlanRequest{}),
			injectErr:      errors.New("injected generic error"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// I'm not extremely against mocking frameworks
			// Used them on the past, like testify mocks
			// My overall feeling is that they made the tests
			// more bloated and it was easier to end up with
			// odd error messages that were hard to debug on
			// failures. Also depending on how you use mocks
			// you can end up coupling too much on the structure
			// of the code instead of the behavior.
			// So I tend to prefer lightweight handwritten
			// mocks, preferably fakes (like in-memory storages).
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

			if test.wantStatusCode != http.StatusCreated {
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
