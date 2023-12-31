// Core service implement base gorm+mysql/postgresql/sqlite3.
// Jinzhu is the primary developer of gorm so use his name as
// pakcage name as a saluter.

package monogo

import (
	"favor-dao-backend/internal/conf"
	"favor-dao-backend/internal/core"
	"favor-dao-backend/internal/dao/cache"
	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

var (
	_ core.DataService = (*dataServant)(nil)
	_ core.VersionInfo = (*dataServant)(nil)
)

type dataServant struct {
	core.IndexPostsService
	core.TopicService
	core.TweetService
	core.TweetManageService
	core.TweetHelpService
	core.CommentService
	core.CommentManageService
	core.UserManageService
	core.DaoManageService
	core.MsgMangerService
	core.MsgSendMangerService
	core.MsgReadMangerService
	core.MsgSysMangerService
	core.OrganMangerService
}

func NewDataService() (core.DataService, core.VersionInfo) {
	// initialize CacheIndex if needed
	var (
		c core.CacheIndexService
		v core.VersionInfo
	)
	db := conf.MustMongoDB()

	i := newIndexPostsService(db)
	if conf.CfgIf("SimpleCacheIndex") {
		i = newSimpleIndexPostsService(db)
		c, v = cache.NewSimpleCacheIndexService(i)
	} else if conf.CfgIf("BigCacheIndex") {
		c, v = cache.NewBigCacheIndexService(i)
	} else {
		c, v = cache.NewNoneCacheIndexService(i)
	}
	logrus.Infof("use %s as cache index service by version: %s", v.Name(), v.Version())

	ds := &dataServant{
		IndexPostsService:    c,
		TopicService:         newTopicService(db),
		TweetService:         newTweetService(db),
		TweetManageService:   newTweetManageService(db, c),
		TweetHelpService:     newTweetHelpService(db),
		CommentService:       newCommentService(db),
		CommentManageService: newCommentManageService(db),
		UserManageService:    newUserManageService(db),
		DaoManageService:     newDaoManageService(db),
		MsgMangerService:     newMsgManageService(db),
		MsgReadMangerService: newMsgReadMangerService(db),
		MsgSendMangerService: newMsgSendMangerService(db),
		MsgSysMangerService:  newMsgSysMangerService(db),
		OrganMangerService:   newOrganMangerService(db),
	}
	return ds, ds
}

func NewAuthorizationManageService() core.AuthorizationManageService {
	return &authorizationManageServant{
		db: conf.MustMongoDB(),
	}
}

func (s *dataServant) Name() string {
	return "Mongo"
}

func (s *dataServant) Version() *semver.Version {
	return semver.MustParse("v0.1.0")
}
