package gitlab

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const baseURL = "https://gitlab.com/api/v4/projects/34829333/merge_requests"

type MergeRequest struct {
	Title        string
	SourceBranch string
	TargetBranch string
}

func CreateMergeRequest(ctx context.Context, token string, mr MergeRequest) error {
	data := url.Values{
		"id":            {"34829333"},
		"title":         {mr.Title},
		"source_branch": {mr.SourceBranch},
		"target_branch": {mr.TargetBranch},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		_, _ = io.Copy(os.Stdout, resp.Body)

		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}
