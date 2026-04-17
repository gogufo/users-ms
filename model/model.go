package model

import "gorm.io/gorm"

type Users struct {
	gorm.Model

	UID      string `gorm:"column:uid;type:varchar(60);UNIQUE;NOT NULL;" json:"uid"`
	Email    string `gorm:"column:email;type:varchar(254);UNIQUE;DEFAULT '';" json:"email"`
	Username string `gorm:"column:username;type:varchar(60);UNIQUE;DEFAULT '';" json:"username"`

	IsActive   bool `gorm:"column:is_active;type:bool;DEFAULT true;" json:"is_active"`
	IsVerified bool `gorm:"column:is_verified;type:bool;DEFAULT false;" json:"is_verified"`
}

type UsersProfile struct {
	gorm.Model

	UID string `gorm:"column:uid;type:varchar(60);UNIQUE;NOT NULL;" json:"uid"`

	FirstName  string `gorm:"column:first_name;type:varchar(254);DEFAULT '';" json:"first_name"`
	LastName   string `gorm:"column:last_name;type:varchar(254);DEFAULT '';" json:"last_name"`
	MiddleName string `gorm:"column:middle_name;type:varchar(254);DEFAULT '';" json:"middle_name"`

	Phone    string `gorm:"column:phone;type:varchar(60);DEFAULT '';" json:"phone"`
	AvatarID string `gorm:"column:avatar_id;type:varchar(60);DEFAULT '';" json:"avatar_id"`

	BirthDate int64 `gorm:"column:birth_date;type:bigint;DEFAULT 0;" json:"birth_date"`
}

type UsersSettings struct {
	gorm.Model

	UID string `gorm:"column:uid;type:varchar(60);UNIQUE;NOT NULL;" json:"uid"`

	Theme      string `gorm:"column:theme;type:varchar(60);DEFAULT 'light';" json:"theme"`
	Language   string `gorm:"column:language;type:varchar(10);DEFAULT 'en';" json:"language"`
	DateFormat string `gorm:"column:date_format;type:varchar(60);DEFAULT '2006-01-02';" json:"date_format"`
}

type UserNotificationPreferences struct {
	gorm.Model

	UID       string `gorm:"column:uid;type:varchar(60);NOT NULL;" json:"uid"`
	EventType string `gorm:"column:event_type;type:varchar(100);NOT NULL;" json:"event_type"` // e.g. "auth.login", "invoice.created"
	Channel   string `gorm:"column:channel;type:varchar(50);NOT NULL;" json:"channel"`        // email, sms, push, telegram, whatsapp, etc.

	Enabled bool `gorm:"column:enabled;type:bool;DEFAULT true;" json:"enabled"`
}
