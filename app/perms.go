package app

type ExternalPerms struct {
	AskForMorePerms bool   // whether the page should ask the user to grant more perms (using GrantURL to initiate the OAuth2 flow)
	GrantURL        string // the URL that will initiate the reauth flow to make perms sufficient (unset if perms are already sufficient)
}
