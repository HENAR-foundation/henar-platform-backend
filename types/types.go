package types

import (
	"time"
)

type Role int

const (
	Admin Role = iota
	Specialist
)

type UserCredentials struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserCredentialsWithoutId struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	Accepted ModerationStatus = iota
	Rejected
	ForRevision
)

type Project struct {
	Id                   int64            `json:"id"`
	CreatedAt            string           `json:"created_at"`
	Slug                 string           `json:"slug"`
	Covers               []string         `json:"covers"`
	Author               int64            `json:"author"`
	Title                string           `json:"title"`
	Description          string           `json:"description"`
	Objective            string           `json:"objective"`
	WhoIsNeeded          string           `json:"who_is_needed"`
	Tags                 []int64          `json:"tags"`
	Applicants           []int64          `json:"applicants"`
	Views                int64            `json:"views"`
	ModerationStatus     ModerationStatus `json:"moderation_status"`
	ReasonOfReject       string           `json:"reason_of_reject"`
	SuccessfulApplicants []int64          `json:"successful_applicants"`
	RejectedApplicants   []int64          `json:"rejected_applicants"`
}

type Research struct {
	Id          int64   `json:"id"`
	CreatedAt   string  `json:"created_at"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Tags        []int64 `json:"tags"`
	Link        string  `json:"link"`
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

type Tags struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}
