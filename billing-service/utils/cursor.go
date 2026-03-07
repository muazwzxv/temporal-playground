package utils

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// PaginationCursor represents the internal cursor structure for pagination
type PaginationCursor struct {
	CreatedAt time.Time `json:"t"`
	ID        int64     `json:"id"`
}

// EncodeCursor encodes cursor values to an opaque Base64 string
func EncodeCursor(createdAt time.Time, id int64) string {
	c := PaginationCursor{CreatedAt: createdAt, ID: id}
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes an opaque cursor string
// Returns zero values if cursor is empty, error if invalid
func DecodeCursor(cursor string) (time.Time, int64, error) {
	if cursor == "" {
		return time.Time{}, 0, nil
	}
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, err
	}
	var c PaginationCursor
	if err := json.Unmarshal(data, &c); err != nil {
		return time.Time{}, 0, err
	}
	return c.CreatedAt, c.ID, nil
}
