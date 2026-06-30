package github

import (
	"time"
)

type WorkflowJobPayload struct {
	Action      string      `json:"action"`
	WorkflowJob WorkflowJob `json:"workflow_job,omitempty"`
	Workflow    Workflow    `json:"workflow,omitempty"`
	WorkflowRun WorkflowRun `json:"workflow_run,omitempty"`
	Repository  Repository  `json:"repository"`
	Sender      Sender      `json:"sender"`
}

type WorkflowJob struct {
	Id              int       `json:"id"`
	RunId           int       `json:"run_id"`
	WorkflowName    string    `json:"workflow_name"`
	HeadBranch      string    `json:"head_branch"`
	RunUrl          string    `json:"run_url"`
	RunAttempt      int       `json:"run_attempt"`
	NodeId          string    `json:"node_id"`
	HeadSha         string    `json:"head_sha"`
	Url             string    `json:"url"`
	HtmlUrl         string    `json:"html_url"`
	Status          string    `json:"status"`
	Conclusion      string    `json:"conclusion"`
	CreatedAt       time.Time `json:"created_at"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	Name            string    `json:"name"`
	Steps           []Step    `json:"steps"`
	CheckRunUrl     string    `json:"check_run_url"`
	Labels          []string  `json:"labels"`
	RunnerId        int       `json:"runner_id"`
	RunnerName      string    `json:"runner_name"`
	RunnerGroupId   int       `json:"runner_group_id"`
	RunnerGroupName string    `json:"runner_group_name"`
}

type Repository struct {
	Id                        int       `json:"id"`
	NodeId                    string    `json:"node_id"`
	Name                      string    `json:"name"`
	FullName                  string    `json:"full_name"`
	Private                   bool      `json:"private"`
	Owner                     Owner     `json:"owner"`
	HtmlUrl                   string    `json:"html_url"`
	Description               string    `json:"description"`
	Fork                      bool      `json:"fork"`
	Url                       string    `json:"url"`
	ForksUrl                  string    `json:"forks_url"`
	KeysUrl                   string    `json:"keys_url"`
	CollaboratorsUrl          string    `json:"collaborators_url"`
	TeamsUrl                  string    `json:"teams_url"`
	HooksUrl                  string    `json:"hooks_url"`
	IssueEventsUrl            string    `json:"issue_events_url"`
	EventsUrl                 string    `json:"events_url"`
	AssigneesUrl              string    `json:"assignees_url"`
	BranchesUrl               string    `json:"branches_url"`
	TagsUrl                   string    `json:"tags_url"`
	BlobsUrl                  string    `json:"blobs_url"`
	GitTagsUrl                string    `json:"git_tags_url"`
	GitRefsUrl                string    `json:"git_refs_url"`
	TreesUrl                  string    `json:"trees_url"`
	StatusesUrl               string    `json:"statuses_url"`
	LanguagesUrl              string    `json:"languages_url"`
	StargazersUrl             string    `json:"stargazers_url"`
	ContributorsUrl           string    `json:"contributors_url"`
	SubscribersUrl            string    `json:"subscribers_url"`
	SubscriptionUrl           string    `json:"subscription_url"`
	CommitsUrl                string    `json:"commits_url"`
	GitCommitsUrl             string    `json:"git_commits_url"`
	CommentsUrl               string    `json:"comments_url"`
	CompareUrl                string    `json:"compare_url"`
	MergesUrl                 string    `json:"merges_url"`
	ArchiveUrl                string    `json:"archive_url"`
	DownloadsUrl              string    `json:"downloads_url"`
	IssuesUrl                 string    `json:"issues_url"`
	PullsUrl                  string    `json:"pulls_url"`
	MilestonesUrl             string    `json:"milestones_url"`
	NotificationsUrl          string    `json:"notifications_url"`
	LabelsUrl                 string    `json:"labels_url"`
	ReleasesUrl               string    `json:"releases_url"`
	DeploymentsUrl            string    `json:"deployments_url"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	PushedAt                  time.Time `json:"pushed_at"`
	GitUrl                    string    `json:"git_url"`
	SshUrl                    string    `json:"ssh_url"`
	CloneUrl                  string    `json:"clone_url"`
	SvnUrl                    string    `json:"svn_url"`
	Homepage                  string    `json:"homepage"`
	Size                      int       `json:"size"`
	StargazersCount           int       `json:"stargazers_count"`
	WatchersCount             int       `json:"watchers_count"`
	Language                  string    `json:"language"`
	HasIssues                 bool      `json:"has_issues"`
	HasProjects               bool      `json:"has_projects"`
	HasDownloads              bool      `json:"has_downloads"`
	HasWiki                   bool      `json:"has_wiki"`
	HasPages                  bool      `json:"has_pages"`
	HasDiscussions            bool      `json:"has_discussions"`
	ForksCount                int       `json:"forks_count"`
	MirrorUrl                 string    `json:"mirror_url"`
	Archived                  bool      `json:"archived"`
	Disabled                  bool      `json:"disabled"`
	OpenIssuesCount           int       `json:"open_issues_count"`
	License                   string    `json:"license"`
	AllowForking              bool      `json:"allow_forking"`
	IsTemplate                bool      `json:"is_template"`
	WebCommitSignoffRequired  bool      `json:"web_commit_signoff_required"`
	HasPullRequests           bool      `json:"has_pull_requests"`
	PullRequestCreationPolicy string    `json:"pull_request_creation_policy"`
	Topics                    []string  `json:"topics"` // TODO: check this
	Visibility                string    `json:"visibility"`
	Forks                     int       `json:"forks"`
	OpenIssues                int       `json:"open_issues"`
	Watchers                  int       `json:"watchers"`
	DefaultBranch             string    `json:"default_branch"`
}

