package monogo

import (
	"context"
	"errors"
	"strings"
	"time"

	"favor-dao-backend/internal/core"
	"favor-dao-backend/internal/model"
	chatModel "favor-dao-backend/internal/model/chat"
	"favor-dao-backend/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	_ core.DaoManageService = (*daoManageServant)(nil)
)

type daoManageServant struct {
	db *mongo.Database
}

func newDaoManageService(db *mongo.Database) core.DaoManageService {
	return &daoManageServant{
		db: db,
	}
}

func (s *daoManageServant) GetDaoByKeyword(keyword string) ([]*model.Dao, error) {
	dao := &model.Dao{}
	keyword = strings.Trim(keyword, " ")
	return dao.FindListByKeyword(context.TODO(), s.db, keyword, 0, 6)
}

func (s *daoManageServant) CreateDao(dao *model.Dao, chatAction func(context.Context, *model.Dao) (string, error)) (*model.Dao, error) {
	err := util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		// check name duplicate
		if dao.CheckNameDuplication(ctx, s.db) {
			return model.ErrDuplicateDAOName
		}

		newDao, err := dao.Create(ctx, s.db)
		if err != nil {
			return err
		}
		book := &model.DaoBookmark{Address: newDao.Address, DaoID: newDao.ID}
		out, err := book.GetByAddress(ctx, s.db, newDao.Address, newDao.ID.Hex(), true)
		if err != nil {
			out, err = book.Create(ctx, s.db)
		} else {
			out.IsDel = 0
			out.ModifiedOn = time.Now().Unix()
			out.DeletedOn = 0
			err = out.Update(ctx, s.db)
		}
		if err != nil {
			return err
		}
		groupId, err := chatAction(ctx, newDao)
		if err != nil {
			return err
		}
		group := &chatModel.Group{
			ID: chatModel.GroupID{
				DaoID:   newDao.ID,
				GroupID: groupId,
				OwnerID: newDao.Address,
			},
		}
		_, err = group.Create(ctx, s.db)
		if err != nil {
			return err
		}
		dao = newDao
		return nil
	})

	if err != nil {
		return nil, err
	}

	return dao, nil
}

func (s *daoManageServant) UpdateDao(dao *model.Dao, chatAction func(context.Context, *model.Dao) error) error {
	return util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		// check name duplicate
		if dao.CheckNameDuplication(ctx, s.db) {
			return model.ErrDuplicateDAOName
		}

		err := dao.Update(ctx, s.db)
		if err != nil {
			return err
		}
		return chatAction(ctx, dao)
	})
}

func (s *daoManageServant) DeleteDao(dao *model.Dao) error {
	return dao.Delete(context.TODO(), s.db)
}

func (s *daoManageServant) GetDaoCount(conditions model.ConditionsT) (int64, error) {
	return (&model.Dao{}).Count(s.db, conditions)
}

func (s *daoManageServant) GetDaoList(conditions model.ConditionsT, offset, limit int) ([]*model.Dao, error) {
	return (&model.Dao{}).List(s.db, conditions, offset, limit)
}

func (s *daoManageServant) GetDaoByName(dao *model.Dao) (*model.Dao, error) {
	return dao.GetByName(context.TODO(), s.db)
}

func (s *daoManageServant) GetDao(dao *model.Dao) (*model.Dao, error) {
	return dao.Get(context.TODO(), s.db)
}

func (s *daoManageServant) GetMyDaoList(dao *model.Dao) (list []*model.DaoFormatted, err error) {
	list, err = dao.GetListByAddress(context.TODO(), s.db)
	if err != nil {
		return
	}
	for k := range list {
		list[k].IsJoined = true
		list[k].IsSubscribed = true
	}
	return
}

func (s *daoManageServant) DaoBookmarkCount(address string) int64 {
	book := &model.DaoBookmark{Address: address}
	return book.CountMark(context.TODO(), s.db)
}

func (s *daoManageServant) GetDaoBookmarkList(userAddress string, q *core.QueryReq, offset, limit int) (list []*model.DaoFormatted) {
	query := bson.M{
		"address": userAddress,
		"is_del":  0,
	}
	dao := &model.Dao{}
	if q.Query != "" {
		query[dao.Table()+".name"] = bson.M{"$regex": q.Query}
	}
	if len(q.Addresses) > 0 {
		query[dao.Table()+".address"] = bson.M{"$in": q.Addresses}
	}
	pipeline := mongo.Pipeline{
		{{"$lookup", bson.M{
			"from":         dao.Table(),
			"localField":   "dao_id",
			"foreignField": "_id",
			"as":           "dao",
		}}},
		{{"$match", query}},
		{{"$sort", bson.M{"dao.follow_count": -1}}},
		{{"$skip", offset}},
		{{"$limit", limit}},
		{{"$unwind", "$dao"}},
	}
	book := &model.DaoBookmark{Address: userAddress}
	list = book.GetList(context.TODO(), s.db, pipeline)
	for k := range list {
		list[k].IsJoined = true
	}
	return
}

func (s *daoManageServant) GetDaoBookmarkListByAddress(address string) []*model.DaoBookmark {
	book := &model.DaoBookmark{}
	return book.FindList(context.TODO(), s.db, bson.M{"address": address, "is_del": 0})
}

