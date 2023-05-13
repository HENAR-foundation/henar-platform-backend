package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	Admin      Role = "admin"
	Specialist Role = "specialist"
)

type UserCredentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Contacts struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Facebook  string `json:"facebook"`
	Instagram string `json:"instagram"`
	Linkedin  string `json:"linkedin"`
}

type UserBody struct {
	ID                        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Email                     string               `json:"email" validate:"required,email"`
	Password                  string               `json:"password" validate:"required"`
	Avatar                    string               `json:"avatar"`
	FullName                  string               `json:"full_name"`
	Description               string               `json:"description"`
	Contacts                  []string             `json:"contacts"`
	Location                  primitive.ObjectID   `json:"location"`
	Role                      Role                 `json:"role"`
	Job                       string               `json:"job"`
	Tags                      []primitive.ObjectID `json:"tags"`
	IncomingContactRequests   []primitive.ObjectID `json:"incoming_contact_requests,omitempty" bson:"incoming_contact_requests,omitempty"`
	OutgoingContactRequests   []primitive.ObjectID `json:"outgoing_contact_requests,omitempty" bson:"outgoing_contact_requests,omitempty"`
	ConfirmedContactsRequests []primitive.ObjectID `json:"confirmed_contacts_requests,omitempty" bson:"confirmed_contacts_requests,omitempty"`
	BlockedUsers              []primitive.ObjectID `json:"blocked_users,omitempty" bson:"blocked_users,omitempty"`
	ProjectsApplications      []primitive.ObjectID `json:"projects_applications,omitempty" bson:"projects_applications,omitempty"`
	ConfirmedApplications     []primitive.ObjectID `json:"confirmed_applications,omitempty" bson:"confirmed_applications,omitempty"`
	ProjectsHistory           []primitive.ObjectID `json:"projects_history,omitempty" bson:"projects_history,omitempty"`
	CreatedProjects           []primitive.ObjectID `json:"created_projects,omitempty" bson:"created_projects,omitempty"`
	Notifications             []primitive.ObjectID `json:"notification,omitempty" bson:"notification,omitempty"`
}

type User struct {
	ID                        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Email                     string               `json:"email" validate:"required,email"`
	HashedPassword            []byte               `json:"-" validate:"required"`
	Avatar                    string               `json:"avatar"`
	FullName                  string               `json:"full_name" bson:"full_name"`
	Description               string               `json:"description"`
	Contacts                  []string             `json:"contacts"`
	Location                  primitive.ObjectID   `json:"location"`
	Role                      Role                 `json:"role"`
	Job                       string               `json:"job"`
	Tags                      []primitive.ObjectID `json:"tags"`
	IncomingContactRequests   []primitive.ObjectID `json:"incoming_contact_requests,omitempty" bson:"incoming_contact_requests,omitempty"`
	OutgoingContactRequests   []primitive.ObjectID `json:"outgoing_contact_requests,omitempty" bson:"outgoing_contact_requests,omitempty"`
	ConfirmedContactsRequests []primitive.ObjectID `json:"confirmed_contacts_requests,omitempty" bson:"confirmed_contacts_requests,omitempty"`
	BlockedUsers              []primitive.ObjectID `json:"blocked_users,omitempty" bson:"blocked_users,omitempty"`
	ProjectsApplications      []primitive.ObjectID `json:"projects_applications,omitempty" bson:"projects_applications,omitempty"`
	ConfirmedApplications     []primitive.ObjectID `json:"confirmed_applications,omitempty" bson:"confirmed_applications,omitempty"`
	ProjectsHistory           []primitive.ObjectID `json:"projects_history,omitempty" bson:"projects_history,omitempty"`
	CreatedProjects           []primitive.ObjectID `json:"created_projects,omitempty" bson:"created_projects,omitempty"`
	Notifications             []primitive.ObjectID `json:"notification,omitempty" bson:"notification,omitempty"`
}

type NotificationStatus int

