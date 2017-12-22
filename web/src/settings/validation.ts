/**
 * Regular expression to identify valid organization names.
 */
export const VALID_ORG_NAME_REGEXP = /^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$/

/**
 * Regular expression to identify valid username.
 */
export const VALID_USERNAME_REGEXP = /^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$/

/**
 * Regular expression to identify valid password.
 */
export const VALID_PASSWORD_REGEXP = /^.{6,}$/
