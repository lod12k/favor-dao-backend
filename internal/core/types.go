package core

import "favor-dao-backend/internal/model"

type (
	User                 = model.User
	Dao                  = model.Dao
	Post                 = model.Post
	ConditionsT          = model.ConditionsT
	PostFormatted        = model.PostFormatted
	UserFormatted        = model.UserFormatted
	PostContentFormatted = model.PostContentFormatted
)

const (
	PostVisitPublic  = model.PostVisitPublic
	PostVisitPrivate = model.PostVisitPrivate

	PostMember1 = model.PostMember1

	DaoSubscribeSuccess = model.DaoSubscribeSuccess
	DaoSubscribeFailed  = model.DaoSubscribeFailed
	DaoSubscribeSubmit  = model.DaoSubscribeSubmit
	DaoSubscribeRefund  = model.DaoSubscribeRefund
)

var (
	AllQueryPostType = []model.PostType{model.SMS, model.VIDEO}
)

type (
	PostVisibleT  = model.PostVisibleT
	PostType      = model.PostType
	DaoSubscribeT = model.DaoSubscribeT
)
