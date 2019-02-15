import * as sentry from '@sentry/browser'
import { authenticatedUser } from './auth'

if (window.context.sentryDSN) {
    sentry.init({
        dsn: window.context.sentryDSN,
        release: 'frontend@' + window.context.version,
    })
    authenticatedUser.subscribe(user => {
        sentry.configureScope(scope => {
            if (user) {
                const { id, username, email } = user
                scope.setUser({ id, username, email })
            }
        })
    })
}
