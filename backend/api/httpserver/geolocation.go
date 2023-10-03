package httpserver

import (
	"github.com/labstack/echo/v4"
	"github.com/pindamonhangaba/apiculi/endpoint"
	"github.com/pindamonhangaba/monoboi/backend/service"
	"github.com/pkg/errors"
)

// GeoLocationHandler service to create handler
type GeoLocationHandler struct {
	oapi               *endpoint.OpenAPI
	searchGeoLocations func(service.FilterGeoLocation) ([]service.GeoLocationView, error)
	zipcodeUpdate      func(string) (*service.Region, error)
}

func (handler *GeoLocationHandler) SearchGeoLocations() (string, string, echo.HandlerFunc) {
	return endpoint.Echo(
		endpoint.Get("/api/search/geo-locations"),
		handler.oapi.Route("GeoLocation.List", `Return a search of GeoLocationView`),
		func(in endpoint.EndpointInput[any, any, struct {
			service.FilterGeoLocation
			Context string `json:"context,omitempty"`
		}, any]) (res endpoint.DataResponse[endpoint.CollectionItemData[service.GeoLocationView]], err error) {

			filter := in.Query.FilterGeoLocation

			if filter.CEP != nil {
				_, _ = handler.zipcodeUpdate(*filter.CEP)
			}

			items, err := handler.searchGeoLocations(in.Query.FilterGeoLocation)
			if err != nil {
				return res, errors.Wrap(err, "search geo locations")
			}

			res.Context = in.Query.Context
			res.Data.DataDetail.Kind = "searchGeoLocations"
			res.Data.Items = items
			return res, nil
		},
	)
}
