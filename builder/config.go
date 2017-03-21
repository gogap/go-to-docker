package builder

type BranchTag struct {
	Server       string   `json:"server"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Organization string   `json:"organization"`
	Tags         []string `json:"tags"`
}

type BranchTagsConfig struct {
	Branchs map[string]BranchTag `branchs`
}
