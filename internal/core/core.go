package core

type DataService interface {
	TopicService
	IndexPostsService

	TweetService
	TweetManageService
	TweetHelpService

	CommentService
	CommentManageService

	UserManageService

	DaoManageService

	MsgMangerService
	MsgReadMangerService
	MsgSendMangerService
	MsgSysMangerService

	OrganMangerService
}
