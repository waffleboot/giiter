package git

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/waffleboot/giiter/internal/app"
)

type nullHash struct {
	hash  string
	valid bool
}

func DiffHash(ctx context.Context, sha string) (nullHash, error) {
	files, err := run(ctx, "diff-tree", "-r", "--name-only", sha)
	if err != nil {
		return nullHash{}, err
	}

	files = files[1:]

	hash := sha256.New()

	for i := range files {
		file := files[i]
		diff, err := run(ctx, "diff-tree", "--unified=0", sha, "--", file)
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

		if app.Config.Debug {
			fmt.Printf("--- diff %s %s\n", sha, file)
			for _, line := range diff {
				fmt.Println(line)
			}
		}

	}

	sum := hash.Sum(nil)

	strSum := fmt.Sprintf("%x", sum)

	if app.Config.Debug {
		fmt.Printf("Commit: %s DiffHash: %s\n", sha, strSum)
	}

	return nullHash{hash: strSum, valid: true}, nil
}
