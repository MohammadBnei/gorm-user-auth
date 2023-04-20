# Golang User & Auth

This is a simple authentication server with a user database.

This project will use mysql and [gorm](https://gorm.io/docs).

## Setup

Create a github repository, and copy it's url (removing the *https://* part). Then, we'll the repo name to setup the golang project :

```sh
export REPO_NAME=#put your repository name here
go mod init $REPO_NAME
```

It's useful to synchronize the name of the repo with the go.mod file so that your project is automatically packaged, and installable with **go install** (or go get).

You should see at the root of the directory the go.mod file, with the correct name in it. This file will contain all the dependencies of your project.

Don't forget to run **```go mod tidy```** to sync your source code with your dependencies.

## The folder structure

We will divide our code with logical slices, and we will have the following folders :
 - Model : this package will contain all the models used with the database orm.


## The User model

We need to create a user with basic data (id, email, password and timestamps). 
Create a new file called **model/user.go**, and initialize it with the correct package name.

Here, we will define our User model, and add to it the hooks to handle timestamps and password hash (**we never save the password in plain text !**)

This is the code for the User struct 
```go
type User struct {
	gorm.Model
	Email    string `json:"email"`
	Password string `json:"password"`
}
```

Next, we write the hooks to handle user specific logic
```go
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
```
Using [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt), we set up the hooks to hash the password upon new creation. We also handle the timestamps this way.

Finally, as the password is hashed, we need to provide a function to test a plaintext password 
```go
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
```
This function returns an error if the passwords do not match, nil otherwise.

## Database Connection



