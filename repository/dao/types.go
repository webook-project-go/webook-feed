package dao

import "time"

type Inbox struct {
	ID       int64 `gorm:"primaryKey;autoIncrement"`
	TargetID int64 `gorm:"not null;index:idx_user_typ,priority:1"`
	ActorID  int64
	MetaData []byte `gorm:"type:Blob"`
	Type     uint8  `gorm:"not null;index:idx_user_typ, priority:2"`
	Ctime    int64  `gorm:"autoCreateTime:milli"`
	Utime    int64  `gorm:"autoUpdateTime:milli"`
}

type OutBox struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	ActorID  int64  `gorm:"not null;index:idx_actor_time_typ, priority:1"`
	MetaData []byte `gorm:"type:Blob"`
	Type     uint8  `gorm:"not null;index:idx_actor_time_typ,priority:3"`
	Ctime    int64  `gorm:"autoCreateTime:index:idx_actor_time_typ,priority:2"`
	Utime    int64  `gorm:"autoUpdateTime:milli"`
}

type PushTask struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	ActorID   int64      `gorm:"not null;index"`
	MetaData  []byte     `gorm:"type:Blob"`
	Type      uint8      `gorm:"not null;index"`
	ExecuteAt *time.Time `gorm:"index:exec_status,priority:1"`
	Status    uint8      `gorm:"index:exec_status,priority:2"`
	Ctime     int64      `gorm:"autoUpdateTime:milli"`
	Utime     int64      `gorm:"autoUpdateTime:milli"`
}