type Owner struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravitarUrl       string `json:"gravitar_url"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type Sender struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravitarUrl       string `json:"gravitar_url"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type Step struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	Number      int       `json:"number"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type WorkflowRun struct {
	Id                  int        `json:"id"`
	Name                string     `json:"name"`
	NodeId              string     `json:"node_id"`
	HeadBranch          string     `json:"head_branch"`
	HeadSha             string     `json:"head_sha"`
	Path                string     `json:"path"`
	DisplayTitle        string     `json:"display_title"`
	RunNumber           int        `json:"run_number"`
	Event               string     `json:"event"`
	Status              string     `json:"status"`
	Conclusion          string     `json:"conclusion"`
	WorkflowId          int        `json:"workflow_id"`
	CheckSuiteId        int        `json:"check_suite_id"`
	CheckSuiteNodeId    string     `json:"check_suite_node_id"`
	Url                 string     `json:"url"`
	HtmlUrl             string     `json:"html_url"`
	PullRequests        []string   `json:"pull_requests"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Actor               Actor      `json:"actor"`
	RunAttempt          int        `json:"run_attempt"`
	ReferencedWorkflows []Workflow `json:"referenced_workflows"` // Todo: double check this is the correct type
	RunStartedAt        time.Time  `json:"run_started_at"`
	TriggeringActor     Actor      `json:"triggering_actor"`
	JobsUrl             string     `json:"jobs_url"`
	LogsUrl             string     `json:"logs_url"`
	CheckSuiteUrl       string     `json:"check_suite_url"`
	ArtifactsUrl        string     `json:"artifacts_url"`
	CancelUrl           string     `json:"cancel_url"`
	RerunUrl            string     `json:"rerun_url"`
	PreviousAttemptUrl  string     `json:"previous_attempt_url"`
	WorkflowUrl         string     `json:"workflow_url"`
	HeadCommit          Commit     `json:"head_commit"`
	Repository          Repository `json:"repository"`
	HeadRepository      Repository `json:"head_repository"`
}

type Actor struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravitarUrl       string `json:"gravitar_url"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type Workflow struct{}

type Commit struct {
	Id        string    `json:"id"`
	TreeId    string    `json:"tree_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Author    Committer `json:"author"`
	Committer Committer `json:"committer"`
}

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
