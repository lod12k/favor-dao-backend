package model

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrDuplicateNickname = errors.New("user nickname duplicate")
)

type User struct {
	ID         primitive.ObjectID `json:"id"               bson:"_id,omitempty"`
	CreatedOn  int64              `json:"created_on"       bson:"created_on"`
	ModifiedOn int64              `json:"modified_on"      bson:"modified_on"`
	DeletedOn  int64              `json:"deleted_on"       bson:"deleted_on"`
	IsDel      int                `json:"is_del"           bson:"is_del"`
	Nickname   string             `json:"nickname"         bson:"nickname"`
	Address    string             `json:"address"          bson:"address"`
	Avatar     string             `json:"avatar"           bson:"avatar"`
	Role       string             `json:"role"             bson:"role"`
	Token      string             `json:"token"            bson:"token"`
	LoginAt    int64              `json:"login_at"         bson:"login_at"`
}

type UserFormatted struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

func (m *User) Format() *UserFormatted {
	return &UserFormatted{
		ID:       m.ID.Hex(),
		Nickname: m.Nickname,
		Address:  m.Address,
		Avatar:   m.Avatar,
		Role:     m.Role,
	}
}

func (m *User) Table() string {
	return "d_user"
}

func (m *User) Get(ctx context.Context, db *mongo.Database) (*User, error) {
	var (
		user User
		res  *mongo.SingleResult
	)
	if !m.ID.IsZero() {
		filter := bson.D{{"_id", m.ID}, {"is_del", 0}}
		res = db.Collection(m.Table()).FindOne(ctx, filter)
	} else if m.Address != "" {
		filter := bson.D{{"address", m.Address}, {"is_del", 0}}
		res = db.Collection(m.Table()).FindOne(ctx, filter)
	}
	err := res.Err()
	if err != nil {
		return &user, err
	}
	err = res.Decode(&user)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

func (m *User) GetOne(db *mongo.Database, conditions *ConditionsT) (*User, error) {
	var (
		user  User
		res   *mongo.SingleResult
		query bson.M
	)
	if len(*conditions) == 0 {
		if query != nil {
			query = findQuery([]bson.M{query})
		} else {
			query = bson.M{"is_del": 0}
		}
	}
	for k, v := range *conditions {
		if k != "ORDER" {
			if query != nil {
				query = findQuery([]bson.M{query, v})
			} else {
				query = findQuery([]bson.M{v})
			}
		}

	}
	res = db.Collection(m.Table()).FindOne(context.TODO(), query)
	err := res.Err()
	if err != nil {
		return &user, err
	}
	err = res.Decode(&user)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

func (m *User) CheckNicknameDuplication(ctx context.Context, db *mongo.Database) bool {
	filter := bson.M{"nickname": m.Nickname, ID: bson.M{"$ne": m.ID}}
	res := db.Collection(m.Table()).FindOne(ctx, filter)
	if res.Err() == nil {
		return true
	}
	return false
}

func (m *User) List(db *mongo.Database, conditions *ConditionsT, offset, limit int) ([]*User, error) {
	var (
		users  []*User
		err    error
		cursor *mongo.Cursor
		query  bson.M
	)
	finds := make([]*options.FindOptions, 0, 3)
	if limit > 0 {
		finds = append(finds, options.Find().SetSkip(int64(offset)))
		finds = append(finds, options.Find().SetLimit(int64(limit)))
	}
	if len(*conditions) == 0 {
		if query != nil {
			query = findQuery([]bson.M{query})
		} else {
			query = bson.M{"is_del": 0}
		}
	}
	for k, v := range *conditions {
		if k != "ORDER" {
			if query != nil {
				query = findQuery([]bson.M{query, v})
			} else {
				query = findQuery([]bson.M{v})
			}
		} else {
			finds = append(finds, options.Find().SetSort(v))
		}
	}
	if cursor, err = db.Collection(m.Table()).Find(context.TODO(), query, finds...); err != nil {
		return nil, err
	}
	for cursor.Next(context.TODO()) {
		var user User
		if cursor.Decode(&user) != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (m *User) FindListByKeyword(ctx context.Context, db *mongo.Database, keyword string, offset, limit int) (users []*User, err error) {
	var filter bson.M
	if keyword != "" {
		filter = bson.M{
			"nickname": bson.M{"$regex": keyword},
		}
	}
	finds := make([]*options.FindOptions, 0, 3)
	finds = append(finds, options.Find().SetSkip(int64(offset)))
	finds = append(finds, options.Find().SetLimit(int64(limit)))
	finds = append(finds, options.Find().SetSort(bson.M{"address": 1}))
	cur, err := db.Collection(m.Table()).Find(ctx, filter, finds...)
	if err != nil {
		return
	}
	err = cur.All(ctx, &users)
	return
}

func (m *User) Create(ctx context.Context, db *mongo.Database) (*User, error) {
	now := time.Now().Unix()
	m.CreatedOn = now
	m.ModifiedOn = now
	res, err := db.Collection(m.Table()).InsertOne(ctx, &m)
	if err != nil {
		return nil, err
	}
	m.ID = res.InsertedID.(primitive.ObjectID)
	return m, nil
}

func (m *User) Update(ctx context.Context, db *mongo.Database) error {
	filter := bson.D{
		{"$or", bson.A{
			bson.M{"_id": m.ID},
			bson.M{"address": m.Address},
		},
		},
	}
	m.ModifiedOn = time.Now().Unix()
	res := db.Collection(m.Table()).FindOneAndReplace(ctx, filter, &m)
	return res.Err()
}

func (m *User) Delete(ctx context.Context, db *mongo.Database) error {
	filter := bson.D{
		{"$or", bson.A{
			bson.M{"_id": m.ID},
			bson.M{"address": m.Address},
		},
		},
	}
	m.DeletedOn = time.Now().Unix()
	res := db.Collection(m.Table()).FindOneAndReplace(ctx, filter, &m)
	return res.Err()
}
