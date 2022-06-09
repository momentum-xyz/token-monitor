package types

type ERC1155MetaData struct {
	Description  string `json:"description"`
	ExternalURL  string `json:"external_url"`
	Image        string `json:"image"`
	AnimationURL string `json:"animation_url"`
	Name         string `json:"name"`
	Attributes   []struct {
		TraitType string `json:"trait_type"`
		Value     string `json:"value"`
	} `json:"attributes"`
}
