package users

func CreateUser(Mail string, Password string) (*User, error) {
	return nil, nil
}

func GetUserByMail(Mail string) (*User, error) {
	return nil, nil
}

func GetUserById(Id string) (*User, error) {
	return nil, nil
}

func (u User) CheckPassword(Password string) (bool, error) {
	return false, nil
}

func (u User) UpdatePassword(Password string) error {
	return nil
}

func (u User) CreateUserSession() (*UserSession, error) {
	return nil, nil
}

func GetUserSessionByToken(Token string) (*User, *UserSession, error) {
	return nil, nil, nil
}
