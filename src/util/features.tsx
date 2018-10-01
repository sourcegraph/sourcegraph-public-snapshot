// This file contains information about the features supported by the Sourcegraph server.
//
// TODO(sqs): It should probably be populated directly by the server (and be respected
// by the server) instead of being computed here. That change can be made gradually.

/**
 * Whether the server can list all of its repositories. False for Sourcegraph.com,
 * which is a mirror of all public GitHub.com repositories.
 */
export const canListAllRepositories = !window.context.sourcegraphDotComMode

/**
 * Whether the application should show the user marketing elements (links, etc.)
 * that are intended for Sourcegraph.com.
 */
export const showDotComMarketing = window.context.sourcegraphDotComMode

/**
 * Whether the signup form should show terms and privacy policy links.
 */
export const signupTerms = window.context.sourcegraphDotComMode
