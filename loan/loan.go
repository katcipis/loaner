// Package loan defines simple functions to do loan related calculations
package loan

import (
	"fmt"

	"github.com/shopspring/decimal"
)

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
	durationInMonths uint,
) (decimal.Decimal, error) {

	if durationInMonths == 0 {
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
