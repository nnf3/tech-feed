package postgres

import (
	"time"

	"github.com/lib/pq"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type userRecord struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Email     string    `gorm:"column:email"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (userRecord) TableName() string { return "users" }

func userFromDomain(user domain.User) userRecord {
	return userRecord{ID: user.ID, Email: user.Email, Name: user.Name}
}

func (r userRecord) toDomain() domain.User {
	return domain.User{ID: r.ID, Email: r.Email, Name: r.Name}
}

type profileRecord struct {
	UserID       string         `gorm:"column:user_id;primaryKey"`
	InterestTags pq.StringArray `gorm:"column:interest_tags;type:text[]"`
	ExcludeTags  pq.StringArray `gorm:"column:exclude_tags;type:text[]"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

func (profileRecord) TableName() string { return "profiles" }

func profileFromDomain(profile domain.Profile) profileRecord {
	return profileRecord{
		UserID:       profile.UserID,
		InterestTags: nonNilTags(profile.InterestTags),
		ExcludeTags:  nonNilTags(profile.ExcludeTags),
	}
}

func (r profileRecord) toDomain() domain.Profile {
	return domain.Profile{
		UserID:       r.UserID,
		InterestTags: nonNilTags(r.InterestTags),
		ExcludeTags:  nonNilTags(r.ExcludeTags),
	}
}

func nonNilTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
