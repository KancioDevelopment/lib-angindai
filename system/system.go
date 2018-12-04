package system

import (
	uuid "github.com/satori/go.uuid"
)

// Use this user id if you need to insert data
// created by system
func SystemUserID() uuid.UUID {
	systemUserID, _ := uuid.FromString("819a6572-d825-4dc4-8d0a-71177e62e795")
	return systemUserID
}
