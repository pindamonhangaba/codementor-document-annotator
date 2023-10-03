package service

// Region is the model for the response of api cepaberto
type Region struct {
	Altitude    *float64 `json:"altitude"`
	Cep         string   `json:"cep"`
	Latitude    string   `json:"latitude"`
	Longitude   string   `json:"longitude"`
	Logradouro  string   `json:"logradouro"`
	Bairro      string   `json:"bairro"`
	Complemento string   `json:"complemento"`
	Cidade      City     `json:"cidade"`
	Estado      State    `json:"estado"`
}

//City is the complement of the region model
type City struct {
	Altitude *float64 `json:"altitude"`
	Cep      string   `json:"cep"`
	Ddd      int64    `json:"ddd"`
	Ibge     string   `json:"ibge"`
	Nome     string   `json:"nome"`
}

//State is the complement of the region model
type State struct {
	Sigla string `json:"sigla"`
}

// CepabertoGetter interface service
type CepabertoGetter struct {
	Run func(string) (*Region, error)
}
