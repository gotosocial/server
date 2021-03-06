/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package gtsmodel

import "time"

// Notification models an alert/notification sent to an account about something like a reblog, like, new follow request, etc.
type Notification struct {
	// ID of this notification in the database
	ID string `pg:"type:CHAR(26),pk,notnull"`
	// Type of this notification
	NotificationType NotificationType `pg:",notnull"`
	// Creation time of this notification
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Which account does this notification target (ie., who will receive the notification?)
	TargetAccountID string `pg:"type:CHAR(26),notnull"`
	// Which account performed the action that created this notification?
	OriginAccountID string `pg:"type:CHAR(26),notnull"`
	// If the notification pertains to a status, what is the database ID of that status?
	StatusID string `pg:"type:CHAR(26)"`
	// Has this notification been read already?
	Read bool

	/*
		NON-DATABASE fields
	*/

	// gts model of the target account, won't be put in the database, it's just for convenience when passing the notification around.
	GTSTargetAccount *Account `pg:"-"`
	// gts model of the origin account, won't be put in the database, it's just for convenience when passing the notification around.
	GTSOriginAccount *Account `pg:"-"`
	// gts model of the relevant status, won't be put in the database, it's just for convenience when passing the notification around.
	GTSStatus *Status `pg:"-"`
}

// NotificationType describes the reason/type of this notification.
type NotificationType string

const (
	// NotificationFollow -- someone followed you
	NotificationFollow NotificationType = "follow"
	// NotificationFollowRequest -- someone requested to follow you
	NotificationFollowRequest NotificationType = "follow_request"
	// NotificationMention -- someone mentioned you in their status
	NotificationMention NotificationType = "mention"
	// NotificationReblog -- someone boosted one of your statuses
	NotificationReblog NotificationType = "reblog"
	// NotificationFave -- someone faved/liked one of your statuses
	NotificationFave NotificationType = "favourite"
	// NotificationPoll -- a poll you voted in or created has ended
	NotificationPoll NotificationType = "poll"
	// NotificationStatus -- someone you enabled notifications for has posted a status.
	NotificationStatus NotificationType = "status"
)
