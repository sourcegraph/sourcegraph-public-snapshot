import { handleErrorWithSentry } from '@sentry/sveltekit'
import * as Sentry from '@sentry/sveltekit'

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

// If you have a custom error handler, pass it to `handleErrorWithSentry`
export const handleError = handleErrorWithSentry()
