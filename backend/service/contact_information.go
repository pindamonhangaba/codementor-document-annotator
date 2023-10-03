package service

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

const (
	Telegram = "telegram"
	WhatsApp = "whatsapp"

	BrazilPhoneCode = "+55"
)

type ContactPhone struct {
	Phone      string   `db:"phone" json:"phone"`
	LinkedApps []string `db:"linked_apps" json:"linkedApps"`
}
type ContactPhones []ContactPhone

func (contacts ContactPhones) First(linkedApps ...string) string {
	if len(contacts) == 0 {
		return ""
	}
	if len(linkedApps) > 0 {
		cellPhones := ContactPhones{}
		for _, contact := range contacts {
			phone := nonPhoneNumberRegex.ReplaceAllString(contact.Phone, "")
			isValidPhone := len(phone) > 0
			// if the number is from Brazil, validate
			if isValidPhone && (string(phone[0]) != "+" || strings.Contains(phone, BrazilPhoneCode)) {
				isValidPhone = isCellPhoneBR(phone)
			}
			if isValidPhone {
				cellPhones = append(cellPhones, contact)
			}
		}
		if len(cellPhones) == 0 {
			cellPhones = contacts
		}
		for _, contact := range cellPhones {
			la := strings.Join(contact.LinkedApps, ", ")
			for _, app := range linkedApps {
				if strings.Contains(la, app) {
					return nonPhoneNumberRegex.ReplaceAllString(contact.Phone, "")
				}
			}
		}
		return nonPhoneNumberRegex.ReplaceAllString(cellPhones[0].Phone, "")
	}
	validPhones := ContactPhones{}
	for _, contact := range contacts {
		phone := nonPhoneNumberRegex.ReplaceAllString(contact.Phone, "")
		if len(phone) > 0 {
			validPhones = append(validPhones, contact)
		}
	}
	if len(validPhones) == 0 {
		validPhones = contacts
	}
	return nonPhoneNumberRegex.ReplaceAllString(validPhones[0].Phone, "")
}

func (contacts ContactPhones) Normalize() ContactPhones {
	validPhones := ContactPhones{}
	for _, contact := range contacts {
		phone := NormalizePhone(contact.Phone)
		if len(phone) == 0 {
			continue
		}
		contact.Phone = phone
		validPhones = append(validPhones, contact)
	}
	return validPhones
}

type contactPhoneAppsValidator struct {
	list []string
}

func (sch contactPhoneAppsValidator) List() []string {
	return sch.list
}

func (sch contactPhoneAppsValidator) Validate(s string) bool {
	for _, item := range sch.list {
		if item == s {
			return true
		}
	}
	return false
}

var ContactPhoneAppsValidator = contactPhoneAppsValidator{
	list: []string{
		WhatsApp,
		Telegram,
	},
}

// Value implements the driver Valuer interface.
func (i ContactPhones) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *ContactPhones) Scan(src interface{}) error {
	var source []byte
	// let's support string and []byte
	switch d := src.(type) {
	case string:
		source = []byte(d)
	case []byte:
		source = d
	default:
		return errors.New("incompatible type for PatientContactPhone")
	}
	return json.Unmarshal(source, i)
}
