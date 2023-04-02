package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role int

const (
	Admin Role = iota
	Specialist
)

type Contacts struct {
	Emain    string `json:"emain"`
	Phone    string `json:"phone"`
	Facebook string `json:"facebook"`
	Insta    string `json:"insta"`
	Linkedin string `json:"linkedin"`
}

type User struct {
	Id                        int64    `json:"id"`
	CreatedAt                 string   `json:"created_at"`
	Avatar                    string   `json:"avatar"`
	FullName                  string   `json:"full_name"`
	Description               string   `json:"description"`
	Contacts                  []string `json:"contacts"`
	Location                  int64    `json:"location"`
	Email                     string   `json:"email"`
	Password                  string   `json:"password"`
	Role                      Role     `json:"role"`
	Job                       string   `json:"job"`
	Tags                      []int64  `json:"tags"`
	IncomingContactRequests   []int64  `json:"incoming_contact_requests"`
	OutgoingContactRequests   []int64  `json:"outgoing_contact_requests"`
	ConfirmedContactsRequests []int64  `json:"confirmed_contacts_requests"`
	BlockedUsers              []int64  `json:"blocked_users"`
	ProjectsApplications      []int64  `json:"projects_applications"`
	ConfirmedApplications     []int64  `json:"confirmed_applications"`
	ProjectsHistory           []int64  `json:"projects_history"`
	CreatedProjects           []int64  `json:"created_projects"`
	Notifications             []int64  `json:"notification"`
}

type NotificationStatus int

const (
	New NotificationStatus = iota
	Readed
)

type NotificationType int

const (
	ProjectUpdate NotificationType = iota
	ContactUpdate
)

// TODO: update notification struct
type Notification struct {
	id        int64              `json:"id"`
	CreatedAt string             `json:"created_at"`
	Status    NotificationStatus `json:"status"`
	Type      NotificationType   `json:"type"`
	Message   string             `json:"message"`
}

type Event struct {
	Id           int64     `json:"id"`
	CreatedAt    string    `json:"created_at"`
	Cover        string    `json:"cover"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Location     int64     `json:"location"`
	Date         time.Time `json:"date"`
	TermsOfVisit string    `json:"terms_of_visit"`
	Tags         []int64   `json:"tags"`
}

type ModerationStatus int

const (
	ForRevision ModerationStatus = iota
	Accepted
	Rejected
)

type Project struct {
	ID                   primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Slug                 string               `json:"slug" bson:"slug"`
	Covers               []string             `json:"covers,omitempty" bson:"covers,omitempty"`
	Author               primitive.ObjectID   `json:"author" validate:"required" bson:"author"`
	Title                string               `json:"title" validate:"required" bson:"title"`
	Description          string               `json:"description" validate:"required" bson:"description"`
	Objective            string               `json:"objective" validate:"required" bson:"objective"`
	WhoIsNeeded          string               `json:"who_is_needed" validate:"required" bson:"who_is_needed"`
	Tags                 []primitive.ObjectID `json:"tags" validate:"required" bson:"tags"`
	Applicants           []primitive.ObjectID `json:"applicants,omitempty" bson:"applicants,omitempty"`
	Views                int64                `json:"views" bson:"views"`
	ModerationStatus     ModerationStatus     `json:"moderation_status" bson:"moderation_status"`
	ReasonOfReject       string               `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
	SuccessfulApplicants []primitive.ObjectID `json:"successful_applicants,omitempty" bson:"successful_applicants,omitempty"`
	RejectedApplicants   []primitive.ObjectID `json:"rejected_applicants,omitempty" bson:"rejected_applicants,omitempty"`
}

type Research struct {
	ID          primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Tags        []primitive.ObjectID `json:"tags" validate:"required" bson:"tags"`
	Link        string               `json:"link"`
}

type Statistic struct {
	Id        int64  `json:"id"`
	Title     string `json:"title"`
	Year      int64  `json:"year"`
	Count     int64  `json:"count"`
	YearDelta int64  `json:"year_delta"`
}

type Location struct {
	Id        int64  `json:"id"`
	State     string `json:"state"`
	City      string `json:"city"`
	Street    string `json:"street"`
	ExtraInfo string `json:"extra_info"`
}

type Tag struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Translations Translations       `json:"translations,omitempty" bson:"translations,omitempty"`
}

type Translations struct {
	En string `bson:"en" json:"en"`
	Ru string `bson:"ru" json:"ru"`
	Hy string `bson:"hy" json:"hy"`
}
