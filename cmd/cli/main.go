// MIT License
//
// Copyright (c) 2024 [El Rio Community School]
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// This program reads from two Excel spreadsheets and generates
// a new Excel spreadsheet with correlated data from the two.  The
// purpose is to generate a list of all students and their respective
// donations from either parents/guardians or other donors (e.g. friends,
// grandparents, etc.)
//
// The first spreadsheet contains a list of parents and their children.
// This is the starting point of the data gathering given that we need
// to generate an inventory of all the students.
//
// Once the students and their parents are gathered from the first spreadsheet,
// we then read from the second spreadsheet which contains donation transaction
// data. The transaction data is then mapped to the students and their respective
// primary donors.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jotacamou/datacor/internal/runctx"
	"github.com/jotacamou/datacor/internal/types"

	excelize "github.com/xuri/excelize/v2"
)

type AllStudents map[string]types.Student

type DonationTransaction struct {
	Date               string
	Name               string
	Amount             string
	FirstStudentName   string
	FirstStudentClass  string
	SecondStudentName  string
	SecondStudentClass string
	ThirdStudentName   string
	ThirdStudentClass  string
	AccountNumber      string
}

var context *runctx.RunContext = new(runctx.RunContext)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <transactions-file>\n", os.Args[0])
		os.Exit(1)
	}

	context.NewTxnReport = os.Args[1]

	fmt.Println(context.NewTxnReport)
	fmt.Println(context.GetNewReportDate())
	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		fmt.Printf("File does not exist: %s\n", context.NewTxnReport)
		os.Exit(1)
	}

	if !strings.HasSuffix(strings.ToLower(os.Args[1]), ".xlsx") {
		fmt.Printf("Transactions file must have an .xlsx extension\n")
		os.Exit(1)
	}

	GenerateDonationsByStudentReport()
}

