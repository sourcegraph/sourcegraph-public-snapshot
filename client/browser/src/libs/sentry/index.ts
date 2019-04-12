import * as Sentry from '@sentry/browser'
import { once } from 'lodash'

import { getExtensionVersionSync } from '../../browser/runtime'
import storage from '../../browser/storage'
import { isInPage } from '../../context'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { CodeHost } from '../code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'

const isExtensionStackTrace = (stacktrace: Sentry.Stacktrace, extensionID: string): boolean =>
    !!(stacktrace.frames && stacktrace.frames.some(({ filename }) => !!(filename && filename.includes(extensionID))))

const callSentryInit = once((extensionID: string) => {
    Sentry.init({
        dsn: 'https://32613b2b6a5b4da2aa50660a60297d79@sentry.io/1334031',
        beforeSend: event => {
            // Filter out events if we can tell from the stack trace that
            // they didn't originate from extension code.
            let keep = true
            if (event.exception && event.exception.values) {
                keep = event.exception.values.some(
                    ({ stacktrace }) => !!(stacktrace && isExtensionStackTrace(stacktrace, extensionID))
                )
            } else if (event.stacktrace) {
                keep = isExtensionStackTrace(event.stacktrace, extensionID)
            }
            return keep ? event : null
        },
    })
})

/** Initialize Sentry for error reporting. */
export function initSentry(script: 'content' | 'options' | 'background'): void {
    if (process.env.NODE_ENV !== 'production') {
        return
    }

    storage.observeSync('featureFlags').subscribe(flags => {
        const allowed = flags.allowErrorReporting

        // Don't initialize if user hasn't allowed us to report errors or in Phabricator.
        if (!allowed || isInPage) {
            return
        }

        callSentryInit(browser.runtime.id)

        Sentry.configureScope(async scope => {
            scope.setTag('script', script)
            scope.setTag('extension_version', getExtensionVersionSync())

            const codeHosts: CodeHost[] = [bitbucketServerCodeHost, githubCodeHost, gitlabCodeHost, phabricatorCodeHost]
            for (const { check, name } of codeHosts) {
                const is = await Promise.resolve(check())
                if (is) {
                    scope.setTag('code_host', name)
                }
                return
            }
        })

        storage.observeSync('sourcegraphURL').subscribe(url => {
            Sentry.configureScope(scope => {
                scope.setTag('using_dot_com', url === 'https://sourcegraph.com' ? 'true' : 'false')
            })
        })
    })
}
