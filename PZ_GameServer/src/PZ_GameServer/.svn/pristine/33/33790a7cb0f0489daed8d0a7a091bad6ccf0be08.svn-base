package model

import "time"

type Suggestion struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	UserID  int    `json:"user_id" sql:"index"`
	Title   string `json:"question"`
	Content string `json:"answer"`
}

type SuggestionModel struct {
	CommonModel
}

func GetSuggestionModel() *SuggestionModel {
	return &SuggestionModel{CommonModel{db: commonDb}}
}
