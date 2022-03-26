package git

import (
	"crypto/sha256"
	"fmt"
)

func (g *git) DiffHash(commit string) (string, error) {
	diff, err := g.run("diff", "--unified=0", commit+"~", commit)
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
		fmt.Printf("--- diff %s\n", commit)
		for _, line := range diff {
			fmt.Println(line)
		}
		fmt.Printf("Commit: %s DiffHash: %s\n", commit, strSum)
	}

	return strSum, nil
}
