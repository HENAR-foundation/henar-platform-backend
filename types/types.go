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
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password,omitempty" bson:"password,omitempty"`
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPassword struct {
	Password        *string `json:"password" validate:"required"`
	PasswordConfirm *string `json:"password_confirm" validate:"required"`
	Token           string  `json:"token,omitempty" validate:"required,hexadecimal"`
}

type PasswordUpdate struct {
	Password    *string `json:"password" validate:"required"`
	NewPassword *string `json:"new_password" validate:"required"`
}

type Contacts struct {
	Phone     string `json:"phone"`
	Facebook  string `json:"facebook"`
	Instagram string `json:"instagram"`
	Linkedin  string `json:"linkedin"`
}

type RequestMessage struct {
	Message string
}

type StatisticsCategory struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Steps []string           `json:"steps" bson:"steps"`
}

type ContactsRequest struct {
	IncomingContactRequests   map[primitive.ObjectID]string `json:"incoming_contact_requests" bson:"incoming_contact_requests"`
	OutgoingContactRequests   map[primitive.ObjectID]string `json:"outgoing_contact_requests" bson:"outgoing_contact_requests"`
	ConfirmedContactsRequests map[primitive.ObjectID]string `json:"confirmed_contacts_requests" bson:"confirmed_contacts_requests"`
	BlockedUsers              map[primitive.ObjectID]string `json:"blocked_users" bson:"blocked_users"`
	ApprovedContacts          map[primitive.ObjectID]string `json:"approved_contacts" bson:"approved_contacts"`
}

type UserProjects struct {
	ProjectsApplications  map[primitive.ObjectID]primitive.ObjectID `json:"projects_applications" bson:"projects_applications"`
	ConfirmedApplications map[primitive.ObjectID]primitive.ObjectID `json:"confirmed_applications" bson:"confirmed_applications"`
	RejectedApplicants    map[primitive.ObjectID]primitive.ObjectID `json:"rejected_applicants" bson:"rejected_applicants"`
	CreatedProjects       map[primitive.ObjectID]bool               `json:"created_projects" bson:"created_projects"`
}

type UserBody struct {
	Banned          bool                        `json:"banned" bson:"banned,omitempty"`
	Avatar          string                      `json:"avatar"`
	FirstName       string                      `json:"first_name" bson:"first_name,omitempty"`
	LastName        string                      `json:"last_name" bson:"last_name,omitempty"`
	Description     string                      `json:"description"`
	Contacts        Contacts                    `json:"contacts"`
	Location        *primitive.ObjectID         `json:"location" bson:"location,omitempty"`
	Role            *Role                       `json:"role" bson:"role,omitempty"`
	Job             string                      `json:"job"`
	Language        string                      `json:"language,omitempty"`
	Tags            []primitive.ObjectID        `json:"tags"`
	Notifications   []primitive.ObjectID        `json:"notifications,omitempty" bson:"notifications,omitempty"`
	Events          map[primitive.ObjectID]bool `json:"events" bson:"events"`
	Researches      map[primitive.ObjectID]bool `json:"researches" bson:"researches"`
	ContactsRequest `json:"contacts_request" bson:"contacts_request"`
	UserProjects    `json:"user_projects" bson:"user_projects"`
}

type User struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	IsActivated     bool               `json:"-" bson:"is_activated"`
	IsEmailVerified *bool              `json:"-" bson:"is_email_verified"`
	UserCredentials `bson:"user_credentials"`
	UserBody        `bson:"user_body"`
}

type NotificationStatus string

const (
	New  NotificationStatus = "new"
	Read NotificationStatus = "read"
)

type NotificationType string

const (
	ApproveApplicant        NotificationType = "approve_applicant"
	ContactsRequested       NotificationType = "contact_requested"
	ContactsRequestApproved NotificationType = "contact_request_approved"
	ProjectRequest          NotificationType = "project_request"
	ProjectApproved         NotificationType = "project_approved"
	ProjetcDeclined         NotificationType = "project_declined"
	NewComment              NotificationType = "new_comment"
)

type NotificationAcceptiongRequestBody struct {
	NotificationsIds []string `json:"notificationsIds"`
}

type NotificationBody struct {
	PersonID       primitive.ObjectID `json:"personId" bson:"person_id"`
	PersonFullName string             `json:"personFullName" bson:"person_full_name"`
	ProjectID      string             `json:"projectId" bson:"project_id"`
	ProjectTitle   string             `json:"projectTitle" bson:"project_title"`
	Avatar         string             `json:"avatar" bson:"avatar"`
}
type Notification struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at,omitempty"`
	Status    NotificationStatus `json:"status"`
	Type      NotificationType   `json:"type"`
	User      primitive.ObjectID `json:"userId" bson:"user_id"`
	Body      NotificationBody   `json:"body" bson:"body,omitempty"`
}

