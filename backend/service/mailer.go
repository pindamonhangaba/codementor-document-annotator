package service

import (
	"github.com/gofrs/uuid"
)

type AccessModel string

const (
	TokenAccessModel = AccessModel("token")
	LinkAccessModel  = AccessModel("link")
)

// TwoFactorRequest email Model
type TwoFactorRequest struct {
	Name            string
	PasswordNumber  string
	ConfirmationURL string
	Email           string
	ImagesTemplate  string
}

// ConfigEmailTemplates holds email templates for transactional emails
type ConfigEmailTemplates struct {
	ResetPasswordHtml                   string `json:"resetPasswordHtml"`
	EmailAuthenticateHTML               string `json:"emailAuthenticateHTML"`
	EmailAuthenticateLinkHTML           string `json:"emailAuthenticateLinkHTML"`
	ResetPasswordLinkHtml               string `json:"resetPasswordLinkHtml"`
	ConfirmRegistrationHtml             string `json:"confirmRegistrationHtml"`
	InvitationHtml                      string `json:"invitationHtml"`
	NewMessageHtml                      string `json:"newMessageHtml"`
	ScheduleDoctorHtml                  string `json:"scheduleDoctorHtml"`
	MessagesUnreadHtml                  string `json:"messagesUnreadHtml"`
	SendPrescriptionsHTML               string `json:"sendPrescriptionsHTML"`
	DocumentReportHtml                  string `json:"documentReportHtml"`
	ControlledAppointmentInvitationHTML string `json:"controlledAppointmentInvitationHTML"`
	AlertPaymentBlockHTML               string `json:"alertPaymentBlockHTML"`
	AlertNewDeviceLoginHTML             string `json:"alertNewDeviceLoginHTML"`
	AppointmentSchedulerHtml            string `json:"appointmentSchedulerHtml"`
	VoucherSharingHtml                  string `json:"voucherSharingHtml"`
}

// AlertNewDeviceLogin Model
type AlertNewDeviceLogin struct {
	UserID          uuid.UUID `db:"user_id" json:"userID"`
	RecipientName   string    `db:"recipient_name" json:"recipientName"`
	RecipientEmail  string    `db:"recipient_email" json:"recipientEmail"`
	Location        string    `db:"location" json:"location"`
	Date            string    `db:"date" json:"date"`
	AlertEmail      bool      `db:"alert_email" json:"alertEmail"`
	ImagesTemplate  string    `db:"images_template" json:"imagesTemplate"`
	ImagesAgentType string    `db:"images_agent_type" json:"imagesAgentType"`
	InvitationLink  string    `db:"invitation_link" json:"invitationLink"`
}

type notificationChannel string

const (
	NotificationChannelPhone notificationChannel = "phone"
	NotificationChannelEmail notificationChannel = "email"
)

type NotificationChannels []notificationChannel

func (nc NotificationChannels) Validate() error {
	validationErrors := NewModelValidationError("NotificationChannels")

	for _, c := range nc {
		if c != NotificationChannelEmail && c != NotificationChannelPhone {
			validationErrors.Append("value", "invalid channel: "+string(c))
		}
	}

	return validationErrors.Check()
}

func (nc NotificationChannels) Has(ch notificationChannel) bool {

	for _, c := range nc {
		if c == ch {
			return true
		}
	}

	return false
}
