package models

type Post struct {
	Title     string `json:"title" bson:"title"`
	Content   string `json:"content" bson:"content"`
	Day       uint   `json:"day,omitempty" bson:"day"`
	Month     uint   `json:"month,omitempty" bson:"month"`
	Year      uint   `json:"year,omitempty" bson:"year"`
	CreatedAt string `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt,omitempty" bson:"updatedAt"`
}
