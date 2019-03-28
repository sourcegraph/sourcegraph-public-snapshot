import { RepoSpec, RevSpec } from '../../../../../shared/src/util/url'
import { CodeHostContext } from '../code_intelligence/code_intelligence'

// example pathname: /projects/TEST/repos/testing/browse/src/extension.ts
const PATH_REGEX = /^\/projects\/(\w+)\/repos\/(\w+)\//

function getRepoSpecFromLocation(location: Pick<Location, 'hostname' | 'pathname'>): RepoSpec | null {
    const { hostname, pathname } = location
    const match = pathname.match(PATH_REGEX)
    if (!hostname || !match) {
        return null
    }
    return {
        repoName: `${hostname}/${match[1]}/${match[2]}`,
    }
}

interface RevisionRefInfo {
    latestCommit?: string
}

function getRevSpecFromRevisionSelector(): RevSpec | null {
    const branchNameElement = document.querySelector(
        '#repository-layout-revision-selector span.name[data-revision-ref]'
    )
    if (!branchNameElement) {
        return null
    }
    const revisionRefStr = branchNameElement.getAttribute('data-revision-ref')
    let revisionRefInfo: RevisionRefInfo | null = null
    try {
        revisionRefInfo = revisionRefStr && JSON.parse(revisionRefStr)
    } catch {
        return null
    }
    if (revisionRefInfo && revisionRefInfo.latestCommit) {
        return {
            rev: revisionRefInfo.latestCommit,
        }
    } else {
        return null
    }
}

export function getContext(): CodeHostContext {
    const repoSpec = getRepoSpecFromLocation(window.location)
    if (!repoSpec) {
        throw new Error('Could not determine bitbucket code host context')
    }
    const revSpec = getRevSpecFromRevisionSelector() || {}
    return {
        ...repoSpec,
        ...revSpec,
    }
}
