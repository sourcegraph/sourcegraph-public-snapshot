// Package fs implements issues.Service using a filesystem.
package fs

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/apps/issues/issues"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

// NewService ...
func NewService(rootDir string) issues.Service {
	return service{
		root: rootDir,
	}
}

type service struct {
	// root directory for issue storage for all repos.
	root string
}

// dir returns the path to root of issue storage for the given repo.
func (s service) dir(repo issues.RepoSpec) string {
	return filepath.Join(s.root, filepath.FromSlash(repo.URI))
}

func (s service) List(ctx context.Context, repo issues.RepoSpec, opt issues.IssueListOptions) ([]issues.Issue, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	var is []issues.Issue

	dirs, err := readDirIDs(s.dir(repo))
	if err != nil {
		return is, err
	}
	for i := len(dirs); i > 0; i-- {
		dir := dirs[i-1]
		if !dir.IsDir() {
			continue
		}

		var issue issue
		err = jsonDecodeFile(filepath.Join(s.dir(repo), dir.Name(), "0"), &issue)
		if err != nil {
			return is, err
		}

		if issue.State != opt.State {
			continue
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
		if err != nil {
			return is, err
		}
		is = append(is, issues.Issue{
			ID:    dir.ID,
			State: issue.State,
			Title: issue.Title,
			Comment: issues.Comment{
				User: issues.User{
					Login:     user.Login,
					AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
					HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
				},
				CreatedAt: issue.CreatedAt,
			},
		})
	}

	return is, nil
}

func (s service) Count(_ context.Context, repo issues.RepoSpec, opt issues.IssueListOptions) (uint64, error) {
	var count uint64

	dirs, err := readDirIDs(s.dir(repo))
	if err != nil {
		return 0, err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		var issue issue
		err = jsonDecodeFile(filepath.Join(s.dir(repo), dir.Name(), "0"), &issue)
		if err != nil {
			return 0, err
		}

		if issue.State != opt.State {
			continue
		}

		count++
	}

	return count, nil
}

func (s service) Get(ctx context.Context, repo issues.RepoSpec, id uint64) (issues.Issue, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	var issue issue
	err := jsonDecodeFile(filepath.Join(s.dir(repo), formatUint64(id), "0"), &issue)
	if err != nil {
		return issues.Issue{}, err
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}
	return issues.Issue{
		ID:    id,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			User: issues.User{
				Login:     user.Login,
				AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
				HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
			},
			CreatedAt: issue.CreatedAt,
		},
	}, nil
}

func (s service) ListComments(ctx context.Context, repo issues.RepoSpec, id uint64, opt interface{}) ([]issues.Comment, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	var comments []issues.Comment

	dir := filepath.Join(s.dir(repo), formatUint64(id))
	fis, err := readDirIDs(dir)
	if err != nil {
		return comments, err
	}
	for _, fi := range fis {
		var comment comment
		err = jsonDecodeFile(filepath.Join(dir, fi.Name()), &comment)
		if err != nil {
			return comments, err
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
		if err != nil {
			return comments, err
		}
		comments = append(comments, issues.Comment{
			User: issues.User{
				Login:     user.Login,
				AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
				HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
			},
			CreatedAt: comment.CreatedAt,
			Body:      comment.Body,
		})
	}

	return comments, nil
}

func (s service) ListEvents(ctx context.Context, repo issues.RepoSpec, id uint64, opt interface{}) ([]issues.Event, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	var events []issues.Event

	dir := filepath.Join(s.dir(repo), formatUint64(id), "events")
	fis, err := readDirIDs(dir)
	if err != nil {
		return events, err
	}
	for _, fi := range fis {
		var event event
		err = jsonDecodeFile(filepath.Join(dir, fi.Name()), &event)
		if err != nil {
			return events, err
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: event.ActorUID})
		if err != nil {
			return events, err
		}
		events = append(events, issues.Event{
			Actor: issues.User{
				Login:     user.Login,
				AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
				HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
			},
			CreatedAt: event.CreatedAt,
			Type:      event.Type,
			Rename:    event.Rename,
		})
	}

	return events, nil
}

