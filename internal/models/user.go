package models

import "time"

// DateLayout is the canonical date format used across the API for dob.
const DateLayout = "2006-01-02"

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,datetime=2006-01-02"`
}

// UpdateUserRequest is the payload for updating a user.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,datetime=2006-01-02"`
}

// UserResponse is returned by create/update endpoints (no age field).
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
}

// UserWithAgeResponse is returned by get/list endpoints (includes age).
type UserWithAgeResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
	Age  int    `json:"age"`
}

// CalculateAge returns the full years elapsed between dob and the reference
// time `now`, accounting for whether the birthday has occurred yet this year.
func CalculateAge(dob time.Time, now time.Time) int {
	dob = dob.UTC()
	now = now.UTC()

	age := now.Year() - dob.Year()
	// If the birthday hasn't occurred yet this year, subtract one.
	if now.Month() < dob.Month() ||
		(now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	if age < 0 {
		age = 0
	}
	return age
}
