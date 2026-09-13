package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// reviewThreadsQuery fetches all review threads (inline comment threads) for a
// pull request, including every comment in each thread. path/line/diffSide
// live on both the thread and the comment in GitHub's schema; we read them
// from the comment so replies keep their own position when it differs.
const reviewThreadsQuery = `
query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 50, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          isResolved
          isOutdated
          path
          line
          startLine
          diffSide
          resolvedBy { login }
          comments(first: 100) {
            totalCount
            pageInfo { hasNextPage endCursor }
            nodes {
              id
              fullDatabaseId
              url
              path
              line
              originalLine
              startLine
              originalStartLine
              outdated
              body
              createdAt
              updatedAt
              author { login }
              replyTo { id }
              commit { oid }
              reactions(first: 50) {
                nodes { content }
              }
            }
          }
        }
      }
    }
  }
}
`

type actor struct {
	Login string `json:"login"`
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type reactionNode struct {
	Content string `json:"content"`
}

type reviewComment struct {
	ID                string    `json:"id"`
	FullDatabaseID    string    `json:"fullDatabaseId"`
	URL               string    `json:"url"`
	Path              string    `json:"path"`
	Line              *int      `json:"line"`
	OriginalLine      *int      `json:"originalLine"`
	StartLine         *int      `json:"startLine"`
	OriginalStartLine *int      `json:"originalStartLine"`
	Outdated          bool      `json:"outdated"`
	Body              string    `json:"body"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Author            *actor    `json:"author"`
	ReplyTo           *struct {
		ID string `json:"id"`
	} `json:"replyTo"`
	Commit *struct {
		OID string `json:"oid"`
	} `json:"commit"`
	Reactions struct {
		Nodes []reactionNode `json:"nodes"`
	} `json:"reactions"`
}

type reviewThread struct {
	ID         string `json:"id"`
	IsResolved bool   `json:"isResolved"`
	IsOutdated bool   `json:"isOutdated"`
	Path       string `json:"path"`
	Line       *int   `json:"line"`
	StartLine  *int   `json:"startLine"`
	DiffSide   string `json:"diffSide"`
	ResolvedBy *actor `json:"resolvedBy"`
	Comments   struct {
		TotalCount int             `json:"totalCount"`
		PageInfo   pageInfo        `json:"pageInfo"`
		Nodes      []reviewComment `json:"nodes"`
	} `json:"comments"`
}

type graphqlError struct {
	Message string `json:"message"`
}

type reviewThreadsResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					PageInfo pageInfo       `json:"pageInfo"`
					Nodes    []reviewThread `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
	Errors []graphqlError `json:"errors"`
}

// fetchReviewThreads pulls every review thread (and every comment within it)
// for the given PR, following pagination until reviewThreads is exhausted.
func fetchReviewThreads(owner, repo string, number int) ([]reviewThread, error) {
	var threads []reviewThread
	cursor := ""

	for {
		args := []string{
			"api", "graphql",
			"-f", "query=" + reviewThreadsQuery,
			"-F", "owner=" + owner,
			"-F", "repo=" + repo,
			"-F", "number=" + strconv.Itoa(number),
		}
		if cursor != "" {
			args = append(args, "-F", "cursor="+cursor)
		}

		cmd := exec.Command("gh", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("gh api graphql failed: %w: %s", err, stderr.String())
		}

		var resp reviewThreadsResponse
		if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
			return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
		}
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("GraphQL error: %s", resp.Errors[0].Message)
		}

		rt := resp.Data.Repository.PullRequest.ReviewThreads
		threads = append(threads, rt.Nodes...)

		for _, t := range rt.Nodes {
			if t.Comments.PageInfo.HasNextPage {
				fmt.Fprintf(os.Stderr, "warning: thread %s has more than 100 comments; some replies were not fetched\n", t.ID)
			}
		}

		if !rt.PageInfo.HasNextPage {
			break
		}
		cursor = rt.PageInfo.EndCursor
	}

	return threads, nil
}