func (s service) CreateComment(ctx context.Context, repo issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	comment := comment{
		AuthorUID: putil.UserFromContext(ctx).UID,
		CreatedAt: time.Now(),
		Body:      c.Body,
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
	if err != nil {
		return issues.Comment{}, err
	}

	// Commit to storage.
	dir := filepath.Join(s.dir(repo), formatUint64(id))
	commentID, err := nextID(dir)
	if err != nil {
		return issues.Comment{}, err
	}
	err = jsonEncodeFile(filepath.Join(dir, formatUint64(commentID)), comment)
	if err != nil {
		return issues.Comment{}, err
	}

	return issues.Comment{
		User: issues.User{
			Login:     user.Login,
			AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
			HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
		},
		CreatedAt: comment.CreatedAt,
		Body:      comment.Body,
	}, nil
}

func (s service) Create(ctx context.Context, repo issues.RepoSpec, i issues.Issue) (issues.Issue, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	issue := issue{
		State: issues.OpenState,
		Title: i.Title,
		comment: comment{
			AuthorUID: putil.UserFromContext(ctx).UID,
			CreatedAt: time.Now(),
			Body:      i.Body,
		},
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}

	// Commit to storage.
	issueID, err := nextID(s.dir(repo))
	if err != nil {
		return issues.Issue{}, err
	}
	dir := filepath.Join(s.dir(repo), formatUint64(issueID))
	err = os.MkdirAll(s.dir(repo), 0755) // Only needed for first issue in the repo. TODO: Consider MkdirAll or even better?
	if err != nil {
		return issues.Issue{}, err
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return issues.Issue{}, err
	}
	err = os.Mkdir(filepath.Join(dir, "events"), 0755)
	if err != nil {
		return issues.Issue{}, err
	}
	err = jsonEncodeFile(filepath.Join(dir, "0"), issue)
	if err != nil {
		return issues.Issue{}, err
	}

	return issues.Issue{
		ID:    issueID,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			User: issues.User{
				Login:     user.Login,
				AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
				HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
			},
			CreatedAt: issue.CreatedAt,
			Body:      issue.Body,
		},
	}, nil
}

func (s service) Edit(ctx context.Context, repo issues.RepoSpec, id uint64, ir issues.IssueRequest) (issues.Issue, error) {
	if err := ir.Validate(); err != nil {
		return issues.Issue{}, err
	}

	sg := sourcegraph.NewClientFromContext(ctx)

	// Get from storage.
	var issue issue
	err := jsonDecodeFile(filepath.Join(s.dir(repo), formatUint64(id), "0"), &issue)
	if err != nil {
		return issues.Issue{}, err
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}

	if ir.State != nil {
		issue.State = *ir.State
	}
	origTitle := issue.Title
	if ir.Title != nil {
		issue.Title = *ir.Title
	}

	// Commit to storage.
	err = jsonEncodeFile(filepath.Join(s.dir(repo), formatUint64(id), "0"), issue)
	if err != nil {
		return issues.Issue{}, err
	}

	// THINK: Is this the best place to do this? Should it be returned from this func? How would GH backend do it?
	// Create event and commit to storage.
	eventID, err := nextID(filepath.Join(s.dir(repo), formatUint64(id), "events"))
	if err != nil {
		return issues.Issue{}, err
	}
	event := event{
		ActorUID:  putil.UserFromContext(ctx).UID,
		CreatedAt: time.Now(),
	}
	switch {
	case ir.State != nil && *ir.State == issues.OpenState:
		event.Type = issues.Reopened
	case ir.State != nil && *ir.State == issues.ClosedState:
		event.Type = issues.Closed
	case ir.Title != nil:
		event.Type = issues.Renamed
		event.Rename = &issues.Rename{
			From: origTitle,
			To:   *ir.Title,
		}
	}
	err = jsonEncodeFile(filepath.Join(s.dir(repo), formatUint64(id), "events", formatUint64(eventID)), event)
	if err != nil {
		return issues.Issue{}, err
	}

	return issues.Issue{
		ID:    id,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			User: issues.User{
				Login:     user.Login,
				AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
				HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
			},
			CreatedAt: issue.CreatedAt,
		},
	}, nil
}

// nextID returns the next id for the given dir. If there are no previous elements, it begins with id 1.
func nextID(dir string) (uint64, error) {
	fis, err := readDirIDs(dir)
	if err != nil {
		return 0, err
	}
	if len(fis) == 0 {
		return 1, nil
	}
	return fis[len(fis)-1].ID + 1, nil
}

// TODO.
func (service) CurrentUser(ctx context.Context) (issues.User, error) {
	sg := sourcegraph.NewClientFromContext(ctx)
	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: putil.UserFromContext(ctx).UID})
	if err != nil {
		return issues.User{}, err
	}
	return issues.User{
		Login:     user.Login,
		AvatarURL: avatarURL(user.Login),                            //template.URL(user.AvatarURL),
		HTMLURL:   template.URL("https://github.com/" + user.Login), // TODO.
	}, nil
}

var (
	gh        = github.NewClient(nil)
	ghAvatars = make(map[string]template.URL)
)

// TODO.
func avatarURL(login string) template.URL {
	if avatarURL, ok := ghAvatars[login]; ok {
		return avatarURL
	}

	user, _, err := gh.Users.Get(login)
	if err != nil || user.AvatarURL == nil {
		return ""
	}
	ghAvatars[login] = template.URL(*user.AvatarURL + "&s=96")
	return ghAvatars[login]
}

func jsonDecodeFile(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(v)
	_ = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func jsonEncodeFile(path string, v interface{}) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(v)
	_ = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func formatUint64(n uint64) string { return strconv.FormatUint(n, 10) }
