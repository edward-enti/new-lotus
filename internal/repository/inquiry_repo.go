package repository

import (
	"database/sql"

	"sonlotus/internal/domain"
)

// InquiryRepo writes and reads customer inquiries.
type InquiryRepo struct{ DB *sql.DB }

// Create inserts an inquiry and returns its id.
func (r *InquiryRepo) Create(inq domain.Inquiry) (int64, error) {
	res, err := r.DB.Exec(`INSERT INTO inquiries(customer_name, phone, email, company_name, inquiry_type, product_id, surface_area_m2, message, status, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'new', ?)`,
		inq.CustomerName, inq.Phone, inq.Email, inq.CompanyName, inq.InquiryType, inq.ProductID, inq.SurfaceAreaM2, inq.Message, inq.IPAddress)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns inquiries newest first (admin use, Phase 3).
func (r *InquiryRepo) List() ([]domain.Inquiry, error) {
	rows, err := r.DB.Query(`SELECT id, customer_name, phone, email, company_name, inquiry_type, product_id, surface_area_m2, message, status, admin_notes, ip_address, created_at, updated_at FROM inquiries ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Inquiry
	for rows.Next() {
		var inq domain.Inquiry
		if err := rows.Scan(&inq.ID, &inq.CustomerName, &inq.Phone, &inq.Email, &inq.CompanyName, &inq.InquiryType, &inq.ProductID, &inq.SurfaceAreaM2, &inq.Message, &inq.Status, &inq.AdminNotes, &inq.IPAddress, &inq.CreatedAt, &inq.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, inq)
	}
	return out, rows.Err()
}

// ValidStatuses are the SPEC §4.1 CHECK-constraint values. Status
// management moves an inquiry between these values; no sequential-only
// transition rule is enforced (an admin may close a new inquiry
// directly), only membership in this set.
var ValidStatuses = []string{"new", "contacted", "quoted", "closed"}

// ValidStatus reports whether s is an allowed inquiry status.
func ValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// UpdateStatus sets an inquiry's status (admin use). Callers must
// validate with ValidStatus first; the CHECK constraint is the backstop.
func (r *InquiryRepo) UpdateStatus(id int64, status string) error {
	_, err := r.DB.Exec(`UPDATE inquiries SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, id)
	return err
}

// CountByStatus returns the number of inquiries in a status.
func (r *InquiryRepo) CountByStatus(status string) (int, error) {
	var n int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM inquiries WHERE status = ?`, status).Scan(&n)
	return n, err
}