type NotificationResponse struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at,omitempty"`
	Status    NotificationStatus `json:"status"`
	Type      NotificationType   `json:"type"`
	Body      NotificationBody   `json:"body" bson:"body,omitempty"`
}

type Event struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedBy   primitive.ObjectID `json:"created_by" bson:"created_by,omitempty"`
	Slug        string             `json:"slug"`
	Cover       string             `json:"cover"`
	Title       Translations       `json:"title"`
	Description Translations       `json:"description"`
	Orgs        string             `json:"orgs"`
	Location    primitive.ObjectID `json:"location" bson:"location,omitempty"`
	Date        time.Time          `json:"date" validate:"required"`
	Links       string             `json:"links" bson:"terms_of_visit"`
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

// TODO: author can update project moderation_status after reject
type Project struct {
	ID                   primitive.ObjectID          `json:"_id" bson:"_id,omitempty"`
	Slug                 *string                     `json:"slug" bson:"slug,omitempty"`
	Covers               []string                    `json:"covers" bson:"covers,omitempty"`
	CreatedBy            primitive.ObjectID          `json:"created_by" bson:"created_by,omitempty"`
	Title                Translations                `json:"title"`
	Description          Translations                `json:"description"`
	Objective            Translations                `json:"objective"`
	WhoIsNeeded          Translations                `json:"who_is_needed" bson:"who_is_needed"`
	Tags                 []primitive.ObjectID        `json:"tags" bson:"tags"`
	Views                *int64                      `json:"views" bson:"views"`
	HowToHelpTheProject  string                      `json:"how_to_help_the_project" bson:"how_to_help_the_project,omitempty"`
	ProjectStatus        ProjectStatus               `json:"project_status" bson:"project_status,omitempty"`
	ModerationStatus     *ModerationStatus           `json:"moderation_status" bson:"moderation_status,omitempty"`
	ReasonOfReject       *string                     `json:"reason_of_reject" bson:"reason_of_reject,omitempty"`
	Applicants           map[primitive.ObjectID]bool `json:"applicants" bson:"applicants,omitempty"`
	SuccessfulApplicants map[primitive.ObjectID]bool `json:"successful_applicants" bson:"successful_applicants,omitempty"`
	RejectedApplicants   map[primitive.ObjectID]bool `json:"rejected_applicants" bson:"rejected_applicants,omitempty"`
	Links                string                      `json:"links" bson:"links,omitempty"`
	Request              string                      `json:"request" bson:"request,omitempty"`
	Phase                string                      `json:"phase" bson:"phase,omitempty"`
}

type Research struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Title  string             `json:"title"`
	Link   string             `json:"link" validate:"required"`
	Source string             `json:"source" validate:"required"`
}

type StatisticTranslation struct {
	Title  string `json:"title"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

type StatisticTranslationEN struct {
	Title  string `json:"title" validate:"required"`
	Value  string `json:"value" validate:"required"`
	Source string `json:"source" validate:"required"`
}

type Statistic struct {
	ID       primitive.ObjectID     `json:"_id" bson:"_id,omitempty"`
	EN       StatisticTranslationEN `json:"en"`
	HY       StatisticTranslation   `json:"hy"`
	RU       StatisticTranslation   `json:"ru"`
	Category primitive.ObjectID     `json:"category" validate:"required"`
}

// TODO: create location on user and event create
type Location struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
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
	ID    primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
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

type VerificationDataType string

const (
	PassReset        VerificationDataType = "pass_reset"
	MailConfirmation VerificationDataType = "mail_confirmation"
)

type VerificationData struct {
	ID             primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	User           primitive.ObjectID   `json:"userId" bson:"user_id"`
	Email          string               `json:"email" bson:"email"`
	Code           string               `json:"code,omitempty" bson:"code,omitempty"`
	ResendAttempts int                  `json:"resend_attempts" bson:"resend_attempts"`
	CreatedAt      time.Time            `json:"created_at,omitempty" bson:"created_at,omitempty"`
	ExpiresAt      time.Time            `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	UsedAt         *time.Time           `json:"used_at,omitempty" bson:"used_at,omitempty"`
	Type           VerificationDataType `json:"type,omitempty" bson:"type,omitempty"`
}

type ResendVerificationEmail struct {
	Token string `json:"token,omitempty" validate:"omitempty,hexadecimal"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}
