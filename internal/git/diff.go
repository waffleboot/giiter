package git

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type nullHash struct {
	hash  string
	valid bool
}

func (g *git) DiffHash(sha string) (nullHash, error) {
	files, err := g.run("diff-tree", "-r", "--name-only", sha)
	if err != nil {
		return nullHash{}, err
	}

	files = files[1:]

	hash := sha256.New()

	for i := range files {
		file := files[i]
		diff, err := g.run("diff-tree", "--unified=0", sha, "--", file)
		if err != nil {
			return nullHash{}, err
		}

		diff = diff[2:]

		if strings.HasPrefix(diff[0], "new file") {
			hash.Write([]byte("new file"))
			hash.Write([]byte(file))
			continue
		}

		if strings.HasPrefix(diff[0], "deleted file") {
			hash.Write([]byte("deleted file"))
			hash.Write([]byte(file))
			continue
		}

		if strings.HasPrefix(diff[1], "Binary") {
			return nullHash{valid: false}, nil
		}

		diff = diff[3:]

		for _, line := range diff {
			hash.Write([]byte(line))
		}

		if g.debug {
			fmt.Printf("--- diff %s %s\n", sha, file)
			for _, line := range diff {
				fmt.Println(line)
			}
		}

	}

	sum := hash.Sum(nil)

	strSum := fmt.Sprintf("%x", sum)

	if g.debug {
		fmt.Printf("Commit: %s DiffHash: %s\n", sha, strSum)
	}

	return nullHash{hash: strSum, valid: true}, nil
}
