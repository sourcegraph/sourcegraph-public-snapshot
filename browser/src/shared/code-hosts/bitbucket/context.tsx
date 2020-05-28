import { RawRepoSpec, RevSpec } from '../../../../../shared/src/util/url'
import { CodeHostContext } from '../shared/codeHost'

// example pathname: /projects/TEST/repos/some-repo/browse/src/extension.ts
const PATH_REGEX = /\/projects\/([^/]+)\/repos\/([^/]+)\//

function getRawRepoSpecFromLocation(location: Pick<Location, 'hostname' | 'pathname'>): RawRepoSpec {
    const { hostname, pathname } = location
    const match = pathname.match(PATH_REGEX)
    if (!match) {
        throw new Error(`location pathname does not match path regex: ${pathname}`)
    }
    const [, projectName, repoName] = match
    return {
        rawRepoName: `${hostname}/${projectName}/${repoName}`,
    }
}

interface RevisionRefInfo {
    latestCommit?: string
}

function getRevSpecFromRevisionSelector(): RevSpec {
    const branchNameElement = document.querySelector('#repository-layout-revision-selector .name[data-revision-ref]')
    if (!branchNameElement) {
        throw new Error('branchNameElement not found')
    }
    const revisionRefStr = branchNameElement.getAttribute('data-revision-ref')
    let revisionRefInfo: RevisionRefInfo | null = null
    if (revisionRefStr) {
        try {
            revisionRefInfo = JSON.parse(revisionRefStr)
        } catch {
            throw new Error(`Could not parse revisionRefStr: ${revisionRefStr}`)
        }
    }
    if (revisionRefInfo?.latestCommit) {
        return {
            rev: revisionRefInfo.latestCommit,
        }
    }
    throw new Error(`revisionRefInfo is empty or has no latestCommit (revisionRefStr: ${String(revisionRefStr)})`)
}

export function getContext(): CodeHostContext {
    const repoSpec = getRawRepoSpecFromLocation(window.location)
    let revSpec: Partial<RevSpec> = {}
    try {
        revSpec = getRevSpecFromRevisionSelector()
    } catch {
        // RevSpec is optional in CodeHostContext
    }
    return {
        ...repoSpec,
        ...revSpec,
        privateRepository: window.location.hostname !== 'bitbucket.org',
    }
}
