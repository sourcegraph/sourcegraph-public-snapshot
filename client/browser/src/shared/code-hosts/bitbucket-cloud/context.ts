import { CodeHostContext } from '../shared/codeHost'
import { RepoURLParseError } from '../shared/errors'

export function getContext(): CodeHostContext {
    const rawRepoName = getRawRepoName()

    return {
        rawRepoName,
        revision: '',
        privateRepository: false,
    }
}

export function getRawRepoName(): string {
    const { host, pathname } = location
    const [user, repoName] = pathname.slice(1).split('/')
    if (!user || !repoName) {
        throw new RepoURLParseError(`Could not parse repoName from Bitbucket Cloud url: ${location.href}`)
    }

    return `${host}/${user}/${repoName}`
}
