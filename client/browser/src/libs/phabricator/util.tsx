import { ChangeState, DifferentialState, DiffusionState, PhabricatorMode, RevisionState } from '.'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { getRepoDetailsFromCallsign, getRepoDetailsFromDifferentialID } from './backend'

const TAG_PATTERN = /r([0-9A-z]+)([0-9a-f]{40})/
function matchPageTag(): RegExpExecArray | null {
    const el = document.getElementsByClassName('phui-tag-core').item(0)
    if (!el) {
        return null
    }
    return TAG_PATTERN.exec(el.children[0].getAttribute('href') as string)
}

function getCallsignFromPageTag(): string | null {
    const match = matchPageTag()
    if (!match) {
        return null
    }
    return match[1]
}

function getCommitIDFromPageTag(): string | null {
    const match = matchPageTag()
    if (!match) {
        return null
    }
    return match[2]
}

function isDifferentialLanded(): boolean {
    const closedElement = document.getElementsByClassName('visual-only phui-icon-view phui-font-fa fa-check-square-o')
    if (closedElement.length === 0) {
        return false
    }
    return true
}

const DIFF_LINK = /D[0-9]+\?id=([0-9]+)/i
function getMaxDiffFromTabView(): { diffID: number; revDescription: string } | null {
    // first, find Revision contents table box
    const headerShells = document.getElementsByClassName('phui-header-header')
    let revisionContents: Element | null = null
    for (const headerShell of Array.from(headerShells)) {
        if (headerShell.textContent === 'Revision Contents') {
            revisionContents = headerShell
        }
    }
    if (!revisionContents) {
        return null
    }
    const parentContainer = revisionContents.parentElement!.parentElement!.parentElement!.parentElement!.parentElement!
    const tables = parentContainer.getElementsByClassName('aphront-table-view')
    for (const table of Array.from(tables)) {
        const tableRows = (table as HTMLTableElement).rows
        const row = tableRows[0]
        // looking for the history tab of the revision contents table
        if (row.children[0].textContent !== 'Diff') {
            continue
        }
        const links = table.getElementsByTagName('a')
        let max: { diffID: number; revDescription: string } | null = null
        for (const link of Array.from(links)) {
            const linkHref = link.getAttribute('href')
            if (!linkHref) {
                continue
            }
            const matches = DIFF_LINK.exec(linkHref)
            if (!matches) {
                continue
            }
            if (!link.parentNode!.parentNode!.childNodes[2].childNodes[0]) {
                continue
            }
            const revDescription = (link.parentNode!.parentNode!.childNodes[2].childNodes[0] as any).href
            const shaMatch = TAG_PATTERN.exec(revDescription)
            if (!shaMatch) {
                continue
            }
            max =
                max && max.diffID > parseInt(matches[1], 10)
                    ? max
                    : { diffID: parseInt(matches[1], 10), revDescription: shaMatch[2] }
        }
        return max
    }
    return null
}

const DIFF_PATTERN = /Diff ([0-9]+)/
function getDiffIdFromDifferentialPage(): string | null {
    const diffsContainer = document.getElementById('differential-review-stage')
    if (!diffsContainer) {
        console.error(`no element with id differential-review-stage found on page.`)
        return null
    }
    const wrappingDiffBox = diffsContainer.parentElement
    if (!wrappingDiffBox) {
        console.error(`parent container of diff container not found.`)
        return null
    }
    const diffTitle = wrappingDiffBox.children[0].getElementsByClassName('phui-header-header').item(0)
    if (!diffTitle || !diffTitle.textContent) {
        return null
    }
    const matches = DIFF_PATTERN.exec(diffTitle.textContent)
    if (!matches) {
        return null
    }
    return matches[1]
}

// tslint:disable-next-line
const PHAB_DIFFUSION_REGEX = /^\/?(source|diffusion)\/([A-Za-z0-9\-\_]+)\/browse\/([\w-]+\/)?([^;$]+)(;[0-9a-f]{40})?(?:\$[0-9]+)?/i
const PHAB_DIFFERENTIAL_REGEX = /^\/?(D[0-9]+)(?:\?(?:(?:id=([0-9]+))|(vs=(?:[0-9]+|on)&id=[0-9]+)))?/i
const PHAB_REVISION_REGEX = /^\/?r([0-9A-z]+)([0-9a-f]{40})/i
// http://phabricator.aws.sgdev.org/source/nmux/change/master/mux.go
const PHAB_CHANGE_REGEX = /^\/?(source|diffusion)\/([A-Za-z0-9]+)\/change\/([\w-]+)\/([^;]+)(;[0-9a-f]{40})?/i
const PHAB_CHANGESET_REGEX = /^\/?\/differential\/changeset.*/i
const COMPARISON_REGEX = /^vs=((?:[0-9]+|on))&id=([0-9]+)/i

