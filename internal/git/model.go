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
	FeatureSHA   string
	FeatureSubj  string
	ReviewSHA    string
	ReviewSubj   string
	ReviewBranch string
	// DiffHash     string
	MergeRequest int
}
