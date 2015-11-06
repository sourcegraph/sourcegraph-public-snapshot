package issues

import (
	"errors"
	"time"
)

var Issues = &IssueService{}

type IssueService struct {
}

type IssueSpec struct {
	Repo string
	ID   int64
}

type ReplySpec struct {
	ReplyNum int
	IssueSpec
}

func (s *IssueService) Get(spec IssueSpec) (*Issue, error) {
	issueList, err := s.List(spec.Repo)
	if err != nil {
		return nil, err
	}
	for _, issue := range issueList.Issues {
		if spec.ID == issue.UID {
			return issue, nil
		}
	}
	return nil, ErrNotFound
}

func (s *IssueService) List(repo string) (*IssueList, error) {
	return readIssues(repo)
}

func (s *IssueService) Upsert(repo string, issue *Issue) (*Issue, error) {
	id, err := writeIssue(repo, &issue.issueInternal)
	if err != nil {
		return nil, err
	}
	issue.UID = int64(id)
	return issue, nil
}

func (s *IssueService) UpdateReply(spec IssueSpec, replyNum int, body string) (*Issue, error) {
	issue, err := Issues.Get(spec)
	if err != nil {
		return nil, err
	}

	reply := issue.GetEvent(replyNum)
	reply.Body = body
	issue, err = Issues.Upsert(spec.Repo, issue)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *IssueService) CreateReply(spec IssueSpec, body string, authorID int32) (*Event, error) {
	issue, err := s.Get(spec)
	if err != nil {
		return nil, err
	}
	issue.Events = append(issue.Events, Event{
		Created:   time.Now(),
		UID:       len(issue.Events) + 1,
		AuthorUID: authorID,
		Body:      body,
	})
	_, err = writeIssue(spec.Repo, &issue.issueInternal)
	return &issue.Events[len(issue.Events)-1], err
}

var ErrNotFound = errors.New("not found")
var ErrNotAuthorized = errors.New("not authorized")
