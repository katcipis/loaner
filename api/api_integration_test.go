package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/katcipis/loaner/api"
	"github.com/katcipis/loaner/loan"
)

func TestLoanPlanCreationIntegration(t *testing.T) {
	type Test struct {
		name           string
		request        api.CreateLoanPlanRequest
		wantStatusCode int
		want           api.CreateLoanPlanResponse
	}

	tests := []Test{
		{
			name: "SuccessOn2000LoanWith1.0RateIn2Months",
			request: api.CreateLoanPlanRequest{
				LoanAmount:  "2000.0",
				NominalRate: "1.0",
				Duration:    2,
				StartDate:   "2018-01-01T00:00:00Z",
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := api.New(loan.CreatePlan)
			server := httptest.NewServer(service)
			defer server.Close()

			createLoanPlanURL := server.URL + api.CreateLoanPlanPath
			request := newRequest(t, http.MethodPost, createLoanPlanURL, toJSON(t, test.request))
			client := server.Client()

			res, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != test.wantStatusCode {
				t.Fatalf("got response %d want %d", res.StatusCode, test.wantStatusCode)
			}

			got := api.CreateLoanPlanResponse{}
			fromJSON(t, res.Body, &got)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("api: POST %s mismatch (-want +got):\n%s", api.CreateLoanPlanPath, diff)
			}

		})
	}
}
