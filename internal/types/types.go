package types

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

// type AllStudents map[string]Student
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

// StorageObjectData contains metadata of the Cloud Storage object.
type StorageObjectData struct {
	Bucket string `json:"bucket,omitempty"`
	Name   string `json:"name,omitempty"`
}
