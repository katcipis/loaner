// Package loan defines simple functions to do loan related calculations
package loan

import "github.com/shopspring/decimal"

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
// It returns an error if any of the parameters is invalid, like the duration
// in months being zero.
func CalculateAnnuity(
	totalLoanAmount decimal.Decimal,
	annualInterestRate decimal.Decimal,
	durationInMonths uint,
) (decimal.Decimal, error) {
	return decimal.Decimal{}, ErrInvalidParameter
}

func (e Error) Error() string {
	return string(e)
}
