package model

type CreateCloudflareIPRulesModel struct {
	Configuration struct {
		Target string `json:"target"`
		Value  string `json:"value"`
	} `json:"configuration"`
	Mode  string `json:"mode"`
	Notes string `json:"notes"`
}