function getBaseCommitIDFromRevisionPage(): string | null {
    const keyElements = document.getElementsByClassName('phui-property-list-key')
    for (const keyElement of Array.from(keyElements)) {
        if (keyElement.textContent === 'Parents ') {
            const parentUrl = ((keyElement.nextSibling as HTMLElement).children[0].children[0] as HTMLLinkElement).href
            const url = new URL(parentUrl)
            const revisionMatch = PHAB_REVISION_REGEX.exec(url.pathname)
            if (revisionMatch) {
                return revisionMatch[2]
            }
        }
    }
    return null
}

export async function getPhabricatorState(
    loc: Location,
    queryGraphQL: PlatformContext['requestGraphQL']
): Promise<DiffusionState | DifferentialState | RevisionState | ChangeState | null> {
    const stateUrl = loc.href.replace(loc.origin, '')
    const diffusionMatch = PHAB_DIFFUSION_REGEX.exec(stateUrl)
    const { protocol, hostname, port } = loc
    if (diffusionMatch) {
        const match = {
            protocol,
            hostname,
            port,
            viewType: diffusionMatch[1],
            callsign: diffusionMatch[2],
            branch: diffusionMatch[3],
            filePath: diffusionMatch[4],
            revInUrl: diffusionMatch[5], // only on previous versions
        }
        if (match.branch && match.branch.endsWith('/')) {
            // Remove trailing slash (included b/c branch group is optional)
            match.branch = match.branch.substr(match.branch.length - 1)
        }

        const callsign = getCallsignFromPageTag()
        if (!callsign) {
            console.error('could not locate callsign for differential page')
            return null
        }
        match.callsign = callsign
        const { repoName } = await getRepoDetailsFromCallsign(callsign, queryGraphQL)
        const commitID = getCommitIDFromPageTag()
        if (!commitID) {
            console.error('cannot determine commitIDision from page')
            return null
        }
        return {
            repoName,
            filePath: match.filePath,
            mode: PhabricatorMode.Diffusion,
            commitID,
        }
    }
    const differentialMatch = PHAB_DIFFERENTIAL_REGEX.exec(stateUrl)
    if (differentialMatch) {
        const match = {
            protocol,
            hostname,
            port,
            differentialID: differentialMatch[1],
            diffID: differentialMatch[6],
            comparison: differentialMatch[7],
        }

        const differentialID = parseInt(match.differentialID.split('D')[1], 10)
        let diffID = match.diffID ? parseInt(match.diffID, 10) : undefined

        const { callsign } = await getRepoDetailsFromDifferentialID(differentialID, queryGraphQL)
        if (!callsign) {
            console.error(`callsign not found`)
            return null
        }
        if (!diffID) {
            const fromPage = getDiffIdFromDifferentialPage()
            if (fromPage) {
                diffID = parseInt(fromPage, 10)
            }
        }
        if (!diffID) {
            console.error(`differential id not found on page.`)
            return null
        }
        const { repoName } = await getRepoDetailsFromCallsign(callsign, queryGraphQL)
        let baseRev = `phabricator/base/${diffID}`
        let headRev = `phabricator/diff/${diffID}`

        let leftDiffID: number | undefined

        const maxDiff = getMaxDiffFromTabView()
        const diffLanded = isDifferentialLanded()
        if (diffLanded && !maxDiff) {
            console.error(
                'looking for the final diff id in the revision contents table failed. expected final row to have the commit in the description field.'
            )
            return null
        }
        if (match.comparison) {
            // urls that looks like this: http://phabricator.aws.sgdev.org/D3?vs=on&id=8&whitespace=ignore-most#toc
            // if the first parameter (vs=) is not 'on', not sure how to handle
            const comparisonMatch = COMPARISON_REGEX.exec(match.comparison)!
            const leftID = comparisonMatch[1]
            if (leftID !== 'on') {
                leftDiffID = parseInt(leftID, 10)
                baseRev = `phabricator/diff/${leftDiffID}`
            } else {
                baseRev = `phabricator/base/${comparisonMatch[2]}`
            }
            headRev = `phabricator/diff/${comparisonMatch[2]}`
            if (diffLanded && maxDiff && comparisonMatch[2] === `${maxDiff.diffID}`) {
                headRev = maxDiff.revDescription
                baseRev = headRev.concat('~1')
            }
        } else {
            // check if the diff we are viewing is the max diff. if so,
            // right is the merged rev into master, and left is master~1
            if (diffLanded && maxDiff && diffID === maxDiff.diffID) {
                headRev = maxDiff.revDescription
                baseRev = maxDiff.revDescription.concat('~1')
            }
        }
        return {
            baseRepoName: repoName,
            baseRev,
            headRepoName: repoName,
            headRev, // This will be blank on GitHub, but on a manually staged instance should exist
            differentialID,
            diffID,
            leftDiffID,
            mode: PhabricatorMode.Differential,
        }
    }

    const revisionMatch = PHAB_REVISION_REGEX.exec(stateUrl)
    if (revisionMatch) {
        const match = {
            protocol,
            hostname,
            port,
            callsign: revisionMatch[1],
            rev: revisionMatch[2],
        }
        const { repoName } = await getRepoDetailsFromCallsign(match.callsign, queryGraphQL)
        const headCommitID = match.rev
        const baseCommitID = getBaseCommitIDFromRevisionPage()
        if (!baseCommitID) {
            console.error(`did not successfully determine parent revision.`)
            return null
        }
        return {
            repoName,
            baseCommitID,
            headCommitID,
            mode: PhabricatorMode.Revision,
        }
    }

    const changeMatch = PHAB_CHANGE_REGEX.exec(stateUrl)
    if (changeMatch) {
        const match = {
            protocol: changeMatch[1],
            hostname: changeMatch[2],
            tld: changeMatch[3],
            port: changeMatch[4],
            viewType: changeMatch[5],
            callsign: changeMatch[6],
            branch: changeMatch[7],
            filePath: changeMatch[8],
            revInUrl: changeMatch[9], // only on previous versions
        }

        const callsign = getCallsignFromPageTag()
        if (!callsign) {
            console.error('could not locate callsign for differential page')
            return null
        }
        match.callsign = callsign
        const { repoName } = await getRepoDetailsFromCallsign(callsign, queryGraphQL)
        const commitID = getCommitIDFromPageTag()
        if (!commitID) {
            console.error('cannot determine revision from page.')
            return null
        }
        return {
            repoName,
            filePath: match.filePath,
            mode: PhabricatorMode.Change,
            commitID,
        }
    }

    const changesetMatch = PHAB_CHANGESET_REGEX.exec(stateUrl)
    if (changesetMatch) {
        const crumbs = document.querySelector('.phui-crumbs-view')
        if (!crumbs) {
            throw new Error('failed parsing changeset dom')
        }

        const [, differentialHref, diffHref] = crumbs.querySelectorAll('a')

        const differentialMatch = differentialHref.getAttribute('href')!.match(/D(\d+)/)
        if (!differentialMatch) {
            throw new Error('failed parsing differentialID')
        }
        const differentialID = parseInt(differentialMatch[1], 10)

        const diffMatch = diffHref.getAttribute('href')!.match(/\/differential\/diff\/(\d+)/)
        if (!diffMatch) {
            throw new Error('failed parsing diffID')
        }
        const diffID = parseInt(diffMatch[1], 10)

        const { callsign } = await getRepoDetailsFromDifferentialID(differentialID, queryGraphQL)
        if (!callsign) {
            console.error(`callsign not found`)
            return null
        }

        const { repoName } = await getRepoDetailsFromCallsign(callsign, queryGraphQL)
        let baseRev = `phabricator/base/${diffID}`
        let headRev = `phabricator/diff/${diffID}`

        const maxDiff = getMaxDiffFromTabView()
        const diffLanded = isDifferentialLanded()
        if (diffLanded && !maxDiff) {
            console.error(
                'looking for the final diff id in the revision contents table failed. expected final row to have the commit in the description field.'
            )
            return null
        }

        // check if the diff we are viewing is the max diff. if so,
        // right is the merged rev into master, and left is master~1
        if (diffLanded && maxDiff && diffID === maxDiff.diffID) {
            headRev = maxDiff.revDescription
            baseRev = maxDiff.revDescription.concat('~1')
        }

        return {
            baseRepoName: repoName,
            baseRev,
            headRepoName: repoName,
            headRev, // This will be blank on GitHub, but on a manually staged instance should exist
            differentialID,
            diffID,
            mode: PhabricatorMode.Differential,
        }
    }

    return null
}

