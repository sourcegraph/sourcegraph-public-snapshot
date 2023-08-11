import type { Context } from 'mocha'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { getConfig } from '@sourcegraph/shared/src/testing/config'
import type { Driver } from '@sourcegraph/shared/src/testing/driver'

const { gitHubDotComToken } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

interface CloneReposConfig {
    driver: Driver
    mochaContext: Context
    repoSlugs: string[]
}

export async function cloneRepos(config: CloneReposConfig): Promise<void> {
    const { driver, mochaContext, repoSlugs } = config

    // Cloning the repositories takes ~1 minute, so give initialization 2
    // minutes instead of 1 (which would be inherited from
    // `jest.setTimeout(1 * 60 * 1000)` above).
    mochaContext.timeout(5 * 60 * 1000)

    await driver.ensureHasExternalService({
        kind: ExternalServiceKind.GITHUB,
        displayName: 'test-test-github',
        config: JSON.stringify({
            url: 'https://github.com',
            token: gitHubDotComToken,
            repos: repoSlugs,
        }),
        ensureRepos: repoSlugs.map(slug => `github.com/${slug}`),
    })
}
