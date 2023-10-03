package service

import (
	"strings"
	"testing"
)

func TestContactPhonesNormalize(t *testing.T) {
	empty := ContactPhones{}
	n := empty.Normalize()
	if len(n) != 0 {
		t.Errorf("expecting length 0 got %d", len(n))
	}
	contacts := ContactPhones{
		ContactPhone{
			Phone:      "",
			LinkedApps: []string{WhatsApp},
		},
		ContactPhone{
			Phone:      "+5571900000000",
			LinkedApps: []string{WhatsApp},
		},
		ContactPhone{
			Phone:      "71900000000",
			LinkedApps: []string{WhatsApp, Telegram},
		},
		ContactPhone{
			Phone:      "+55 (71) 90000-0000",
			LinkedApps: []string{Telegram, WhatsApp},
		},
		ContactPhone{
			Phone:      "5571900000000",
			LinkedApps: []string{Telegram},
		},
	}

	newContacts := contacts.Normalize()
	validContacts := contacts[1:]
	normalizedPhones := []string{}
	for i, c := range newContacts {
		normalizedPhones = append(normalizedPhones, c.Phone)
		newApps := strings.Join(c.LinkedApps, ", ")
		defaultApps := strings.Join(validContacts[i].LinkedApps, ", ")
		if newApps != defaultApps {
			t.Errorf("[%d]: expecting %s got %s", i, defaultApps, newApps)
		}
	}
	expectedPhones := "+5571900000000, +5571900000000, +5571900000000, +5571900000000"
	if strings.Join(normalizedPhones, ", ") != expectedPhones {
		t.Errorf("expecting %s got %s", expectedPhones, normalizedPhones)
	}
}

func TestContactPhonesFirst(t *testing.T) {
	cellPhone1 := "+5571988706269"
	cellPhone2 := "7196210085"
	cellPhone3 := "71988706269"
	landline1 := "7130277065"

	empty := ContactPhones{}
	p1E := empty.First()
	if len(p1E) != 0 {
		t.Errorf("expecting length 0 got %d", len(p1E))
	}
	contactsT1 := ContactPhones{
		ContactPhone{
			Phone:      landline1,
			LinkedApps: []string{},
		},
		ContactPhone{
			Phone:      cellPhone1,
			LinkedApps: []string{WhatsApp},
		},
		ContactPhone{
			Phone:      cellPhone2,
			LinkedApps: []string{Telegram},
		},
		ContactPhone{
			Phone:      cellPhone3,
			LinkedApps: []string{WhatsApp, Telegram},
		},
	}
	p1T1 := contactsT1.First(WhatsApp)
	if p1T1 != contactsT1[1].Phone {
		t.Errorf("expecting %s, got %s", contactsT1[1].Phone, p1T1)
	}
	p2T1 := contactsT1.First(Telegram)
	if p2T1 != contactsT1[2].Phone {
		t.Errorf("expecting %s, got %s", contactsT1[2].Phone, p2T1)
	}
	p3T1 := contactsT1.First()
	if p3T1 != contactsT1[0].Phone {
		t.Errorf("expecting %s, got %s", contactsT1[0].Phone, p3T1)
	}

	contactsT2 := ContactPhones{
		ContactPhone{
			Phone:      cellPhone1,
			LinkedApps: []string{},
		},
		ContactPhone{
			Phone:      cellPhone2,
			LinkedApps: []string{},
		},
	}
	p1T2 := contactsT2.First(WhatsApp)
	if p1T2 != contactsT2[0].Phone {
		t.Errorf("expecting %s, got %s", contactsT2[0].Phone, p1T2)
	}

	contactsT3 := ContactPhones{
		ContactPhone{
			Phone:      "",
			LinkedApps: []string{WhatsApp, Telegram},
		},
		ContactPhone{
			Phone:      cellPhone2,
			LinkedApps: []string{Telegram},
		},
		ContactPhone{
			Phone:      cellPhone1,
			LinkedApps: []string{WhatsApp},
		},
	}
	p1T3 := contactsT3.First(WhatsApp)
	if p1T3 != contactsT3[2].Phone {
		t.Errorf("expecting %s, got %s", contactsT3[2].Phone, p1T3)
	}
	p2T3 := contactsT3.First(Telegram)
	if p2T3 != contactsT3[1].Phone {
		t.Errorf("expecting %s, got %s", contactsT3[1].Phone, p2T3)
	}
	p3T3 := contactsT3.First()
	if p3T3 != contactsT3[1].Phone {
		t.Errorf("expecting %s, got %s", contactsT3[1].Phone, p3T3)
	}

	contactsT4 := ContactPhones{
		ContactPhone{
			Phone:      landline1,
			LinkedApps: []string{WhatsApp, Telegram},
		},
		ContactPhone{
			Phone:      cellPhone1,
			LinkedApps: []string{},
		},
	}
	p1T4 := contactsT4.First(WhatsApp)
	if p1T4 != contactsT4[1].Phone {
		t.Errorf("expecting %s, got %s", contactsT4[1].Phone, p1T4)
	}
	p2T4 := contactsT4.First()
	if p2T4 != contactsT4[0].Phone {
		t.Errorf("expecting %s, got %s", contactsT4[0].Phone, p2T4)
	}
}
