package models

import (
	"time"
)

// Reservation holds reservation data
type Reservation struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type User struct {
	FirstName    string
	LastName     string
	Email        string
	ID           uint
	PasswordHash []byte
	AccessLevel  int
	Created_at   time.Time
	Updated_at   time.Time
}

type Post struct {
	ID         uint
	UID        uint
	Likes      uint
	Content    string
	Created_at time.Time
	Updated_at time.Time
}
