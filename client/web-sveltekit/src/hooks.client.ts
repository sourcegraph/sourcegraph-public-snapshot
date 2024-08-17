import { handleErrorWithSentry } from '@sentry/sveltekit'
import * as Sentry from '@sentry/sveltekit'
import type { HandleClientError, HttpError, Redirect } from '@sveltejs/kit'

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

function isRedirect(value: unknown): value is Redirect {
    return !!value && typeof value === 'object' && 'status' in value && 'location' in value
}

function isHttpError(value: unknown): value is HttpError {
    return !!value && typeof value === 'object' && 'status' in value && 'body' in value
}

// This class is used to make TypeScript happy. `handleError` has to return an `Error` object.
export class SgRedirect extends Error {
    constructor(public redirect: Redirect) {
        super('Redirect')
    }
}

const sentryErrorHandler = handleErrorWithSentry()

export const handleError: HandleClientError = input => {
    // When throwing _expected errors_ in data loaders via SvelteKit's `error`function,
    // the error is wrapped in an object with `status` and `body` properties (an instance
    // of `HTTPError`).
    // Usually SvelteKit will unwrap the error itself and actually not call this function.
    // But this doesn't work in production builds with bazel due to the fact that the `HTTPError`
    // class is defined multiple times, which makes the `instanceof` check fail (it's not
    // clear why the class is defined multiple times).
    // By unwrapping and returning the error here we can still render the proper error message
    // in the UI, otherwise it would show a generic "Internal Error" message.
    if (isHttpError(input.error)) {
        return input.error.body
    }

    // The same applies to redirects, which are instances of `Redirect`, but are also not
    // detected correctly in production builds. In this case we "manually" perform the
    // redirect (which is not quite what SvelteKit would do; SvelteKit would call
    // `beforeNavigation` handlers and do client side navigation if the redirect is to
    // a known route).
    // The redirect is performed in the `+error.svelte` component. We can't perform the redirect
    // here because this function is also run during preloading and we don't want to make a
    // sudden redirect when the user is just hovering over a link.
    if (isRedirect(input.error)) {
        return new SgRedirect(input.error)
    }

    // This issue is tracked in SRCH-926

    // We call the Sentry error handler here to report the error to Sentry. We wrap the
    // Sentry handle in our own handler instead of the other way around, because
    // otherwise sentry would report redirects as errors.

    // @ts-expect-error Unclear why Sentry's error handler type doesn't match
    return sentryErrorHandler(input)
}
