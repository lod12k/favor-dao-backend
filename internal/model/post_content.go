package model

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Type, 1 title, 2 text paragraph, 3 picture address, 4 video address, 5 voice address, 6 link address, 7 attachment resource

type PostContentT int

const (
	CONTENT_TYPE_TITLE PostContentT = iota + 1
	CONTENT_TYPE_TEXT
	CONTENT_TYPE_IMAGE
	CONTENT_TYPE_VIDEO
	CONTENT_TYPE_AUDIO
	CONTENT_TYPE_LINK
	CONTENT_TYPE_ATTACHMENT
	CONTENT_TYPE_CHARGE_ATTACHMENT
)

var (
	mediaContentType = []PostContentT{
		CONTENT_TYPE_IMAGE,
		CONTENT_TYPE_VIDEO,
		CONTENT_TYPE_AUDIO,
		CONTENT_TYPE_ATTACHMENT,
		CONTENT_TYPE_CHARGE_ATTACHMENT,
	}
)

type PostContent struct {
	*Model
	PostID  int64        `json:"post_id"`
	UserID  int64        `json:"user_id"`
	Content string       `json:"content"`
	Type    PostContentT `json:"type"`
	Sort    int64        `json:"sort"`
}

type PostContentFormated struct {
	ID      int64        `json:"id"`
	PostID  int64        `json:"post_id"`
	Content string       `json:"content"`
	Type    PostContentT `json:"type"`
	Sort    int64        `json:"sort"`
}

func (p *PostContent) DeleteByPostId(db *mongo.Database, postId int64) error {
	return db.Model(p).Where("post_id = ?", postId).Updates(map[string]interface{}{
		"deleted_on": time.Now().Unix(),
		"is_del":     1,
	}).Error
}

func (p *PostContent) MediaContentsByPostId(db *mongo.Database, postId int64) (contents []string, err error) {
	err = db.Model(p).Where("post_id = ? AND type IN ?", postId, mediaContentType).Select("content").Find(&contents).Error
	return
}

func (p *PostContent) Create(db *mongo.Database) (*PostContent, error) {
	err := db.Create(&p).Error

	return p, err
}

func (p *PostContent) Format() *PostContentFormated {
	if p.Model == nil {
		return nil
	}
	return &PostContentFormated{
		ID:      p.ID,
		PostID:  p.PostID,
		Content: p.Content,
		Type:    p.Type,
		Sort:    p.Sort,
	}
}

func (p *PostContent) List(db *mongo.Database, conditions *ConditionsT, offset, limit int) ([]*PostContent, error) {
	var contents []*PostContent
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if p.PostID > 0 {
		db = db.Where("id = ?", p.PostID)
	}

	for k, v := range *conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&contents).Error; err != nil {
		return nil, err
	}

	return contents, nil
}

func (p *PostContent) Get(db *mongo.Database) (*PostContent, error) {
	var content PostContent
	if p.Model != nil && p.ID > 0 {
		db = db.Where("id = ? AND is_del = ?", p.ID, 0)
	} else {
		return nil, gorm.ErrRecordNotFound
	}

	err := db.First(&content).Error
	if err != nil {
		return &content, err
	}

	return &content, nil
}
