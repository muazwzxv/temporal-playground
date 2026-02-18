// Package entity contains the domain entities representing core business objects.
package entity

import (
	"time"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusArchived UserStatus = "archived"

	// used for redis cache idempotency
	UserStatusProcessing UserStatus = "processing"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusArchived:
		return true
	default:
		return false
	}
}

func (s UserStatus) String() string {
	return string(s)
}

type User struct {
	ID        int64      `db:"id" json:"id"`
	UserUUID  string     `db:"user_uuid" json:"user_uuid"`
	Name      string     `db:"name" json:"name"`
	Status    UserStatus `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *User) IsActive() bool {
	return e.Status == UserStatusActive
}

func (e *User) MarkAsActive() {
	e.Status = UserStatusActive
	e.UpdatedAt = time.Now()
}

func (e *User) MarkAsInactive() {
	e.Status = UserStatusInactive
	e.UpdatedAt = time.Now()
}

func (e *User) MarkAsArchived() {
	e.Status = UserStatusArchived
	e.UpdatedAt = time.Now()
}