const (
	New NotificationStatus = iota
	Readed
)

type NotificationType int

const (
	ContactsRequested NotificationType = iota
	ContactsRequestApproved
	ProjectApproved
	ProjetcDeclined
	NewComment
)

type Notification struct {
	id        int64              `json:"id"`
	CreatedAt string             `json:"created_at"`
	Status    NotificationStatus `json:"status"`
	Type      NotificationType   `json:"type"`
	Message   string             `json:"message"`
	User      int64              `json:"userId"`
}

type Event struct {
	ID               primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Slug             string               `json:"slug"`
	Cover            string               `json:"cover"`
	Title            Translations         `json:"title"`
	Description      Translations         `json:"description"`
	Location         primitive.ObjectID   `json:"location" validate:"required"`
	Date             time.Time            `json:"date" validate:"required"`
	TermsOfVisit     Translations         `json:"terms_of_visit" bson:"terms_of_visit"`
	Tags             []primitive.ObjectID `json:"tags" validate:"required"`
	Author           primitive.ObjectID   `json:"author" validate:"required" bson:"author"`
	ModerationStatus ModerationStatus     `json:"moderation_status" bson:"moderation_status"`
	ReasonOfReject   string               `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
}

type ModerationStatus int

const (
	ForRevision ModerationStatus = iota
	Accepted
	Rejected
)

type Project struct {
	ID                   primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Slug                 string               `json:"slug"`
	Covers               []string             `json:"covers,omitempty" bson:"covers,omitempty"`
	Author               primitive.ObjectID   `json:"author" validate:"required" bson:"author"`
	Title                Translations         `json:"title"`
	Description          Translations         `json:"description"`
	Objective            Translations         `json:"objective"`
	WhoIsNeeded          Translations         `json:"who_is_needed"`
	Tags                 []primitive.ObjectID `json:"tags" validate:"required" bson:"tags"`
	Applicants           []primitive.ObjectID `json:"applicants,omitempty" bson:"applicants,omitempty"`
	Views                int64                `json:"views" bson:"views"`
	ModerationStatus     ModerationStatus     `json:"moderation_status" bson:"moderation_status"`
	ReasonOfReject       string               `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
	SuccessfulApplicants []primitive.ObjectID `json:"successful_applicants,omitempty" bson:"successful_applicants,omitempty"`
	RejectedApplicants   []primitive.ObjectID `json:"rejected_applicants,omitempty" bson:"rejected_applicants,omitempty"`
}

type Research struct {
	ID               primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Slug             string               `json:"slug"`
	Title            Translations         `json:"title"`
	Description      Translations         `json:"description"`
	Tags             []primitive.ObjectID `json:"tags" validate:"required"`
	Link             string               `json:"link" validate:"required"`
	Author           primitive.ObjectID   `json:"author" validate:"required"`
	ModerationStatus ModerationStatus     `json:"moderation_status" bson:"moderation_status"`
	ReasonOfReject   string               `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
}

type Statistic struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     Translations       `json:"translations"`
	Year      int64              `json:"year" validate:"required"`
	Count     int64              `json:"count" validate:"required"`
	YearDelta *int64             `json:"year_delta" validate:"required"`
}

type Location struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Value      string             `json:"value" validate:"required"`
	Country    string             `json:"country" validate:"required"`
	Region     string             `json:"region"`
	City       string             `json:"city"`
	Settlement string             `json:"settlement"`
	Street     string             `json:"street"`
	House      string             `json:"house"`
	ExtraInfo  string             `json:"extra_info"`
}

type Suggestions struct {
	Suggestions []map[string]interface{}
}

type Tag struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title Translations       `json:"title"`
}

type Translations struct {
	En string `bson:"en" json:"en" validate:"required_without_all=Ru Hy"`
	Ru string `bson:"ru" json:"ru" validate:"required_without_all=En Hy"`
	Hy string `bson:"hy" json:"hy" validate:"required_without_all=En Ru"`
}