func GenerateDonationsByStudentReport() {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Donations By Student"
	i, err := f.NewSheet(sheetName)
	if err != nil {
		fmt.Println(err)
		return
	}

	f.SetActiveSheet(i)

	err = f.DeleteSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	donationsByStudentHeader := []interface{}{
		"Student",
		"Class",
		"Care Giver 1",
		"Care Giver 2",
		"Care Giver 3",
		"Primary Donor 1",
		"Primary Donor 2",
		"Primary Donor 3",
		"Primary Donors Per Student",
		"Primary Donor 1 Donation Amount",
		"Primary Donor 2 Donation Amount",
		"Primary Donor 3 Donation Amount",
		"Total Donation Amount",
	}

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Calibri",
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Determine the desired range of cells to format.
	// For instance, you might typically use only up to Z (column 26) and 100 rows.
	maxColumns := 26
	maxRows := 300

	for col := 1; col <= maxColumns; col++ {
		// colName, _ := excelize.ColumnNumberToName(col)
		for row := 1; row <= maxRows; row++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			if err := f.SetCellStyle(sheetName, cell, cell, style); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	dollarAmountStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 165,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = f.SetColStyle(sheetName, "J:M", dollarAmountStyle)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = f.SetCellValue(sheetName, "A1", "Last Updated:")
	if err != nil {
		fmt.Println(err)
	}

	// Create a new style with a yellow background.
	dateStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"}, // Yellow color in HEX format
			Pattern: 1,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	err = f.SetCellStyle(sheetName, "B1", "B1", dateStyle)
	if err != nil {
		fmt.Println(err)
	}
	err = f.SetCellValue(sheetName, "B1", context.GetNewReportDate())
	if err != nil {
		fmt.Println(err)
	}

	err = f.SetSheetRow(sheetName, "A2", &donationsByStudentHeader)
	if err != nil {
		fmt.Println(err)
		return

	}

	if err := f.SetColWidth(sheetName, "A", "A", 20); err != nil {
		fmt.Println(err)
	}

	students, err := makeStudentRows()
	if err != nil {
		fmt.Println(err)
		return
	}

	donations, err := readTransactions()
	if err != nil {
		fmt.Println(err)
		return
	}

	assignDonationsToStudents(students, donations)

	// data contains the rows to be written to the worksheet.
	// This slice of interface slices is what the excelize
	// library expects to write to the worksheet.
	var data [][]interface{}

	for _, student := range students {
		data = append(data, []interface{}{
			student.Name,
			student.Class,
			student.Parent1,
			student.Parent2,
			student.Parent3,
			student.PrimaryDonor1,
			student.PrimaryDonor2,
			student.PrimaryDonor3,
			student.PrimaryDonorsPerStudent,
			student.PrimaryDonor1DonationAmount,
			student.PrimaryDonor2DonationAmount,
			student.PrimaryDonor3DonationAmount,
			student.TotalDonationAmount,
		})
	}

	for i := 3; i < (len(data) + 3); i++ {
		err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", i), &data[i-3])
		if err != nil {
			fmt.Println(err)
		}
	}

	// Resize cells to accomodate value lenghts
	err = xlsAdjustColumnsWidth(f, sheetName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write non care giver donations to the worksheet (separate sheet)
	nonCareGiverDonationsSheetName := "Non Care Giver Donations"

	i, err = f.NewSheet(nonCareGiverDonationsSheetName)
	if err != nil {
		fmt.Println(err)
		return
	}

	f.SetActiveSheet(i)

	nonCareGiverDonationsHeader := []interface{}{
		"Date",
		"Name",
		"Amount",
	}

	err = f.SetSheetRow(nonCareGiverDonationsSheetName, "A1", &nonCareGiverDonationsHeader)
	if err != nil {
		fmt.Println(err)
		return
	}

	nonCareGiverDonations, err := readTransactionsByNonCareGivers()
	if err != nil {
		fmt.Println(err)
		return
	}

	var nonCareGiverDonationsData [][]interface{}

	for _, donation := range nonCareGiverDonations {
		nonCareGiverDonationsData = append(nonCareGiverDonationsData, []interface{}{
			donation.Date,
			donation.Name,
			donation.Amount,
		})
	}

	for i := 2; i < (len(nonCareGiverDonationsData) + 2); i++ {
		err := f.SetSheetRow(
			nonCareGiverDonationsSheetName,
			fmt.Sprintf("A%d", i),
			&nonCareGiverDonationsData[i-2],
		)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Resize cells to accomodate value lenghts
	err = xlsAdjustColumnsWidth(f, nonCareGiverDonationsSheetName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileName := fmt.Sprintf("donations_by_student-%s.xlsx", time.Now().Format("2006-01-02"))
	if err = f.SaveAs(fileName); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Donations by student report saved to %s\n", fileName)
}

// Auto adjust column width based on the content
func xlsAdjustColumnsWidth(f *excelize.File, sheet string) error {
	cols, err := f.GetCols(sheet)
	if err != nil {
		return err
	}

	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell)
			// cellWidth := utf8.RuneCountInString(rowCell) + 2
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			return err
		}

		err = f.SetColWidth(sheet, name, name, float64(largestWidth))
		if err != nil {
			return err
		}
	}

	return nil
}

// assignDonationsToStudents distributes the donation amounts to the respective students based on the donation transactions
func assignDonationsToStudents(students AllStudents, donations []*types.DonationTransaction) {
	for _, txn := range donations {
		siblings := []string{txn.FirstStudentName, txn.SecondStudentName, txn.ThirdStudentName}
		validSiblings := []string{}
		for _, sibling := range siblings {
			if sibling != "" {
				validSiblings = append(validSiblings, sibling)
			}
		}

		// Skip if there are no valid siblings.  We'll have to deal with this separately
		if len(validSiblings) == 0 {
			continue
		}

		donationPerStudent := parseDollarAmount(txn.Amount) / float64(len(validSiblings))

		// fmt.Println("Donation per student:", donationPerStudent, txn.Amount, float64(len(validSiblings)))

		// Update each student's donation information
		for _, sibling := range validSiblings {
			student, exists := students[sibling]
			if !exists {
				fmt.Println("Student does not exist:", sibling)
				continue // Skip if the student does not exist
			}

			// Update the total donation amount
			student.TotalDonationAmount += donationPerStudent

			// Update the primary donors and their donation amounts
			switch {
			case student.PrimaryDonor1 == txn.Name:
				student.PrimaryDonor1DonationAmount += donationPerStudent
			case student.PrimaryDonor2 == txn.Name:
				student.PrimaryDonor2DonationAmount += donationPerStudent
			case student.PrimaryDonor3 == txn.Name:
				student.PrimaryDonor3DonationAmount += donationPerStudent
			case student.PrimaryDonor1 == "":
				student.PrimaryDonor1 = txn.Name
				student.PrimaryDonor1DonationAmount = donationPerStudent
			case student.PrimaryDonor2 == "":
				student.PrimaryDonor2 = txn.Name
				student.PrimaryDonor2DonationAmount = donationPerStudent
			case student.PrimaryDonor3 == "":
				student.PrimaryDonor3 = txn.Name
				student.PrimaryDonor3DonationAmount = donationPerStudent
			}

			// Update the number of primary donors
			student.PrimaryDonorsPerStudent = 0
			if student.PrimaryDonor1 != "" {
				student.PrimaryDonorsPerStudent++
			}
			if student.PrimaryDonor2 != "" {
				student.PrimaryDonorsPerStudent++
			}
			if student.PrimaryDonor3 != "" {
				student.PrimaryDonorsPerStudent++
			}

			// Save the updated student back to the map
			students[sibling] = student
		}

	}
}

// parseDollarAmount takes a string formatted as a dollar amount (e.g., "$10.00")
// and converts it to a float64.
func parseDollarAmount(amount string) float64 {
	// Remove the dollar sign from the start of the string
	cleanedAmount := strings.TrimPrefix(amount, "$")
	// Convert the string to a float64
	value, err := strconv.ParseFloat(cleanedAmount, 64)
	if err != nil {
		fmt.Println("Error parsing dollar amount:", err)
		return 0
	}
	return value
}

func makeStudentRows() (map[string]types.Student, error) {
	parents, err := getParents()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	students := make(AllStudents)

	for _, parent := range parents {
		for _, child := range parent.Children {
			if existingChild, ok := students[child.Name]; ok {
				addParent(&existingChild, parent.Name)
				students[child.Name] = existingChild
			} else {
				addParent(&child, parent.Name)
				students[child.Name] = child
			}
		}
	}

	return students, nil
}

func addParent(student *types.Student, parentName string) {
	parents := []*string{
		&student.Parent1,
		&student.Parent2,
		&student.Parent3,
	}
	for _, parent := range parents {
		if *parent == "" {
			*parent = parentName
			break
		}
	}
}

// getParents reads the parent-child data from an Excel spreadsheet
func getParents() ([]*types.Parent, error) {
	// TODO: parameterize the file name
	fileName := "parents-kids-classes.xlsx"
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	rows, err := f.GetRows("Data")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var parents []*types.Parent

	for rowIndex, row := range rows {
		// Skip the header row
		if rowIndex == 0 {
			continue
		}

		parentName := row[0]
		accountNumber := row[7]

		// All parents on this list have at least one child
		firstChildName := row[1]
		firstChildClass := row[2]

		parent := &types.Parent{
			Name: parentName,
			Children: []types.Student{
				{ // First child
					Name:  firstChildName,
					Class: firstChildClass,
				},
			},
			AccountNumber: accountNumber,
		}

		// Check if parent has a second children
		if row[3] != "" {
			parent.Children = append(parent.Children, types.Student{
				Name:  row[3],
				Class: row[4],
			})
		}

		// Check if parent has a third children, only likely if they have a second child
		if row[5] != "" {
			parent.Children = append(parent.Children, types.Student{
				Name:  row[5],
				Class: row[6],
			})
		}

		parents = append(parents, parent)

	}

	return parents, nil
}

func readTransactions() ([]*types.DonationTransaction, error) {
	f, err := excelize.OpenFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	rows, err := f.GetRows("Data")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var donations []*types.DonationTransaction

	for rowIndex, row := range rows {
		// Skip header and total rows
		if rowIndex == 0 || rowIndex == 1 {
			continue
		}

		txn := &types.DonationTransaction{
			Date:               row[0],
			Name:               row[1],
			Amount:             strings.ReplaceAll(row[2], " ", ""),
			FirstStudentName:   row[6],
			FirstStudentClass:  row[7],
			SecondStudentName:  row[8],
			SecondStudentClass: row[9],
			ThirdStudentName:   row[10],
			ThirdStudentClass:  row[11],
			AccountNumber:      row[12],
		}

		donations = append(donations, txn)

	}

	return donations, nil
}

func readTransactionsByNonCareGivers() ([]*types.DonationTransaction, error) {
	f, err := excelize.OpenFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	rows, err := f.GetRows("Data")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var donations []*types.DonationTransaction

	for rowIndex, row := range rows {
		// Skip header and total rows
		if rowIndex == 0 || rowIndex == 1 {
			continue
		}

		// Ignore transactions with students associated with them
		if row[6] != "" {
			continue
		}
		if row[8] != "" {
			continue
		}
		if row[10] != "" {
			continue
		}

		txn := &types.DonationTransaction{
			Date:          row[0],
			Name:          row[1],
			Amount:        strings.ReplaceAll(row[2], " ", ""),
			AccountNumber: row[12],
		}

		donations = append(donations, txn)

	}

	return donations, nil
}
