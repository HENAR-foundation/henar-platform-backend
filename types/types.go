package types

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

// validate:"required_without_all=Ru Hy"
// validate:"required_without_all=En Hy"
// validate:"required_without_all=En Ru"

type Role string

const (
	Admin      Role = "admin"
	Specialist Role = "specialist"
)

type UserCredentials struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required" bson:"password"`
}

// Password string `json:"password" validate:"required"`
type Contacts struct {
	Phone     string `json:"phone,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Linkedin  string `json:"linkedin,omitempty"`
}

type ContactsRequest struct {
	IncomingContactRequests   map[primitive.ObjectID]bool `json:"incoming_contact_requests" bson:"incoming_contact_requests"`
	OutgoingContactRequests   map[primitive.ObjectID]bool `json:"outgoing_contact_requests" bson:"outgoing_contact_requests"`
	ConfirmedContactsRequests map[primitive.ObjectID]bool `json:"confirmed_contacts_requests" bson:"confirmed_contacts_requests"`
	BlockedUsers              map[primitive.ObjectID]bool `json:"blocked_users" bson:"blocked_users"`
}

// TODO: do I need projects history?
type UserProjects struct {
	ProjectsApplications  map[primitive.ObjectID]bool `json:"projects_applications" bson:"projects_applications"`
	ConfirmedApplications map[primitive.ObjectID]bool `json:"confirmed_applications" bson:"confirmed_applications"`
	RejectedApplicants    map[primitive.ObjectID]bool `json:"rejected_applicants"omitempty" bson:"rejected_applicants"`
	CreatedProjects       map[primitive.ObjectID]bool `json:"created_projects" bson:"created_projects"`
}

type UserBody struct {
	Avatar          string               `json:"avatar"`
	FullName        Translations         `json:"full_name,omitempty" bson:"full_name,omitempty"`
	Description     string               `json:"description"`
	Contacts        Contacts             `json:"contacts,omitempty"`
	Location        primitive.ObjectID   `json:"location,omitempty" bson:"location,omitempty"`
	Role            *Role                `json:"role,omitempty"`
	Job             string               `json:"job"`
	Language        string               `json:"language"`
	Tags            []primitive.ObjectID `json:"tags"`
	Notifications   []primitive.ObjectID `json:"notification,omitempty" bson:"notification,omitempty"`
	Events          []primitive.ObjectID `json:"events,omitempty" bson:"events,omitempty"`
	Researches      []primitive.ObjectID `json:"researches,omitempty" bson:"researches,omitempty"`
	ContactsRequest `json:"contacts_request,omitempty" bson:"contacts_request"`
	UserProjects    `json:"user_projects,omitempty" bson:"user_projects"`
}

type User struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserCredentials `bson:"user_credentials"`
	UserBody        `bson:"user_body"`
}

