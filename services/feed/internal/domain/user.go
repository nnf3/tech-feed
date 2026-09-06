package domain

import "strings"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (u User) Normalized() (User, error) {
	u.ID = strings.TrimSpace(u.ID)
	u.Email = strings.TrimSpace(u.Email)
	u.Name = strings.TrimSpace(u.Name)
	if u.ID == "" {
		return User{}, ErrUserIDRequired
	}
	return u, nil
}
