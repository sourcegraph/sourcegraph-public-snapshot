// This file contains information about the features supported by the Sourcegraph server.
//
// TODO(sqs): It should probably be populated directly by the server (and be respected
// by the server) instead of being computed here. That change can be made gradually.

/**
 * Whether the server can list all of its repositories. False for Sourcegraph.com,
 * which is a mirror of all public GitHub.com repositories.
 */
export const canListAllRepositories = window.context.onPrem || window.context.debug

/**
 * Whether the application should show the user marketing elements (links, etc.)
 * that are intended for Sourcegraph.com.
 */
export const showDotComMarketing = !window.context.onPrem || window.context.debug

/**
 * Whether the application supports the user "forgot-password" flow.
 *
 * TODO(sqs): This actually is determined by whether the user is non-SSO and
 * the server has Mandrill enabled, but we don't yet have a way of knowing that
 * here and this is a good enough proxy.
 */
export const userForgotPassword = !window.context.onPrem

/**
 * Whether the signup form should show terms and privacy policy links.
 */
export const signupTerms = !window.context.onPrem
