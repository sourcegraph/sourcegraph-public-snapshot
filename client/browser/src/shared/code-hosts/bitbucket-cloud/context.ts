import type { CodeHostContext } from '../shared/codeHost'
import { RepoURLParseError } from '../shared/errors'

export async function getContext(): Promise<CodeHostContext> {
    const rawRepoName = getRawRepoName()

    return Promise.resolve({
        rawRepoName,
        revision: '',
        privateRepository: false,
    })
}

export function getRawRepoName(): string {
    const { host, pathname } = location
    const [user, repoName] = pathname.slice(1).split('/')
    if (!user || !repoName) {
        throw new RepoURLParseError(`Could not parse repoName from Bitbucket Cloud url: ${location.href}`)
    }

    return `${host}/${user}/${repoName}`
}
