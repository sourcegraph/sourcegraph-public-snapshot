/**
 * We rely on the site configuration instead of setting cascade
 * because we don't want to give ability to non-admin users to
 * override (on/off search jobs UI) on the user level.
 *
 * It's on for all users on the instance, or it's off for all users.
 * Only admins can turn it on/off in the site configuration page.
 */
export const isSearchJobsEnabled = (): boolean => window.context?.experimentalFeatures?.searchJobs ?? false
