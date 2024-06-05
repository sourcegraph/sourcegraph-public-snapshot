import { handleErrorWithSentry, replayIntegration } from '@sentry/sveltekit'
import * as Sentry from '@sentry/sveltekit'

Sentry.init({
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

    // This sets the sample rate to be 10%. You may want this to be 100% while
    // in development and sample at a lower rate in production
    replaysSessionSampleRate: 0.1,

    // If the entire session is not sampled, use the below sample rate to sample
    // sessions when an error occurs.
    replaysOnErrorSampleRate: 1.0,

    // If you don't want to use Session Replay, just remove the line below:
    integrations: [replayIntegration()],
})

// If you have a custom error handler, pass it to `handleErrorWithSentry`
export const handleError = handleErrorWithSentry()
