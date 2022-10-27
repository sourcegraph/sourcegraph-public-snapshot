// This file contains information about the features supported by the Sourcegraph server.
//
// TODO(sqs): It should probably be populated directly by the server (and be respected
// by the server) instead of being computed here. That change can be made gradually.

/**
 * Whether the application should show the user marketing elements (links, etc.)
 * that are intended for Sourcegraph.com.
 */
export const showDotComMarketing = window.context?.sourcegraphDotComMode
