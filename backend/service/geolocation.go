package service

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// GeoLocationAddress model
type GeoLocationAddress struct {
	Cep          *string `db:"cep" json:"cep,omitempty"`
	Street       *string `db:"street" json:"street,omitempty"`
	Complement   *string `db:"complement" json:"complement,omitempty"`
	Number       *string `db:"number" json:"number,omitempty"`
	District     *string `db:"district" json:"district,omitempty"`
	City         string  `db:"city" json:"city"`
	State        string  `db:"state" json:"state"`
	StateAcronym string  `db:"state_acronym" json:"stateAcronym"`
	Country      string  `db:"country" json:"country"`
	Coordinate   string  `db:"coordinate" json:"coordinate"`
}

// GeoZipCode  represents table geo_zip_code
type GeoZipCode struct {
	GeziID     string `db:"gezi_id" json:"geziID"`
	GeciID     string `db:"geci_id" json:"geciID"`
	ZipCode    string `db:"zip_code" json:"zipCode"`
	Coordinate string `db:"coordinate" json:"coordinate"`
	Street     string `db:"street" json:"street"`
	Complement string `db:"complement" json:"complement"`
	District   string `db:"district" json:"district"`
}

// CreateZipCode  model
type CreateZipCode struct {
	ZipCode    string `db:"zip_code" json:"zipCode"`
	Coordinate string `db:"coordinate" json:"coordinate"`
	Street     string `db:"street" json:"street"`
	Complement string `db:"complement" json:"complement"`
	District   string `db:"district" json:"district"`
	Identifier string `db:"identifier" json:"identifier"`
}

// GeoLocationView model
type GeoLocationView struct {
	AddressFormatted string             `db:"address_formatted" json:"addressFormatted"`
	Address          GeoLocationAddress `db:"address" json:"address"`
}

type Address struct {
	CEP          string `db:"cep" json:"cep"`
	Street       string `db:"street" json:"street"`
	StreetNumber string `db:"street_number" json:"streetNumber"`
	City         string `db:"city" json:"city"`
	Neighborhood string `db:"neighborhood" json:"neighborhood"`
	State        string `db:"state" json:"state"`
}

// FilterGeoLocation holds values to filter user list
type FilterGeoLocation struct {
	Search *string `json:"search"`
	CEP    *string `json:"cep"`
}

// SearchGeoLocations interface service
type SearchGeoLocations interface {
	Run(FilterGeoLocation) ([]GeoLocationView, error)
}

// UpdateZipCode interface service
type UpdateZipCode interface {
	Run(*CreateZipCode) (*GeoZipCode, error)
}

// Value implements the driver Valuer interface.
func (i GeoLocationAddress) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *GeoLocationAddress) Scan(src interface{}) error {
	var source []byte

	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for GeoLocationAddress")
	}
	return json.Unmarshal(source, i)
}

// Value implements the driver Valuer interface.
func (i Address) Value() (driver.Value, error) {
	b, err := json.Marshal(i)
	return driver.Value(b), err
}

// Scan implements the Scanner interface.
func (i *Address) Scan(src interface{}) error {
	var source []byte

	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for Address")
	}
	return json.Unmarshal(source, i)
}
