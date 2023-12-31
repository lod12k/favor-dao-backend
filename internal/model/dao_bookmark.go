package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DaoBookmark struct {
	ID         primitive.ObjectID `json:"id"               bson:"_id,omitempty"`
	CreatedOn  int64              `json:"created_on"       bson:"created_on"`
	ModifiedOn int64              `json:"modified_on"      bson:"modified_on"`
	DeletedOn  int64              `json:"deleted_on"       bson:"deleted_on"`
	IsDel      int                `json:"is_del"           bson:"is_del"`
	Address    string             `json:"address"          bson:"address"`
	DaoID      primitive.ObjectID `json:"dao_id"           bson:"dao_id"`
}

func (m *DaoBookmark) Table() string {
	return "dao_bookmark"
}

func (m *DaoBookmark) Create(ctx context.Context, db *mongo.Database) (*DaoBookmark, error) {
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

func (m *DaoBookmark) Delete(ctx context.Context, db *mongo.Database) error {
	filter := bson.D{{"_id", m.ID}}
	update := bson.D{{"$set", bson.D{
		{"is_del", 1},
		{"deleted_on", time.Now().Unix()},
	}}}
	res := db.Collection(m.Table()).FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *DaoBookmark) Update(ctx context.Context, db *mongo.Database) error {
	filter := bson.D{{"_id", m.ID}}
	res := db.Collection(m.Table()).FindOneAndReplace(ctx, filter, &m)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *DaoBookmark) GetByAddress(ctx context.Context, db *mongo.Database, address string, daoId string, hasDel ...bool) (*DaoBookmark, error) {
	id, err := primitive.ObjectIDFromHex(daoId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"dao_id":  id,
		"address": address,
	}
	if len(hasDel) == 0 {
		filter["is_del"] = 0
	}
	res := db.Collection(m.Table()).FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var dao DaoBookmark
	err = res.Decode(&dao)
	if err != nil {
		return nil, err
	}
	return &dao, nil
}

func (m *DaoBookmark) Get(ctx context.Context, db *mongo.Database) (*DaoBookmark, error) {
	if m.ID.IsZero() {
		return nil, mongo.ErrNoDocuments
	}
	filter := bson.D{{"_id", m.ID}, {"is_del", 0}}
	res := db.Collection(m.Table()).FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var dao DaoBookmark
	err := res.Decode(&dao)
	if err != nil {
		return nil, err
	}
	return &dao, nil
}

func (m *DaoBookmark) FindList(ctx context.Context, db *mongo.Database, filter interface{}) (list []*DaoBookmark) {
	cursor, err := db.Collection(m.Table()).Find(ctx, filter)
	if err != nil {
		return
	}
	list = []*DaoBookmark{}
	for cursor.Next(context.TODO()) {
		var t DaoBookmark
		if cursor.Decode(&t) != nil {
			return
		}
		list = append(list, &t)
	}
	return
}

func (m *DaoBookmark) GetList(ctx context.Context, db *mongo.Database, pipeline interface{}) (list []*DaoFormatted) {
	list = []*DaoFormatted{}
	cursor, err := db.Collection(m.Table()).Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	type tmp struct {
		Dao *Dao `bson:"dao"`
	}

	for cursor.Next(context.TODO()) {
		var t tmp
		if cursor.Decode(&t) != nil {
			return
		}
		list = append(list, t.Dao.Format())
	}
	return
}

func (m *DaoBookmark) CountMark(ctx context.Context, db *mongo.Database) int64 {
	query := bson.M{
		"address": m.Address,
		"is_del":  0,
	}
	documents, err := db.Collection(m.Table()).CountDocuments(ctx, query)
	if err != nil {
		return 0
	}
	return documents
}
