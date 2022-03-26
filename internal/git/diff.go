package git

import (
	"crypto/sha256"
	"fmt"
)

func (g *git) DiffHash(sha string) (string, error) {
	diff, err := g.run("diff", "--unified=0", sha+"~", sha)
	if err != nil {
		return "", err
	}

	diff = diff[2:]

	hash := sha256.New()
	for _, line := range diff {
		hash.Write([]byte(line))
	}

	sum := hash.Sum(nil)

	strSum := fmt.Sprintf("%x", sum)

	if g.debug {
		fmt.Printf("--- diff %s\n", sha)
		for _, line := range diff {
			fmt.Println(line)
		}
		fmt.Printf("Commit: %s DiffHash: %s\n", sha, strSum)
	}

	return strSum, nil
}
