package git

type Commit struct {
	SHA     string
	Subject string
}

type Branch struct {
	SHA  string
	Name string
}

type Record struct {
	ID           int
	FeatureSHA   string
	FeatureSubj  string
	ReviewSHA    string
	ReviewSubj   string
	ReviewBranch string
}
