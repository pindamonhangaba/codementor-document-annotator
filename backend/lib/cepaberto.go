package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pindamonhangaba/monoboi/backend/service"
	"github.com/pkg/errors"
)

// CepGetterService service to search for a given BR zipcode
type CepGetterService struct {
	CEPAbertoAPIToken string
	UpdateZipCode     service.UpdateZipCode
}

// Run returns the response from the api cepaberto
func (s *CepGetterService) Run(cep string) (*service.Region, error) {

	region, err := fromCEPAberto(s.CEPAbertoAPIToken, cep)

	if err != nil {
		reg, err := fromViaCEP(cep)
		if err != nil {
			return nil, errors.Wrap(err, "from viacep")
		}
		ltlg := fromAchaCEP(cep, log.New(os.Stderr, "", log.LstdFlags))
		if ltlg != nil {
			reg.Latitude = ltlg.Latitude
			reg.Longitude = ltlg.Longitude
		}
		region = reg
	}
	if region == nil {
		return nil, errors.New("cannot find cep")
	}

	zip := service.CreateZipCode{
		ZipCode:    region.Cep,
		Coordinate: "(" + region.Latitude + "," + region.Longitude + ")",
		Street:     region.Logradouro,
		Complement: region.Complemento,
		District:   region.Bairro,
		Identifier: region.Cidade.Ibge,
	}
	_, _ = s.UpdateZipCode.Run(&zip)
	return region, nil
}

func fromCEPAberto(apiToken, cep string) (*service.Region, error) {
	var region *service.Region
	url := "https://www.cepaberto.com/api/v3/cep?cep=" + cep

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cepaberto request")
	}
	req.Header.Set("Authorization", "Token token="+apiToken)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error client request region")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "viacep body")
	}
	err = json.Unmarshal(body, &region)
	return region, err
}

type viacepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
}

func fromViaCEP(cep string) (*service.Region, error) {
	url := "https://viacep.com.br/ws/" + cep + "/json/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "viacep init")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "viacep request")
	}
	defer resp.Body.Close()

	viacep := viacepResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "viacep body")
	}
	err = json.Unmarshal(body, &viacep)
	if err != nil {
		return nil, errors.Wrap(err, "viacep response")
	}

	c := strings.ReplaceAll(viacep.Cep, "-", "")

	region := &service.Region{
		Cep: c,
		// default lat,long from Salvador
		Latitude:    "-12.668209",
		Longitude:   "-38.605957",
		Logradouro:  viacep.Logradouro,
		Bairro:      viacep.Bairro,
		Complemento: viacep.Complemento,
		Estado: service.State{
			Sigla: viacep.UF,
		},
		Cidade: service.City{
			Ibge: viacep.IBGE,
			Cep:  c,
			Nome: viacep.Localidade,
		},
	}
	return region, nil
}
