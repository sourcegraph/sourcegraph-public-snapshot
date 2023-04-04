import { USER_DISPLAY_NAME_MAX_LENGTH, USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'

/**
 * Regular expression to identify valid team names.
 */
export const VALID_TEAM_NAME_REGEXP = VALID_USERNAME_REGEXP

/** Maximum allowed length for a team name. */
export const TEAM_NAME_MAX_LENGTH = USERNAME_MAX_LENGTH

/** Maximum allowed length for a team display name. */
export const TEAM_DISPLAY_NAME_MAX_LENGTH = USER_DISPLAY_NAME_MAX_LENGTH
