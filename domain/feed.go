package domain

type EventType uint8

const (
	UnknownEvent EventType = iota
	ArticleEvent
	LikeEvent
	CommentEvent
)

func TypeFromUint8(typ uint8) EventType {
	switch {
	case typ == 1:
		return ArticleEvent
	case typ == 2:
		return LikeEvent
	case typ == 3:
		return CommentEvent
	default:
		return UnknownEvent
	}
}

type FeedEvent struct {
	ID       int64
	TargetID int64
	ActorID  int64
	Type     EventType
	MetaData map[string]string
	Ctime    int64
}
