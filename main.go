// Balance the Loans Books
//
// - money is borrowed from debt facilities and used to extend loans
// - banks require covenants (restrictions)
// - multiple facilities per bank are allowed
//
// Constrains:
// - loans processed in order received, assigned to cheapest facility that satisfies covenants
// - all yields are non-negative

package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

// Bank is banks.csv
type Bank struct {
	ID       int // [PK]
	BankName string
}

// Facility is facilities.csv
type Facility struct {
	BankID       int     // id of the bank providing this facility. [FK]
	ID           int     // [PK]
	InterestRate float32 // between 0 and 1. We are charged $ * rate by this facility
	Amount       int     // Total Capacity in cents
}

// Covenant is covenants.csv
type Covenant struct {
	BankID      int     // [FK]
	FacilityID  int     // [FK], null means this covenants applies to all of the banks facilities
	MaxDefault  float32 // max allowed default rate for loans in the facility (or the bank's facilities)
	BannedState string  // loan origination from this state prohibited, nullable
}

// Loan is loans.csv
type Loan struct {
	ID           int     // autoincrement
	Amount       int     // size of the loan in cents
	InterestRate float32 // between 0 and 1. Earnings for non-defaulting loan is $ * rate
	DefaultRate  float32 // between 0 and 1. Probability of default === no earnings
	State        string  // state of the loan's origin
}

// Assignment is assignments.csv
type Assignment struct {
	LoanID     int // [FK]
	FacilityID int // [FK]
}

// Yield is yields.csv
type Yield struct {
	FacilityID    int // [FK]
	ExpectedYield int // Facility's yield, rounded to cent. Sum(Loans.InterestRate) for all loans within
}

func getLoanYield(defaultRate float32, loanRate float32, amount int, facilityRate float32) (float32, error) {
	// TODO: define acceptable boundaries for inputs, return Error if breached
	// TODO: typecheck amount casting, return Error
	amountInt := float32(amount)

	expectYield := (1-defaultRate)*loanRate*amountInt - defaultRate*amountInt - facilityRate*amountInt

	return expectYield, nil
}

func main() {
	readSetupFiles()
	readLoans()
	makeAssignments()
	calcYield()
	saveFiles()
}

func readSetupFiles() {
	lines, err := readCsv("banks.csv")
	if err != nil {
		log.Fatal(err)
	}

	// process one line at a time and save each record in a map
	banks := make(map[int]Bank)
	for _, l := range lines {
		id, _ := strconv.Atoi(l[0]) // ignore parsing errors
		b := Bank{
			ID:       id,
			BankName: l[1],
		}
		banks[b.ID] = b // bank ID is a primary key for the record
	}

	lines, err = readCsv("facilities.csv")
	if err != nil {
		log.Fatal(err)
	}

	lines, err = readCsv("covenants.csv")
	if err != nil {
		log.Fatal(err)
	}
}

func readLoans() {
	lines, err := readCsv("loans.csv")
	if err != nil {
		log.Fatal(err)
	}
}

func makeAssignments() {

}

func calcYield() {

}

func saveFiles() {

}

func readCsv(csvFile string) ([][]string, error) {
	fh, err := os.Open(csvFile)
	if err != nil {
		return [][]string{}, err
	}
	defer fh.Close()

	// skip first line header
	reader := csv.NewReader(fh)
	if _, err := reader.Read(); err != nil {
		return [][]string{}, err
	}

	// read the rest of the file
	rows, err := reader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return rows, nil
}
