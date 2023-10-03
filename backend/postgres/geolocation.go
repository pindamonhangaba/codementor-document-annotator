package postgres

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/pindamonhangaba/monoboi/backend/lib/cte"
	"github.com/pindamonhangaba/monoboi/backend/lib/sql_"
	sq "github.com/pindamonhangaba/monoboi/backend/lib/sqrl"
	"github.com/pindamonhangaba/monoboi/backend/service"
)

// SearchGeoLocationsService service to return GeoLocationView
type SearchGeoLocationsService struct {
	DB *sqlx.DB
}

// Run return a search of GeoLocationView
func (g *SearchGeoLocationsService) Run(f service.FilterGeoLocation) ([]service.GeoLocationView, error) {
	res, err := searchGeoLocation(g.DB, f)
	return res, err
}

// searchGeoLocation returns a search of GeoLocationView
func searchGeoLocation(db service.DB, f service.FilterGeoLocation) ([]service.GeoLocationView, error) {
	epv := []service.GeoLocationView{}

	query := psql.Select("address_formatted", "address").
		From("all_results").
		Prefix(`WITH
		geo_location_city AS (
			SELECT
				(cty."name"||', '||gst."name"||' - '||gco."name" ) address_formatted,
				jsonb_build_object('city', cty."name", 'state', gst."name", 'stateAcronym', gst.acronym, 'country', gco."name", 'coordinate', cty.coordinate) AS address
			FROM geo_city cty
			JOIN geo_state gst USING(gest_id)
			JOIN geo_country gco USING(geco_id)
			WHERE TRIM(?) != '' AND
			lower(unaccent(cty."name")) LIKE (lower(unaccent(?))||'%')
		),
		geo_location_address AS (
			SELECT
				(
					CASE WHEN gzc.street IS NULL OR TRIM(gzc.street) = '' THEN '' ELSE gzc.street||', ' END ||
					CASE WHEN gzc.district IS NULL OR TRIM(gzc.district) = '' THEN '' ELSE gzc.district||', ' END ||
					gzc.zip_code||', '||cty."name"||', '||gst."name"||' - '||gco."name"
				) address_formatted,
				jsonb_build_object(
					'cep', gzc.zip_code,
					'street', gzc.street,
					'district', gzc.district,
					'complement', gzc.complement,
					'city', cty."name",
					'state', gst."name",
					'stateAcronym', gst.acronym,
					'country', gco."name",
					'coordinate', cty.coordinate
				) AS address
			FROM geo_zip_code gzc
			JOIN geo_city cty USING(geci_id)
			JOIN geo_state gst USING(gest_id)
			JOIN geo_country gco USING(geco_id)
			WHERE gzc.zip_code = ?
		),
		all_capitals AS (
			SELECT
				(cty."name"||', '||gst."name"||' - '||gco."name" ) address_formatted,
				jsonb_build_object('city', cty."name", 'state', gst."name", 'stateAcronym', gst.acronym, 'country', gco."name", 'coordinate', cty.coordinate) AS address
			FROM geo_state_capital gsc
			JOIN geo_city cty USING(geci_id)
			JOIN geo_state gst ON gsc.gest_id = gst.gest_id
			JOIN geo_country gco USING(geco_id)
			LEFT JOIN geo_location_city glc ON 1 = 1
			LEFT JOIN geo_location_address gla ON 1 = 1
			WHERE glc.address_formatted IS NULL AND gla.address_formatted IS NULL
		),
		all_results AS (
			SELECT address_formatted, address FROM geo_location_city
			UNION
			SELECT address_formatted, address FROM geo_location_address
			UNION
			SELECT address_formatted, address FROM all_capitals
		)`, f.Search, f.Search, f.CEP)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "generating search GeoLocationView sql")
	}

	err = db.Select(&epv, qSQL, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "search GeoLocationView")
	}
	return epv, nil
}

// UpdateZipCodeService service to return GeoZipCode
type UpdateZipCodeService struct {
	DB *sqlx.DB
}

// Run return a search of GeoLocationView
func (g *UpdateZipCodeService) Run(zip *service.CreateZipCode) (*service.GeoZipCode, error) {
	res, err := updateZipCodeDatabase(g.DB, zip)
	return res, err
}

// Create a new zip_code to database
func updateZipCodeDatabase(db service.DB, zip *service.CreateZipCode) (*service.GeoZipCode, error) {
	argsCTE, err := cte.CTEFromString("arguments", `
	SELECT
		$1::VARCHAR AS zip_code,
		$2::VARCHAR AS coordinate,
		$3::VARCHAR AS street,
		$4::VARCHAR AS complement,
		$5::VARCHAR AS district
	`, []interface{}{zip.ZipCode, zip.Coordinate, zip.Street, zip.Complement, zip.District})

	if err != nil {
		return nil, errors.Wrap(err, "arguments cte")
	}

	ctes, args, err := cte.With(argsCTE).AsPrefix()
	if err != nil {
		return nil, err
	}

	query := psql.Insert("geo_zip_code").
		Columns("geci_id", "zip_code", "coordinate", "street", "complement", "district").
		Select(
			sq.Select("gct.geci_id", "arg.zip_code", "(arg.coordinate)::point", "arg.street", "arg.complement", "arg.district").
				From(sql_.Alias("geo_city", "gct"), sql_.Alias(argsCTE.As(), "arg")).
				Where(sq.Eq{"gct.identifier": zip.Identifier}),
		).
		Prefix(ctes, args...).
		Suffix(`
		ON CONFLICT (zip_code) DO UPDATE
		SET coordinate = excluded.coordinate,
			street = excluded.street,
			complement = excluded.complement,
			district = excluded.district
		RETURNING *`)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "update zip_code sql")
	}

	res := service.GeoZipCode{}
	err = db.Get(&res, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "update zip_code")
	}

	return &res, nil
}
