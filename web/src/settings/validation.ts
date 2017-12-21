/**
 * Regular expression to identify valid organization names.
 */
export const VALID_ORG_NAME_REGEXP = /^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$/i

/**
 * Regular expression to identify valid username.
 */
export const VALID_USERNAME_REGEXP = /^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$/i

/**
 * Regular expression to identify valid password.
 */
export const VALID_PASSWORD_REGEXP = /^.{6,}$/
