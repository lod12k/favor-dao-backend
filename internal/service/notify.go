package service

import (
	"favor-dao-backend/internal/model"
	"favor-dao-backend/pkg/errcode"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotifyGroup struct {
	FromInfo    FromInfo           `json:"fromInfo"`
	UnreadCount int64              `json:"unreadCount"`
	Content     string             `json:"content"`
	FromType    model.FromTypeEnum `json:"fromType"`
	CreatedAt   int64              `json:"createdAt"`
}

type FromInfo struct {
	ID     primitive.ObjectID `json:"id"`
	Avatar string             `json:"avatar"`
	Name   string             `json:"name"`
}

type NotifyOrgan struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Key         string             `json:"key"`
	Avatar      string             `json:"avatar"`
	UnreadCount int64              `json:"unreadCount"`
}

func NotifyGroupList(to primitive.ObjectID, pageSize, pageNum int) (*[]NotifyGroup, int64, *errcode.Error) {
	var readAt int64
	oids, err := ds.GetOrganNotShow()
	msl, err := ds.ListMsgSend(to, oids, pageSize, pageNum)
	if err != nil {
		return nil, 0, errcode.MsgSendListFailed
	}
	if msl == nil || len(*msl) == 0 {
		return nil, 0, nil
	}
	list := make([]NotifyGroup, 0, pageSize)
	count, err := ds.CountMsgSend(to, oids)
	if err != nil {
		return nil, 0, errcode.MsgSendCountFailed
	}
	for _, ms := range *msl {

		mr, err := ds.GetMsgRead(ms.From, to)
		if mr == nil {
			readAt = 0
		} else {
			readAt = mr.ReadAt
		}
		unReadCount, err := ds.CountUnreadMsg(ms.From, to, readAt)
		if err != nil {
			return nil, 0, errcode.GetMsgUnReadCountFailed
		}
		msi, err := ds.GetLastMsg(ms.From, to)
		if err != nil {
			return nil, 0, errcode.GetMsgFailed
		}
		msg, err := ds.GetMsgById(msi.MsgID)
		if err != nil {
			return nil, 0, errcode.MsgSendLastFailed
		}
		var fi FromInfo
		switch ms.FromType {
		case model.DAO_TYPE:
			dao := &model.Dao{ID: ms.From}
			dao, err := ds.GetDao(dao)
			if err != nil {
				return nil, 0, errcode.GetDaoFailed
			}
			fi.ID = dao.ID
			fi.Name = dao.Name
			fi.Avatar = dao.Avatar
		case model.USER:
			user, err := ds.GetUserById(ms.From)
			if err != nil {
				return nil, 0, errcode.NewError(000000, "User does not exist")
			}
			fi.ID = user.ID
			fi.Name = user.Nickname
			fi.Avatar = user.Avatar
		case model.ORANGE:
			o, err := ds.GetOrganById(ms.From)
			if err != nil {
				return nil, 0, errcode.GetOrganFailed
			}
			fi.ID = o.ID
			fi.Name = o.Name
			fi.Avatar = o.Avatar
		}
		ng := NotifyGroup{
			FromInfo:    fi,
			UnreadCount: unReadCount,
			FromType:    ms.FromType,
			Content: func() string {
				if msg.Title != "" {
					return msg.Title
				}
				return msg.Content
			}(),
			CreatedAt: msg.CreatedAt,
		}
		list = append(list, ng)
	}
	return &list, count, nil
}

func NotifyOrganList(to primitive.ObjectID) (*[]NotifyOrgan, *errcode.Error) {
	var (
		readAt int64
		c      int64
	)
	organs, err := ds.ListOrgan()
	if err != nil {
	}
	os := *organs
	nos := make([]NotifyOrgan, 0, len(os))
	for _, o := range os {
		if !to.IsZero() {
			mr, err := ds.GetMsgRead(o.ID, to)
			if err != nil {
				logrus.Errorf("get msg_read err: %v\n", err)
			}
			if mr == nil {
				readAt = 0
			} else {
				readAt = mr.ReadAt
			}
			if o.Key == "sys" {
				c, err = ds.CountUnreadSysMsg(readAt)
				if err != nil {
					logrus.Errorf("get unread_msg err: %v\n", err)
				}
			} else {
				c, err = ds.CountUnreadMsg(o.ID, to, readAt)
				if err != nil {
					logrus.Errorf("get unread_msg err: %v\n", err)
				}
			}
		}

		no := NotifyOrgan{
			ID:          o.ID,
			Key:         o.Key,
			Name:        o.Name,
			Avatar:      o.Avatar,
			UnreadCount: c,
		}
		nos = append(nos, no)
	}
	return &nos, nil
}

func NotifyByFrom(from, to primitive.ObjectID, pageSize, pageNum int) (*[]model.Msg, int64, *errcode.Error) {
	msl, err := ds.ListMsg(from, to, pageSize, pageNum)
	if err != nil {
		return nil, 0, errcode.MsgListFailed
	}
	if msl == nil || len(*msl) == 0 {
		return nil, 0, nil
	}
	count, err := ds.CountMsg(from, to)
	if err != nil {
		return nil, 0, errcode.MsgCountFailed
	}
	return msl, count, nil
}

func PutNotifyRead(from, to primitive.ObjectID) (bool, *errcode.Error) {
	mr, _ := ds.GetMsgRead(from, to)
	if mr == nil {
		_, err := ds.CreateMsgRead(from, to)
		if err != nil {
			return false, errcode.CreateMsgReadFailed
		}
		return true, nil
	}
	result, err := ds.UpdateReadAt(mr)
	if err != nil {
		return false, errcode.UpdateMsgReadFailed
	}
	return result, nil
}

func GetNotifyUnread(from, to primitive.ObjectID) int64 {
	ms, err := ds.GetMsgRead(from, to)
	if err != nil || ms == nil {
		return 0
	}
	count, _ := ds.CountUnreadMsg(from, to, ms.ReadAt)
	return count
}

func DeleteNotifyById(id primitive.ObjectID) (bool, *errcode.Error) {
	ms, err := ds.GetMsgSendByMsgId(id)
	if err != nil {
		return false, errcode.GetMsgSendFailed
	}
	_, err = ds.DeleteMsg(ms.MsgID)
	if err != nil {
		return false, errcode.DeleteMsgFailed
	}
	b, err := ds.DeleteMsgSendByMsgId(id)
	if err != nil {
		return false, errcode.DeleteMsgSendFailed
	}
	return b, nil
}

func DeleteNotifyByFrom(from, to primitive.ObjectID) (bool, *errcode.Error) {
	mss, err := ds.GetMsgSend(from, to)
	if err != nil {
		return false, errcode.GetMsgSendFailed
	}
	_, err = ds.DeleteMsgSend(from, to)
	if err != nil {
		return false, errcode.DeleteMsgSendFailed
	}
	_, err = ds.DeleteMsgRead(from, to)
	if err != nil {
		return false, errcode.DeleteMsgReadFailed
	}
	for _, ms := range *mss {
		b, err := ds.DeleteMsg(ms.MsgID)
		if err != nil || !b {
			return false, errcode.DeleteMsgFailed
		}
	}
	return true, nil
}

func NotifySys(from primitive.ObjectID, pageSize, pageNum int) (*[]model.MsgSys, int64, *errcode.Error) {
	list, err := ds.ListMsgSys(from, pageSize, pageNum)
	if err != nil {
		return nil, 0, errcode.MsgSysListFailed
	}
	count, err := ds.CountMsgSys(from)
	if err != nil {
		return nil, 0, errcode.MsgSysCountFailed
	}
	return list, count, nil
}
