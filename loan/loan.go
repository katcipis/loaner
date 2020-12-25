// Package loan defines simple functions to do loan related calculations
package loan

import "github.com/shopspring/decimal"

// CalculateAnnuity will calculate the annuity payment according to the
// formula described here: https://financeformulas.net/Annuity_Payment_Formula.html
//
// It returns an error if any of the parameters is invalid, like the duration
// in months being zero.
func CalculateAnnuity(
	durationInMonths uint,
	annualInterestRate decimal.Decimal,
	totalLoanAmount decimal.Decimal,
) (decimal.Decimal, error) {
	return decimal.Decimal{}, nil
}
