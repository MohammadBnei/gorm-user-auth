package service

import (
	"github.com/MohammadBnei/go-user-auth/model"
	"gorm.io/gorm"
)

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

/*
GetUserByEmail retrieves a user from the database by their email address.

Parameters:
- email (string): the email address of the user to retrieve.

Returns:
- (*model.User): a pointer to the User object representing the retrieved user.
- (error): an error object, which is non-nil if an error occurred during the retrieval.

Example usage:
u, err := userService.GetUserByEmail("alice@example.com")

	if err != nil {
		log.Fatalf("Failed to retrieve user: %v", err)
	}

fmt.Printf("Retrieved user: %#v\n", u)
*/
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

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

func (s *UserService) DeleteUser(id int) error {
	return s.db.Delete(&model.User{}, id).Error
}

/*
UpdateUser updates a User with the given id in the UserService's database.

Parameters:

  - id (int): the id of the User to update
  - data (*model.UserUpdateDTO): a pointer to a UserUpdateDTO containing the data to update the User with

Returns:

  - error: if any error occurred during the update
*/
func (s *UserService) UpdateUser(id int, data *model.UserUpdateDTO) (*model.User, error) {
	user, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}

	user.Email = data.Email

	err = s.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}
