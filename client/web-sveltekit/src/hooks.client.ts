import { handleErrorWithSentry } from '@sentry/sveltekit'
import * as Sentry from '@sentry/sveltekit'

import { asError } from '$lib/common'

Sentry.init({
    // Disabled if dsn is undefined
    dsn: window.context.sentryDSN ?? undefined,
    release: window.context.version,
    initialScope: {
        user: window.context.currentUser
            ? {
                  id: window.context.currentUser.databaseID,
                  graphqlID: window.context.currentUser.id,
                  username: window.context.currentUser.username,
                  displayName: window.context.currentUser.displayName,
                  email: window.context.currentUser.emails.find(email => email.isPrimary)?.email,
              }
            : undefined,
        tags: {
            app: 'sveltekit',
            siteID: window.context.siteID,
            externalURL: window.context.externalURL,
        },
    },

    tracesSampleRate: 1.0,
})

export const handleError = handleErrorWithSentry(({ error }: { error: unknown }) => {
    // When throwing _expected errors_ in data loaders via SvelteKit's `error`function,
    // the error is wrapped in an object with `status` and `body` properties (an instance
    // of `HTTPError`).
    // Usually SvelteKit will unwrap the error itself and actually not call this function.
    // But this doesn't work in production builds with bazel due to the fact that the `HTTPError`
    // class is defined multiple times, which makes the `instanceof` check fail (it's not
    // clear why the class is defined multiple times).
    // By unwrapping and returning the error here we can still render the proper error message
    // in the UI, otherwise it would show a generic "Internal Error" message.
    if (error && typeof error === 'object' && 'body' in error) {
        return error.body as Error
    }
    return asError(error)
})
