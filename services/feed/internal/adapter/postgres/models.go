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

type bookmarkRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	ArticleID string    `gorm:"column:article_id;primaryKey"`
	URL       string    `gorm:"column:url"`
	Title     string    `gorm:"column:title"`
	Source    string    `gorm:"column:source"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (bookmarkRecord) TableName() string { return "bookmarks" }

func bookmarkFromDomain(bookmark domain.Bookmark) bookmarkRecord {
	return bookmarkRecord{
		UserID:    bookmark.UserID,
		ArticleID: bookmark.ArticleID,
		URL:       bookmark.URL,
		Title:     bookmark.Title,
		Source:    bookmark.Source,
	}
}

func (r bookmarkRecord) toDomain() domain.Bookmark {
	return domain.Bookmark{
		UserID:    r.UserID,
		ArticleID: r.ArticleID,
		URL:       r.URL,
		Title:     r.Title,
		Source:    r.Source,
		CreatedAt: r.CreatedAt,
	}
}

type historyRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	ArticleID string    `gorm:"column:article_id;primaryKey"`
	URL       string    `gorm:"column:url"`
	Title     string    `gorm:"column:title"`
	Source    string    `gorm:"column:source"`
	ViewedAt  time.Time `gorm:"column:viewed_at"`
	ViewCount int       `gorm:"column:view_count"`
}

func (historyRecord) TableName() string { return "article_views" }

func historyFromDomain(entry domain.HistoryEntry) historyRecord {
	return historyRecord{
		UserID:    entry.UserID,
		ArticleID: entry.ArticleID,
		URL:       entry.URL,
		Title:     entry.Title,
		Source:    entry.Source,
		ViewCount: entry.ViewCount,
	}
}

func (r historyRecord) toDomain() domain.HistoryEntry {
	return domain.HistoryEntry{
		UserID:    r.UserID,
		ArticleID: r.ArticleID,
		URL:       r.URL,
		Title:     r.Title,
		Source:    r.Source,
		ViewedAt:  r.ViewedAt,
		ViewCount: r.ViewCount,
	}
}

func nonNilTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
