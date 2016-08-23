package e2etestuser

// Prefix is prefixed to all e2etest user account logins to ensure they can be
// filtered out of different systems easily and do not conflict with real user
// accounts.
const Prefix = "e2etestuserx4FF3"

// UserAgent is explicitly filtered out on several metrics/monitoring systems.
const UserAgent = "Sourcegraph e2etest-bot"
