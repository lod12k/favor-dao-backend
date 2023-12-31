package core

import (
	"favor-dao-backend/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrganMangerService interface {
	GetOrganById(id primitive.ObjectID) (*model.Organ, error)
	GetOrganByKey(key string) (*model.Organ, error)
	GetOrganNotShow() (*[]primitive.ObjectID, error)
	ListOrgan() (*[]model.Organ, error)
}
