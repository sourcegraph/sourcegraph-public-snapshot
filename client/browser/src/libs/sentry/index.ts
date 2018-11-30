import * as Sentry from '@sentry/browser'

import storage from '../../browser/storage'
import { isInPage } from '../../context'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { CodeHost } from '../code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'

export function initSentry(script: 'content' | 'options' | 'background'): void {
    if (isInPage) {
        // Don't run Sentry on Phabricator.
        return
    }

    Sentry.init({
        dsn: 'https://32613b2b6a5b4da2aa50660a60297d79@sentry.io/1334031',
    })

    Sentry.configureScope(async scope => {
        scope.setTag('script', script)

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

    storage.observeSync('featureFlags').subscribe(flags => {
        Sentry.configureScope(scope => {
            scope.setTag('extensions', flags.useExtensions ? 'enabled' : 'disabled')
        })
    })
}
