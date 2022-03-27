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

func (r *Record) IsNewCommit() bool {
	return r.ReviewSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}
