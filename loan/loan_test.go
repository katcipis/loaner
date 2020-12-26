package loan_test

import (
	"errors"
	"testing"

	"github.com/katcipis/loaner/loan"
	"github.com/shopspring/decimal"
)

// Here you can find some table driven tests, I'm very fond to them :-).
// This presentation talks a little about them
// (and some other interesting ideas): https://www.youtube.com/watch?v=8hQG7QlcLBk

func TestAnnuityCalculation(t *testing.T) {

	// Test represents a single test case for the annuity calculation
	// We use strings here to make the test descriptor more concise and
	// easier to read as a high level behavior descriptor.
	type Test struct {
		name               string
		totalLoanAmount    string
		annualInterestRate string
		durationInMonths   uint
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
			name:               "ErrorIfDurationIsZero",
			totalLoanAmount:    "500.00",
			annualInterestRate: "3.0",
			durationInMonths:   0,
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
