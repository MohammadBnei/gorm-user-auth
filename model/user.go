package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `json:"email"`
	Password string `json:"password"`
}

/*
BeforeCreate sets the CreatedAt and UpdatedAt fields to the current time,
hashes the user's password, and stores the hashed password in the Password field.

Args:

	u (*User): a pointer to a User struct that includes the password to be hashed.
	tx (*gorm.DB): a GORM database transaction.

Returns:

	err (error): an error that occurred while setting the CreatedAt and UpdatedAt fields, hashing the password, or storing the hashed password in the Password field.
*/
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	u.Password = string(hashedPassword)

	return
}

/*
BeforeSave is a function that updates the User's update time and hashes
the password if it has been changed before saving to the database.

Args:

	u (*User): a pointer to a User struct that includes the password to be hashed.
	tx (*gorm.DB): a GORM database transaction.

Returns:

	err (error): an error that occurred while setting the CreatedAt and UpdatedAt fields, hashing the password, or storing the hashed password in the Password field.
*/
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()

	if tx.Statement.Changed("Password") {
		hashedPassword, error := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if error != nil {
			err = error
			return
		}

		u.Password = string(hashedPassword)
	}

	return
}

/*
CheckPassword takes a password string as input and compares it to the hashed password stored in the User struct.
It returns an error if the comparison fails.

Args:

	u (*User): a pointer to a User struct that includes the password to be hashed.
	tx (*gorm.DB): a GORM database transaction.
	password (string): The password to check against the hashed password stored in the User struct.

Returns:

	(error): An error if the password comparison fails.
*/
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
