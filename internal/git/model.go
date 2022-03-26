package git

type Commit struct {
	SHA     string
	Subject string
}

type Branch struct {
	SHA  string
	Name string
}
