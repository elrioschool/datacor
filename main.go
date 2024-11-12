// MIT License
//
// Copyright (c) 2024 [JR Camou <jr@camou.org>]
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
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	excelize "github.com/xuri/excelize/v2"
)

type Parent struct {
	Name          string
	Children      []Student
	AccountNumber string
}

type Student struct {
	Name                        string
	Grade                       string
	Class                       string
	Parent1                     string
	Parent2                     string
	Parent3                     string
	PrimaryDonor1               string
	PrimaryDonor2               string
	PrimaryDonor3               string
	PrimaryDonorsPerStudent     int
	PrimaryDonor1DonationAmount float64
	PrimaryDonor2DonationAmount float64
	PrimaryDonor3DonationAmount float64
	TotalDonationAmount         float64
}

type AllStudents map[string]Student

type DonationTransation struct {
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

func main() {
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
}

// assignDonationsToStudents distributes the donation amounts to the respective students based on the donation transactions
func assignDonationsToStudents(students AllStudents, donations []*DonationTransation) {
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

		//fmt.Println("Donation per student:", donationPerStudent, txn.Amount, float64(len(validSiblings)))

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

	spew.Dump(students)
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

func makeStudentRows() (map[string]Student, error) {
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

	// spew.Dump(students)

	return students, nil
}

func addParent(student *Student, parentName string) {
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
func getParents() ([]*Parent, error) {
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

	var parents []*Parent

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

		parent := &Parent{
			Name: parentName,
			Children: []Student{
				{ // First child
					Name:  firstChildName,
					Class: firstChildClass,
				},
			},
			AccountNumber: accountNumber,
		}

		// Check if parent has a second children
		if row[3] != "" {
			parent.Children = append(parent.Children, Student{
				Name:  row[3],
				Class: row[4],
			})
		}

		// Check if parent has a third children, only likely if they have a second child
		if row[5] != "" {
			parent.Children = append(parent.Children, Student{
				Name:  row[5],
				Class: row[6],
			})
		}

		parents = append(parents, parent)

	}

	return parents, nil
}

func readTransactions() ([]*DonationTransation, error) {
	fileName := "Watershed_donors_kids_&_class_names_previous_31_days.xlsx"
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

	var donations []*DonationTransation

	for rowIndex, row := range rows {
		// Skip header and total rows
		if rowIndex == 0 || rowIndex == 1 {
			continue
		}

		txn := &DonationTransation{
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
