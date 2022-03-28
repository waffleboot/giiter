package git

type Message struct {
	Subject     string
	Description string
}

type Commit struct {
	SHA     string
	Message Message
}

type Branch struct {
	SHA  string
	Name string
}

type Record struct {
	ID           int
	FeatureSHA   string
	FeatureMsg   Message
	ReviewSHA    string
	ReviewMsg    Message
	ReviewBranch string
}

func (r *Record) IsNewCommit() bool {
	return r.ReviewSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}