/**
 * This hacks javelin Stratcom to ignore command + click actions on sg-clickable tokens.
 * Without this, two windows open when a user command + clicks on a token.
 *
 * TODO could this be eliminated with shadow DOM?
 */
export function metaClickOverride(): void {
    const JX = (window as any).JX
    if (JX.Stratcom._dispatchProxyPreMeta) {
        return
    }
    JX.Stratcom._dispatchProxyPreMeta = JX.Stratcom._dispatchProxy
    JX.Stratcom._dispatchProxy = (proxyEvent: {
        __auto__type: string
        __auto__rawEvent: KeyboardEvent
        __auto__target: HTMLElement
    }) => {
        if (
            proxyEvent.__auto__type === 'click' &&
            proxyEvent.__auto__rawEvent.metaKey &&
            proxyEvent.__auto__target.classList.contains('sg-clickable')
        ) {
            return
        }
        return JX.Stratcom._dispatchProxyPreMeta(proxyEvent)
    }
}

export function normalizeRepoName(origin: string): string {
    let repoName = origin
    repoName = repoName.replace('\\', '')
    if (origin.startsWith('git@')) {
        repoName = origin.substr('git@'.length)
        repoName = repoName.replace(':', '/')
    } else if (origin.startsWith('git://')) {
        repoName = origin.substr('git://'.length)
    } else if (origin.startsWith('https://')) {
        repoName = origin.substr('https://'.length)
    } else if (origin.includes('@')) {
        // Assume the origin looks like `username@host:repo/path`
        const split = origin.split('@')
        repoName = split[1]
        repoName = repoName.replace(':', '/')
    }
    return repoName.replace(/.git$/, '')
}
