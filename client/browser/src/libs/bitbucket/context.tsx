import { RepoSpec, RevSpec } from '../../../../../shared/src/util/url'
import { CodeHostContext } from '../code_intelligence/code_intelligence'

// example pathname: /projects/TEST/repos/some-repo/browse/src/extension.ts
const PATH_REGEX = /^\/projects\/([^\/]+)\/repos\/([^\/]+)\//

function getRepoSpecFromLocation(location: Pick<Location, 'hostname' | 'pathname'>): RepoSpec {
    const { hostname, pathname } = location
    const match = pathname.match(PATH_REGEX)
    if (!match) {
        throw new Error(`location pathname does not match path regex: ${pathname}`)
    }
    const [, projectName, repoName] = match
    return {
        repoName: `${hostname}/${projectName}/${repoName}`,
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
    try {
        revisionRefInfo = revisionRefStr && JSON.parse(revisionRefStr)
    } catch {
        throw new Error(`Could not parse revisionRefStr: ${revisionRefStr}`)
    }
    if (revisionRefInfo && revisionRefInfo.latestCommit) {
        return {
            rev: revisionRefInfo.latestCommit,
        }
    } else {
        throw new Error(`revisionRefInfo is empty or has no latestCommit (revisionRefStr: ${revisionRefStr})`)
    }
}

export function getContext(): CodeHostContext {
    const repoSpec = getRepoSpecFromLocation(window.location)
    let revSpec: Partial<RevSpec> = {}
    try {
        revSpec = getRevSpecFromRevisionSelector()
    } catch (err) {
        // RevSpec is optional in CodeHostContext, log the error for debug purposes
        console.error('Could not determine revSpec from revision selector', err)
    }
    return {
        ...repoSpec,
        ...revSpec,
    }
}
