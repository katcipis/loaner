package loan_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/katcipis/loaner/loan"
	"github.com/shopspring/decimal"
)

// Here you can find some table driven tests, I'm very fond to them :-).
// This presentation talks a little about them
// (and some other interesting ideas): https://www.youtube.com/watch?v=8hQG7QlcLBk

func TestCreatePlan(t *testing.T) {

	// Test represents a single test case for the loan plan creation
	// We use strings here to make the test descriptor more concise and
	// easier to read as a high level behavior descriptor.
	type Test struct {
		name               string
		totalLoanAmount    string
		annualInterestRate string
		durationInMonths   int
		startDate          time.Time
		want               []loan.Payment
		wantErr            error
	}

	tests := []Test{
		{
			name:               "SuccessOn5000LoanWith5.0RateIn3Months",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2018-01-01T00:00:00Z"),
			want: []loan.Payment{
				{
					Date: parseTime(t, "2018-01-01T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2018-02-01T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2018-03-01T00:00:00Z"),
				},
			},
		},
		{
			name:               "TimeAndTimezoneInfoOnDateIsIgnored",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2018-01-01T12:00:00+01:00"),
			want: []loan.Payment{
				{
					Date: parseTime(t, "2018-01-01T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2018-02-01T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2018-03-01T00:00:00Z"),
				},
			},
		},
		{
			name:               "SuccessAcrossYearBoundary",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2020-12-28T00:00:00Z"),
			want: []loan.Payment{
				{
					Date: parseTime(t, "2020-12-28T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2021-01-28T00:00:00Z"),
				},
				{
					Date: parseTime(t, "2021-02-28T00:00:00Z"),
				},
			},
		},
		{
			name:               "ErrorOnStartDateDay29",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2020-12-29T00:00:00Z"),
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorOnStartDateDay30",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2020-12-29T00:00:00Z"),
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorOnStartDateDay31",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   3,
			startDate:          parseTime(t, "2020-12-29T00:00:00Z"),
			wantErr:            loan.ErrInvalidParameter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loanAmount := toDecimal(t, test.totalLoanAmount)
			interestRate := toDecimal(t, test.annualInterestRate)

			got, err := loan.CreatePlan(
				loanAmount,
				interestRate,
				test.durationInMonths,
				test.startDate,
			)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("got error %v; want %v", err, test.wantErr)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("CreatePlan() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAnnuityCalculation(t *testing.T) {

	// Test represents a single test case for the annuity calculation
	// We use strings here to make the test descriptor more concise and
	// easier to read as a high level behavior descriptor.
	type Test struct {
		name               string
		totalLoanAmount    string
		annualInterestRate string
		durationInMonths   int
		want               string
		wantErr            error
	}

	tests := []Test{
		{
			name:               "SuccessOn5000LoanWith5.0RateIn24Months",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   24,
			want:               "219.36",
		},
		{
			name:               "SuccessOn5000LoanWith5.0RateIn2Months",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   2,
			want:               "2515.64",
		},
		{
			name:               "SuccessOn1LoanWith5.0RateIn2Months",
			totalLoanAmount:    "1.0",
			annualInterestRate: "5.0",
			durationInMonths:   2,
			want:               "0.5",
		},
		{
			name:               "SuccessOn1LoanWith5.0RateIn500Months",
			totalLoanAmount:    "1.0",
			annualInterestRate: "5.0",
			durationInMonths:   50,
			want:               "0.02",
		},
		{
			name:               "SuccessOn5000LoanWith5.0RateIn1Month",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "5.0",
			durationInMonths:   1,
			want:               "5020.83",
		},
		{
			name:               "SuccessOn5000LoanWith0.1RateIn24Months",
			totalLoanAmount:    "5000.0",
			annualInterestRate: "0.1",
			durationInMonths:   24,
			want:               "208.55",
		},
		{
			name:               "ErrorIfDurationIsZero",
			totalLoanAmount:    "500.00",
			annualInterestRate: "3.0",
			durationInMonths:   0,
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorIfDurationIsNegative",
			totalLoanAmount:    "500.00",
			annualInterestRate: "3.0",
			durationInMonths:   -1,
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorIfLoanAmountIsZero",
			totalLoanAmount:    "0.00",
			annualInterestRate: "3.0",
			durationInMonths:   2,
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorIfLoanAmountIsNegative",
			totalLoanAmount:    "-10.00",
			annualInterestRate: "3.0",
			durationInMonths:   2,
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorIfInterestRateIsNegative",
			totalLoanAmount:    "10.00",
			annualInterestRate: "-1.0",
			durationInMonths:   2,
			wantErr:            loan.ErrInvalidParameter,
		},
		{
			name:               "ErrorIfInterestRateIsZero",
			totalLoanAmount:    "10.00",
			annualInterestRate: "0",
			durationInMonths:   2,
			wantErr:            loan.ErrInvalidParameter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loanAmount := toDecimal(t, test.totalLoanAmount)
			interestRate := toDecimal(t, test.annualInterestRate)
			want := decimal.Decimal{}

			if test.want != "" {
				want = toDecimal(t, test.want)
			}

			got, err := loan.CalculateAnnuity(loanAmount, interestRate, test.durationInMonths)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("got error %v; want %v", err, test.wantErr)
			}

			if !got.Equal(want) {
				t.Errorf("got result %v; want %v", got, test.want)
			}
		})
	}
}

func toDecimal(t *testing.T, v string) decimal.Decimal {
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