// TODO: add Events, ... [] to user for admin
// TODO: concat User and UserBody
// TODO: add Role by default on create user
// TODO: how to add required for other fields after registration?
// TODO: return only specialist if user is not admin
// TODO: add default language

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
	ID               primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedBy        primitive.ObjectID   `json:"created_by,omitempty" bson:"created_by,omitempty"`
	Slug             string               `json:"slug"`
	Cover            string               `json:"cover"`
	Title            Translations         `json:"title"`
	Description      Translations         `json:"description"`
	Location         primitive.ObjectID   `json:"location,omitempty" bson:"location,omitempty"`
	Date             time.Time            `json:"date" validate:"required"`
	TermsOfVisit     Translations         `json:"terms_of_visit" bson:"terms_of_visit"`
	Tags             []primitive.ObjectID `json:"tags" validate:"required"`
	ModerationStatus *ModerationStatus    `json:"moderation_status,omitempty" bson:"moderation_status,omitempty"`
	ReasonOfReject   *string              `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
}

type ModerationStatus string

const (
	Pending  ModerationStatus = "pending"
	Approved ModerationStatus = "approved"
	Rejected ModerationStatus = "rejected"
)

func (s *ModerationStatus) UnmarshalText(text []byte) error {
	switch string(text) {
	case "pending":
		*s = Pending
	case "approved":
		*s = Approved
	case "rejected":
		*s = Rejected
	default:
		return fmt.Errorf("unknown moderation status: %q", text)
	}
	return nil
}

func (s ModerationStatus) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

func (s ModerationStatus) IsValid() bool {
	switch s {
	case Pending, Approved, Rejected:
		return true
	}

	return false
}

type HowToHelpTheProject string

const (
	Financing HowToHelpTheProject = "financing"
	Expertise HowToHelpTheProject = "expertise"
	Resources HowToHelpTheProject = "resources"
)

func (s *HowToHelpTheProject) UnmarshalText(text []byte) error {
	switch string(text) {
	case "financing":
		*s = Financing
	case "expertise":
		*s = Expertise
	case "resources":
		*s = Resources
	default:
		return fmt.Errorf("unknown how to help the project value: %q", text)
	}
	return nil
}

func (s HowToHelpTheProject) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

func (s HowToHelpTheProject) IsValid() bool {
	switch s {
	case Financing, Expertise, Resources:
		return true
	}

	return false
}

type ProjectStatus string

const (
	Ideation             ProjectStatus = "ideation"
	Implementation       ProjectStatus = "implementation"
	LaunchAndExecution   ProjectStatus = "launchAndExecution"
	PerfomanceAndControl ProjectStatus = "perfomanceAndControl"
	Closed               ProjectStatus = "closed"
)

func (s *ProjectStatus) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ideation":
		*s = Ideation
	case "implementation":
		*s = Implementation
	case "launchAndExecution":
		*s = LaunchAndExecution
	case "perfomanceAndControl":
		*s = PerfomanceAndControl
	case "closed":
		*s = Closed
	default:
		return fmt.Errorf("unknown project status: %q", text)
	}
	return nil
}

func (s ProjectStatus) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

func (s ProjectStatus) ValidateEnum() bool {
	switch s {
	case Ideation, Implementation, LaunchAndExecution, PerfomanceAndControl, Closed:
		return true
	}

	return false
}

type Enum interface {
	IsValid() bool
}

func ValidateEnum(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(Enum)
	if !ok {
		return false
	}
	return value.IsValid()
}

// TODO: how_to_help_the_project can has many values? can be empty?
// TODO: author can update project moderation_status after reject
type Project struct {
	ID                   primitive.ObjectID           `json:"_id,omitempty" bson:"_id,omitempty"`
	Slug                 *string                      `json:"slug, omitempty" bson:"slug,omitempty"`
	Covers               []string                     `json:"covers,omitempty" bson:"covers,omitempty"`
	CreatedBy            primitive.ObjectID           `json:"created_by,omitempty" bson:"created_by,omitempty"`
	Title                Translations                 `json:"title"`
	Description          Translations                 `json:"description"`
	Objective            Translations                 `json:"objective"`
	WhoIsNeeded          Translations                 `json:"who_is_needed" bson:"who_is_needed"`
	Tags                 []primitive.ObjectID         `json:"tags" validate:"required" bson:"tags"`
	Views                *int64                       `json:"views" bson:"views,omitempty"`
	HowToHelpTheProject  map[HowToHelpTheProject]bool `json:"how_to_help_the_project" bson:"how_to_help_the_project,omitempty"`
	ProjectStatus        ProjectStatus                `json:"project_status" bson:"project_status,omitempty"`
	ModerationStatus     *ModerationStatus            `json:"moderation_status,omitempty" bson:"moderation_status,omitempty"`
	ReasonOfReject       *string                      `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
	Applicants           *[]primitive.ObjectID        `json:"applicants,omitempty" bson:"applicants,omitempty"`
	SuccessfulApplicants *[]primitive.ObjectID        `json:"successful_applicants,omitempty" bson:"successful_applicants,omitempty"`
	RejectedApplicants   *[]primitive.ObjectID        `json:"rejected_applicants,omitempty" bson:"rejected_applicants,omitempty"`
}

// TODO: handler for reject applicant

type Research struct {
	ID               primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedBy        primitive.ObjectID   `json:"created_by,omitempty" bson:"created_by,omitempty"`
	Slug             string               `json:"slug"`
	Title            Translations         `json:"title"`
	Description      Translations         `json:"description"`
	Tags             []primitive.ObjectID `json:"tags" validate:"required"`
	Link             string               `json:"link" validate:"required"`
	ModerationStatus *ModerationStatus    `json:"moderation_status,omitempty" bson:"moderation_status,omitempty"`
	ReasonOfReject   *string              `json:"reason_of_reject,omitempty" bson:"reason_of_reject,omitempty"`
}

type Statistic struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     Translations       `json:"translations"`
	Year      int64              `json:"year" validate:"required"`
	Count     int64              `json:"count" validate:"required"`
	YearDelta *int64             `json:"year_delta" validate:"required"`
}

// TODO: create location on user and event create
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

// validate:"required_without_all=Ru Hy
// validate:"required_without_all=En Hy
// validate:"required_without_all=En Ru

type Translations struct {
	En string `bson:"en" json:"en"`
	Ru string `bson:"ru" json:"ru"`
	Hy string `bson:"hy" json:"hy"`
}

type FileResponce struct {
	URL string `bson:"en" json:"url"`
}
