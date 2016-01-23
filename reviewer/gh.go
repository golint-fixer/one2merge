// Copyright © 2016 See CONTRIBUTORS <ignasi.fosch@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reviewer

import (
	"errors"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// LookupEnv contains the function used to lookup environment variables.
var LookupEnv = os.LookupEnv

// NewGHClient contains the constructor for github.Client.
var NewGHClient = github.NewClient

type PullRequestInfo struct {
	number int // id of the pull request
	score  int
}

type PullRequestInfoList []PullRequestInfo

// GetClient returns a github.Client authenticated.
func GetClient() (*github.Client, error) {
	token, errEnv := LookupEnv("REVIEWER_TOKEN")
	if !errEnv {
		return nil, errors.New("An error occurred getting REVIEWER_TOKEN environment variable")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return NewGHClient(tc), nil
}

func getCommentSuccessScore(comment string) int {
	score := 0
	if strings.Contains(comment, "+1") {
		score++
	}
	if strings.Contains(comment, "-1") {
		score--
	}
	return score
}

func GetPullRequestInfos(client *github.Client, owner string, repo string) (*PullRequestInfoList, error) {
	pullRequests, _, err := client.PullRequests.List(owner, repo, nil)
	if err != nil {
		return nil, err
	}
	pris := make(PullRequestInfoList, len(pullRequests))
	for n, pullRequest := range pullRequests {
		pris[n].number = *pullRequest.Number
		comments, _, err := client.Issues.ListComments(owner, repo, *pullRequest.Number, nil)
		if err != nil {
			return nil, err
		}
		for _, comment := range comments {
			if comment.Body == nil {
				continue
			}
			pris[n].score = getCommentSuccessScore(*comment.Body)
		}
	}
	return &pris, nil
}