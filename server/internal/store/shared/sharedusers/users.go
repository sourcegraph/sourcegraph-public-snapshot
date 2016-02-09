package sharedusers

// DefaultAvatarURL is the default AvatarURL for users that don't have AvatarURL set.
//
// This URL can have a "&s=128"-like suffix appended to set the size.
// That allows it to be compatible with User.AvatarURLOfSize.
const DefaultAvatarURL = "https://secure.gravatar.com/avatar?d=mm&f=y"
