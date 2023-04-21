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
```

Using [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt), we set up the hooks to hash the password upon new creation. We also handle the timestamps this way.

	Exercice
	
	Write a Before save hook to handle the update of the updatedAt timestamp. 
	At the same time, verify if the password field is being updated. If so, re-hash it.

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

We want to initialize a database connection on our api startup. For that, we will create two files in a config folder. One will be to load variable from the environnement (aka .env), the other for the mysql database connection.

Let's start with the latter. Create the **config/database.go** file :
```go
/*
InitDB initializes a GORM database connection using the provided Config.

Parameters:
- config (*Config): A pointer to the Config struct containing database connection details.

Returns:
- (*gorm.DB): A pointer to the GORM database object.
- (error): An error object if the connection fails, nil otherwise.
*/
func InitDB() (*gorm.DB, error) {
	dsn := "root:rootme@tcp(127.0.0.1:3306)/go_user_auth?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
```
For now, we will leave the connection string as it is. If you didn't do it already, spin up a mariadb instance using the provided compose.yml file.

Next, we will want to load the connection informations from the environnement.
Let's write in **config/env.yml** :
```go
import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_HOST string
	DB_USER string
	DB_PASS string
	DB_PORT string
	DB_NAME string
}

func InitConfig() *Config {
	godotenv.Load()

	return &Config{
		DB_HOST: os.Getenv("DB_HOST"),
		DB_USER: os.Getenv("DB_USER"),
		DB_PASS: os.Getenv("DB_PASS"),
		DB_PORT: os.Getenv("DB_PORT"),
		DB_NAME: os.Getenv("DB_NAME"),
	}
}
```
We will be using the [godotenv](https://github.com/joho/godotenv) package to automatically read from a .env file.
The env file will look like this 
```sh
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=rootme
DB_NAME=go_user_auth
```

Let's modify our **InitDB** function to take the incoming config.
First, add a parameter to the function 
```go
func InitDB(config *Config) (*gorm.DB, error)
```

Then, format the dsn string to interpolate with the config 
```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_PORT, config.DB_NAME)
```

Let's test this.

Create a **main.go** in the root directory, and tie together our work done so far

```go
func main() {
	conf := config.InitConfig()
	db, err := config.InitDB(conf)
	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&model.User{})
}
```

Don't forget to run ```go mod tidy``` once in a while.

To run our little api, type ```go run main.go``` in your terminal. If you don't see any error, you should be good to continue.

## User Service

We will expose our CRUD User service in a struct, called UserService. We will then have a layer of isolation between our controller and our ORM, which is always a good practice.

Inside **service/user.go**, we will initialize our user service by creating a struct with the according New function 
```go
type UserService struct {
	db *gorm.DB
}

/*
NewUserService returns a new instance of the UserService struct with the provided gorm.DB instance
as its database connection.

Parameters:

- db (*gorm.DB): The gorm.DB instance to use as the database connection.

Returns:

- (*UserService): A pointer to the newly created UserService instance.
*/
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}
```


Now, let's write our basic Read functions :
```go
/*
GetUser retrieves a user by ID from the database.

Parameters:

	s - a pointer to a UserService instance
	id - the ID of the user to retrieve

Return values:

	*model.User - a pointer to the retrieved user object
	error - if any error occurs while retrieving the user, it is returned here
*/
func (s *UserService) GetUser(id int) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

/*
GetUsers retrieves all users from the database.

Returns:

  - []*model.User: A slice of user objects.
  - error: An error object if the query fails.
*/
func (s *UserService) GetUsers() ([]*model.User, error) {
	var users []*model.User
	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}
```

	Exercice

	Write a delete function that takes in an id, deletes the record, and returns the deleted record. Don't forget to handle the possible errors (at least not found).

	Write a GetByEmail that fetches user by email. As we did not specify uniqueness on the email field, these search would possibly return multiple records.

To create and update a user, we will create the according DTO. These struct will help us maintain a clean code, while providing knowledge of the API later with swagger.

In **model/userDTO.go**, we will setup the create and update object. 
```go
type UserCreateDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUpdateDTO struct {
	Email string `json:"email"`
}
```

For good practice, we will disallow the user from directly update his password with the CRUD update function.

With these DTO, let's write the create function
```go
/*
CreateUser creates a new user in the UserService database.

Args:

  - s (*UserService): A pointer to the UserService instance.
  - data (*model.UserCreateDTO): A pointer to the data used to create the new user.

Returns:

  - (*model.User): A pointer to the newly created user.
  - (error): An error if the creation failed.
*/
func (s *UserService) CreateUser(data *model.UserCreateDTO) (*model.User, error) {
	user := &model.User{
		Email:    data.Email,
		Password: data.Password,
	}
	err := s.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}
```

	Exercice

	Write the update function. This function should return the updated user.

## Gin Handler

We have our service, our model and our database connection. Let's add the last functionnal part by creating the controller, aka gin handler.
This controller will be in charge of taking the incomming request, parsing it's body and parameters, and call the appropriate service function.

In a **handler/user.go** file, proceed to write the following initializer 
```go
type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
```

A basic get function will look like this
```go
/*
GetUser gets a user by their ID from the userService and returns it in the response body.

Parameters:
  - c (*gin.Context): the context of the current HTTP request
  - h (*UserHandler): the handler that handles user-related requests

Errors:
  - 400 Bad Request: if the parameter id cannot be converted to an integer, or if there is an error retrieving the user
*/
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, user)
}
```
The id is retrieved from the route, and we will implement the router now to test that our api starts.

Finally, add the following to the main and try starting your api 
```go
userService := service.NewUserService(db)
userHandler := handler.NewUserHandler(userService)

r := gin.Default()

userApi := r.Group("/api/v1/user")
userApi.GET("/:id", userHandler.GetUser)

r.Run()
```

If you can ```curl localhost:8080/api/v1/user/2``` and get a *{"error":"record not found"}*, everything is good.

	Exercice

	Write the rest of the CRUD functions to the service. Add the according route and test it out.