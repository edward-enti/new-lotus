package domain

import "time"

// Inquiry is a customer lead / quote request (SPEC §4.1).
type Inquiry struct {
	ID            int64
	CustomerName  string
	Phone         string
	Email         string
	CompanyName   string
	InquiryType   string
	ProductID     *int64
	SurfaceAreaM2 float64
	Message       string
	Status        string
	AdminNotes    string
	IPAddress     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