func (s *daoManageServant) GetDaoBookmarkByAddressAndDaoID(myAddress string, daoId string) (*model.DaoBookmark, error) {
	book := &model.DaoBookmark{}
	res, err := book.GetByAddress(context.TODO(), s.db, myAddress, daoId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *daoManageServant) CreateDaoFollow(myAddress string, daoID string, chatAction func(context.Context, *model.Dao) (string, error)) (out *model.DaoBookmark, err error) {
	id, err := primitive.ObjectIDFromHex(daoID)
	if err != nil {
		return nil, err
	}

	err = util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		dao, err := (&model.Dao{ID: id, IsDel: 1}).Get(ctx, s.db)
		if err != nil {
			return err
		}
		dao.FollowCount++
		err = dao.Update(ctx, s.db)
		if err != nil {
			return err
		}
		// update book
		book := &model.DaoBookmark{Address: myAddress, DaoID: id}
		out, err = book.GetByAddress(ctx, s.db, myAddress, daoID, true)
		if err != nil {
			out, err = book.Create(ctx, s.db)
		} else {
			out.IsDel = 0
			out.ModifiedOn = time.Now().Unix()
			out.DeletedOn = 0
			err = out.Update(ctx, s.db)
		}
		groupId, err := chatAction(ctx, dao)
		if err != nil {
			return err
		}
		group := &chatModel.Group{
			ID: chatModel.GroupID{
				DaoID:   dao.ID,
				GroupID: groupId,
				OwnerID: myAddress,
			},
		}
		_, err = group.Create(ctx, s.db)
		return err
	})
	return
}

func (s *daoManageServant) DeleteDaoFollow(d *model.DaoBookmark, chatAction func(context.Context, *model.Dao) (string, error)) error {
	return util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		dao, err := (&model.Dao{ID: d.DaoID, IsDel: 1}).Get(ctx, s.db)
		if err != nil {
			return err
		}
		dao.FollowCount--
		err = dao.Update(ctx, s.db)
		if err != nil {
			return err
		}
		err = d.Delete(ctx, s.db)
		if err != nil {
			return err
		}
		groupId, err := chatAction(ctx, dao)
		if err != nil {
			return err
		}
		group := &chatModel.Group{
			ID: chatModel.GroupID{
				DaoID:   d.DaoID,
				GroupID: groupId,
				OwnerID: d.Address,
			},
		}
		return group.Delete(ctx, s.db)
	})
}

func (s *daoManageServant) RealDeleteDAO(address string, chatAction func(context.Context, *model.Dao) (string, error)) error {
	return util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		cursor, err := s.db.Collection(new(model.Dao).Table()).Find(ctx, bson.M{"address": address})
		if err != nil {
			return err
		}
		for cursor.Next(ctx) {
			var d model.Dao
			if cursor.Decode(&d) != nil {
				return err
			}
			groupId, err := chatAction(ctx, &d)
			if err != nil {
				return err
			}
			err = d.RealDelete(ctx, s.db, groupId)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *daoManageServant) IsJoinedDAO(address string, daoID primitive.ObjectID) bool {
	if address == "" {
		return false
	}
	filter := bson.M{
		"address": address,
		"dao_id":  daoID,
		"is_del":  0,
	}
	res := s.db.Collection(new(model.DaoBookmark).Table()).FindOne(context.TODO(), filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return false
	}
	if res.Err() != nil {
		return false
	}
	return true
}

func (s *daoManageServant) IsSubscribeDAO(address string, daoID primitive.ObjectID) bool {
	if address == "" {
		return false
	}
	filter := bson.M{
		"address": address,
		"dao_id":  daoID,
		"status":  model.DaoSubscribeSuccess,
	}
	res := s.db.Collection(new(model.DaoSubscribe).Table()).FindOne(context.TODO(), filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return false
	}
	if res.Err() != nil {
		return false
	}
	return true
}

func (s *daoManageServant) SubscribeDAO(address string, daoID primitive.ObjectID, fn func(ctx context.Context, orderID string, dao *model.Dao) error) error {
	return util.MongoTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		dao, err := (&model.Dao{ID: daoID}).Get(ctx, s.db)
		if err != nil {
			return err
		}
		sub := &model.DaoSubscribe{
			Address:   address,
			DaoID:     daoID,
			Status:    model.DaoSubscribeSubmit,
			PayAmount: dao.Price,
		}
		err = sub.Create(ctx, s.db)
		if err != nil {
			return err
		}
		if fn != nil {
			return fn(ctx, sub.ID.Hex(), dao)
		}
		return nil
	})
}

func (s *daoManageServant) UpdateSubscribeDAO(orderID, txID string, status model.DaoSubscribeT) error {
	id, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}
	md := model.DaoSubscribe{}
	filter := bson.M{
		"_id": id,
	}
	update := bson.D{{"$set", bson.D{
		{"modified_on", time.Now().Unix()},
		{"tx_id", txID},
		{"status", status},
	}}}
	res := s.db.Collection(md.Table()).FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (s *daoManageServant) UpdateSubscribeDAOTxID(orderID, txID string) error {
	id, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}
	md := model.DaoSubscribe{}
	filter := bson.M{
		"_id": id,
	}
	update := bson.D{{"$set", bson.D{
		{"modified_on", time.Now().Unix()},
		{"tx_id", txID},
	}}}
	res := s.db.Collection(md.Table()).FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}
