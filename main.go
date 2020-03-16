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

const csvDir = "large/"

// Bank is banks.csv
type Bank struct {
	ID       int // [PK]
	BankName string
}

// Facility is facilities.csv
type Facility struct {
	ID           int     // [PK]
	BankID       int     // id of the bank providing this facility. [FK]
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

var banks map[int]Bank
var facilities []Facility
var covenants []Covenant
var assignments map[int]Assignment
var yields map[int]Yield

func init() {
	banks = make(map[int]Bank)
	assignments = make(map[int]Assignment)
	yields = make(map[int]Yield)
}

func main() {
	readSetupFiles()
	processLoans()
	saveFiles()
}

func readSetupFiles() {
	lines, err := readCsv(csvDir + "banks.csv")
	if err != nil {
		log.Fatal(err)
	}

	// process one line at a time and save each record in a map
	for _, l := range lines {
		id, _ := strconv.Atoi(l[0]) // ignore parsing errors for now
		b := Bank{
			ID:       id,
			BankName: l[1],
		}
		banks[b.ID] = b // bank ID is a primary key for the record
	}

	lines, err = readCsv(csvDir + "facilities.csv")
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range lines {
		amount, _ := strconv.ParseFloat(l[0], 32) // CSV has int formatted as float
		interest, _ := strconv.ParseFloat(l[1], 32)
		id, _ := strconv.Atoi(l[2])
		bankID, _ := strconv.Atoi(l[3])
		f := Facility{
			ID:           id,
			BankID:       bankID,
			InterestRate: float32(interest),
			Amount:       int(amount),
		}
		facilities = append(facilities, f) // TODO: sort by interest rate
	}

	lines, err = readCsv(csvDir + "covenants.csv")
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range lines {
		facilityID, _ := strconv.Atoi(l[0])
		maxDefault, _ := strconv.ParseFloat(l[1], 32)
		bankID, _ := strconv.Atoi(l[2])
		state := l[3]
		c := Covenant{
			BankID:      bankID,
			FacilityID:  facilityID,
			MaxDefault:  float32(maxDefault),
			BannedState: state,
		}
		covenants = append(covenants, c)
	}

	return
}

func processLoans() {
	// read loans
	lines, err := readCsv(csvDir + "loans.csv")
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range lines {
		interest, _ := strconv.ParseFloat(l[0], 32)
		amount, _ := strconv.Atoi(l[1])
		id, _ := strconv.Atoi(l[2])
		defaultRate, _ := strconv.ParseFloat(l[3], 32)
		state := l[4]
		l := Loan{
			ID:           id,
			Amount:       amount,
			InterestRate: float32(interest),
			DefaultRate:  float32(defaultRate),
			State:        state,
		}

		makeAssignment(l) // send loan for processing
	}
	return
}

func makeAssignment(loan Loan) {
	log.Printf("Loan: %v\n", loan)

	// brute force search for facility to satisfy loan amount
	isAssigned := false
	for fIdx, f := range facilities {
		log.Printf("Evaluating facility: %v\n", f)

		// skip over facilities with too little $
		if f.Amount < loan.Amount {
			log.Printf("Skipping due to amount: %d\n", f.Amount)
			continue
		}

		// brute force scan through covenants
		for _, c := range covenants {
			// ignore mismatching facilities
			if c.FacilityID != f.ID {
				continue
			}

			log.Printf("Covenant: %v\n", c)

			// skip over disallowed states
			if c.BannedState == loan.State {
				log.Printf("Skipping due to state: %s\n", c.BannedState)
				continue
			}

			// skip over if default rates covenant exists and the limit is breached
			if c.MaxDefault > 0 && c.MaxDefault < loan.DefaultRate {
				log.Printf("Skipping due to default rate: %f\n", c.MaxDefault)
				continue
			}

			log.Printf("Assigning facility: %d\n", f.ID)

			// *** critical section -- should be locked if using concurrency
			// create an assignment for the loan
			a := Assignment{
				LoanID:     loan.ID,
				FacilityID: f.ID,
			}
			// reduce facility's balance
			f.Amount -= loan.Amount
			facilities[fIdx] = f

			// save assignemts for output
			assignments[loan.ID] = a

			// save yeilds for output
			y := calcYields(loan, f)
			yields[f.ID] = y

			isAssigned = true
			// ***
			break
		}

		if isAssigned {
			break // no need to scan through the rest of facilities
		}
	}
	return
}

func calcYields(loan Loan, f Facility) Yield {
	loanYield, err := getLoanYield(loan.DefaultRate, loan.InterestRate, loan.Amount, f.InterestRate)
	if err != nil {
		log.Fatal(err)
	}

	y := Yield{
		FacilityID:    f.ID,
		ExpectedYield: int(loanYield),
	}
	// add facility yield if exists
	if yOld, ok := yields[f.ID]; ok {
		y.ExpectedYield += yOld.ExpectedYield
	}
	return y
}

func saveFiles() {
	yData := [][]string{
		{"facility_id", "expected_yeild"},
	}
	for _, y := range yields {
		yData = append(yData, []string{strconv.Itoa(y.FacilityID), strconv.Itoa(y.ExpectedYield)})
	}
	err := writeCsv(csvDir+"yields.csv", yData)
	if err != nil {
		log.Fatal(err)
	}

	aData := [][]string{
		{"loan_id", "facility_id"},
	}
	for _, a := range assignments {
		aData = append(aData, []string{strconv.Itoa(a.LoanID), strconv.Itoa(a.FacilityID)})
	}
	err = writeCsv(csvDir+"assignments.csv", aData)
	if err != nil {
		log.Fatal(err)
	}

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

func writeCsv(csvFile string, rows [][]string) error {
	fh, err := os.Create(csvFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	writer := csv.NewWriter(fh)
	err = writer.WriteAll(rows)
	if err != nil {
		return err
	}

	return nil
}
