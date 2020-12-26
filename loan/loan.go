// Package loan defines simple functions to do loan related calculations
package loan

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Payment represents a loan payment with all its information.
type Payment struct {
	Date                          time.Time
	PaymentAmount                 decimal.Decimal
	Interest                      decimal.Decimal
	Principal                     decimal.Decimal
	InitialOutstandingPrincipal   decimal.Decimal
	RemainingOutstandingPrincipal decimal.Decimal
}

// Error represents an enumeration of errors returned by the loan
// package. These errors can be used as error sentinels to check
// for specific classes of errors.
//
// Always use errors.Is to check since the error sentinel will be
// wrapper with more context.
type Error string

const (
	// ErrInvalidParameter is returned when on of the parameters passed is invalid.
	ErrInvalidParameter Error = "invalid parameter"
)

// CreatePlan will create a payment plan, as a list of payments,
// throughout the lifetime of an annuity loan.
//
// The annual interest rate is informed as a percent, like 5.0, meaning 5 per cent an year.
//
// It returns an error if any of the parameters is invalid, like the duration
// in months being zero.
func CreatePlan(
	totalLoanAmount decimal.Decimal,
	annualInterestRate decimal.Decimal,
	durationInMonths int,
	start time.Time,
) ([]Payment, error) {
	payments := make([]Payment, durationInMonths)

	for i := range payments {
		// TODO: handle corner cases
		year := start.Year()
		month := start.Month() + time.Month(i)
		day := start.Day()

		payments[i] = Payment{
			Date: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		}
	}
	return payments, nil
}

// CalculateAnnuity will calculate the annuity payment according to the
// formula described here: https://financeformulas.net/Annuity_Payment_Formula.html
//
// The annual interest rate is informed as a percent, like 5.0, meaning 5 per cent an year.
//
// It returns an error if any of the parameters is invalid, like the duration
// in months being zero.
func CalculateAnnuity(
	totalLoanAmount decimal.Decimal,
	annualInterestRate decimal.Decimal,
	durationInMonths int,
) (decimal.Decimal, error) {

	if durationInMonths <= 0 {
		return decimal.Zero, fmt.Errorf(
			"can't calculate annuity:%w: duration should be bigger than 0, it is %v",
			ErrInvalidParameter,
			durationInMonths,
		)
	}

	if totalLoanAmount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf(
			"can't calculate annuity:%w: loan amount should be bigger than 0, it is %v",
			ErrInvalidParameter,
			totalLoanAmount,
		)
	}

	if annualInterestRate.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf(
			"can't calculate annuity:%w: interest rate should be bigger than 0, it is %v",
			ErrInvalidParameter,
			annualInterestRate,
		)
	}

	const precision = 2

	// Assuming for all calculation that the default precision of 16 is enough
	// Only the final result is rounded.
	monthlyInterestRate := fromPercentToDecimal(calculateMonthlyInterestRate(annualInterestRate))
	one := decimal.NewFromInt(1)
	numerator := totalLoanAmount.Mul(monthlyInterestRate)
	denominator := one.Add(monthlyInterestRate)
	denominator = denominator.Pow(decimal.NewFromInt(int64(durationInMonths)).Neg())
	denominator = one.Sub(denominator)

	return numerator.Div(denominator).RoundBank(precision), nil
}

func (e Error) Error() string {
	return string(e)
}

func calculateMonthlyInterestRate(annualInterestRate decimal.Decimal) decimal.Decimal {
	monthsInYear := decimal.NewFromInt(12)
	return annualInterestRate.Div(monthsInYear)
}

func fromPercentToDecimal(percentVal decimal.Decimal) decimal.Decimal {
	return percentVal.Div(decimal.NewFromInt(100))
}
