import { ChangeState, DifferentialState, DiffusionState, PhabricatorMode, RevisionState } from '.'
import { CodeCell } from '../../shared/repo'
import { getRepoDetailsFromCallsign, getRepoDetailsFromDifferentialID } from './backend'

const TAG_PATTERN = /r([0-9A-z]+)([0-9a-f]{40})/
function matchPageTag(): RegExpExecArray | null {
    const el = document.getElementsByClassName('phui-tag-core').item(0)
    if (!el) {
        return null
    }
    return TAG_PATTERN.exec(el.children[0].getAttribute('href') as string)
}

export function getCallsignFromPageTag(): string | null {
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
            const shaMatch = TAG_PATTERN.exec(revDescription!)
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

// function getCallsignDifferentialPage(): string | null {
//     const mainColumn = document.getElementsByClassName('phui-main-column').item(0)
//     if (!mainColumn) {
//         console.warn('no \'phui-main-column\'[0] class found')
//         return null
//     }
//     const diffDetailBox = mainColumn.children[1]
//     const repositoryTag = diffDetailBox.getElementsByClassName('phui-property-list-value').item(0)
//     if (!repositoryTag) {
//         console.warn('no \'phui-property-list-value\'[0] class found')
//         return null
//     }
//     let diffusionPath = repositoryTag.children[0].getAttribute('href') // e.g. /source/CALLSIGN/ or /diffusion/CALLSIGN/
//     if (!diffusionPath) {
//         console.warn('no diffusion path found')
//         return null
//     }
//     if (diffusionPath.startsWith('/source/')) {
//         diffusionPath = diffusionPath.substr('/source/'.length)
//     } else if (diffusionPath.startsWith('/diffusion/')) {
//         diffusionPath = diffusionPath.substr('/diffusion/'.length)
//     } else {
//         console.error(`unexpected prefix on diffusion path ${diffusionPath}`)
//         return null
//     }
//     return diffusionPath.substr(0, diffusionPath.length - 1)
// }

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

export function getPhabricatorState(
    loc: Location
): Promise<DiffusionState | DifferentialState | RevisionState | ChangeState | null> {
    return new Promise((resolve, reject) => {
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
                resolve(null)
                return
            }
            match.callsign = callsign
            getRepoDetailsFromCallsign(callsign)
                .then(({ repoPath }) => {
                    const commitID = getCommitIDFromPageTag()
                    if (!commitID) {
                        console.error('cannot determine commitIDision from page')
                        resolve(null)
                        return
                    }
                    resolve({
                        repoPath,
                        filePath: match.filePath,
                        mode: PhabricatorMode.Diffusion,
                        commitID,
                    })
                })
                .catch(reject)
            return
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

            getRepoDetailsFromDifferentialID(differentialID)
                .then(({ callsign }) => {
                    if (!callsign) {
                        console.error(`callsign not found`)
                        resolve(null)
                        return
                    }
                    if (!diffID) {
                        const fromPage = getDiffIdFromDifferentialPage()
                        if (fromPage) {
                            diffID = parseInt(fromPage, 10)
                        }
                    }
                    if (!diffID) {
                        console.error(`differential id not found on page.`)
                        resolve(null)
                        return
                    }
                    getRepoDetailsFromCallsign(callsign)
                        .then(({ repoPath }) => {
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
                            resolve({
                                baseRepoPath: repoPath,
                                baseRev,
                                headRepoPath: repoPath,
                                headRev, // This will be blank on GitHub, but on a manually staged instance should exist
                                differentialID,
                                diffID,
                                leftDiffID,
                                mode: PhabricatorMode.Differential,
                            })
                        })
                        .catch(err => {
                            reject(err)
                        })
                })
                .catch(err => {
                    console.log('uhoh', err)
                    reject(err)
                })
            return
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
            getRepoDetailsFromCallsign(match.callsign)
                .then(({ repoPath }) => {
                    const headCommitID = match.rev
                    const baseCommitID = getBaseCommitIDFromRevisionPage()
                    if (!baseCommitID) {
                        console.error(`did not successfully determine parent revision.`)
                        return null
                    }
                    resolve({
                        repoPath,
                        baseCommitID,
                        headCommitID,
                        mode: PhabricatorMode.Revision,
                    })
                })
                .catch(reject)
            return
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
            getRepoDetailsFromCallsign(callsign)
                .then(({ repoPath }) => {
                    const commitID = getCommitIDFromPageTag()
                    if (!commitID) {
                        console.error('cannot determine revision from page.')
                        return null
                    }
                    resolve({
                        repoPath,
                        filePath: match.filePath,
                        mode: PhabricatorMode.Change,
                        commitID,
                    })
                })
                .catch(reject)
            return
        }

        const changesetMatch = PHAB_CHANGESET_REGEX.exec(stateUrl)
        if (changesetMatch) {
            const crumbs = document.querySelector('.phui-crumbs-view')
            if (!crumbs) {
                reject(new Error('failed parsing changeset dom'))
                return
            }

            const [, differentialHref, diffHref] = crumbs.querySelectorAll('a')

            const differentialMatch = differentialHref.getAttribute('href')!.match(/D(\d+)/)
            if (!differentialMatch) {
                reject(new Error('failed parsing differentialID'))
                return
            }
            const differentialID = parseInt(differentialMatch[1], 10)

            const diffMatch = diffHref.getAttribute('href')!.match(/\/differential\/diff\/(\d+)/)
            if (!diffMatch) {
                reject(new Error('failed parsing diffID'))
                return
            }
            const diffID = parseInt(diffMatch[1], 10)

            getRepoDetailsFromDifferentialID(differentialID)
                .then(({ callsign }) => {
                    if (!callsign) {
                        console.error(`callsign not found`)
                        return null
                    }

                    getRepoDetailsFromCallsign(callsign)
                        .then(({ repoPath }) => {
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

                            resolve({
                                baseRepoPath: repoPath,
                                baseRev,
                                headRepoPath: repoPath,
                                headRev, // This will be blank on GitHub, but on a manually staged instance should exist
                                differentialID,
                                diffID,
                                mode: PhabricatorMode.Differential,
                            })
                        })
                        .catch(reject)
                })
                .catch(reject)
            return
        }

        resolve(null)
    })
}

export function getFilepathFromFile(fileContainer: HTMLElement): { filePath: string; baseFilePath?: string } {
    const filePath = fileContainer.children[3].textContent as string
    const metas = fileContainer.querySelectorAll('.differential-meta-notice')
    let baseFilePath: string | undefined
    const movedFilePrefix = 'This file was moved from '
    for (const meta of metas) {
        let metaText = meta.textContent!
        if (metaText.startsWith(movedFilePrefix)) {
            metaText = metaText.substr(0, metaText.length - 1) // remove trailing '.'
            baseFilePath = metaText.split(movedFilePrefix)[1]
            break
        }
    }
    return { filePath, baseFilePath }
}

export function tryGetBlobElement(file: HTMLElement): HTMLElement | null {
    return (
        (file.querySelector('.repository-crossreference') as HTMLElement) ||
        (file.querySelector('.phabricator-source-code-container') as HTMLElement)
    )
}

export function rowIsNotCode(row: HTMLElement): boolean {
    let el = row
    while (el.tagName !== 'TR' && el.parentElement !== null) {
        el = el.parentElement
    }
    return !!el.getAttribute('data-sigil')
}

export function getNodeToConvert(row: HTMLElement): HTMLElement | null {
    if (rowIsNotCode(row)) {
        return null
    }

    return row
}

/**
 * getCodeCellsForAnnotation code cells which should be annotated
 */
export function getCodeCellsForAnnotation(table: HTMLTableElement): CodeCell[] {
    const cells: CodeCell[] = []
    for (const row of Array.from(table.rows)) {
        if (rowIsNotCode(row)) {
            continue
        }
        let line: number // line number of the current line
        let codeCell: HTMLTableDataCellElement // the actual cell that has code inside; each row contains multiple columns
        let isBlameEnabled = false
        if (
            row.cells[0].classList.contains('diffusion-blame-link') ||
            row.cells[0].classList.contains('phabricator-source-blame-skip')
        ) {
            isBlameEnabled = true
        }
        const lineElem = row.cells[isBlameEnabled ? 2 : 0].childNodes[0]
        if (!lineElem) {
            // No line number; this is likely the empty side of an added or removed file in a diff
            continue
        }
        const lineNumber = getlineNumberForCell(lineElem as Element)
        if (!lineNumber) {
            continue
        }
        line = lineNumber
        codeCell = row.cells[isBlameEnabled ? 3 : 1]
        if (!codeCell) {
            continue
        }

        const innerCode = codeCell.querySelector('.blob-code-inner') // ignore extraneous inner elements, like "comment" button on diff views
        cells.push({
            eventHandler: codeCell, // TODO(john): fix
            cell: (innerCode || codeCell) as HTMLElement,
            line,
        })
    }
    return cells
}

/**
 * getCodeCellsForAnnotation code cells which should be annotated
 */
export function getCodeCellsForDifferentialAnnotations(
    table: HTMLTableElement,
    isSplitView: boolean,
    isBase: boolean
): CodeCell[] {
    const cells: CodeCell[] = []
    // tslint:disable-next-line:prefer-for-of
    for (const row of Array.from(table.rows)) {
        if (rowIsNotCode(row)) {
            continue
        }
        if (isSplitView) {
            const base = row.cells[0]
            const head = row.cells[2]
            const baseLine = getlineNumberForCell(base)
            const headLine = getlineNumberForCell(head)
            const baseCodeCell = row.cells[1]
            const headCodeCell = row.cells[4]

            if (isBase && baseLine && baseCodeCell) {
                cells.push({
                    cell: baseCodeCell,
                    eventHandler: baseCodeCell, // TODO(john): fix
                    line: baseLine,
                    isAddition: false,
                    isDeletion: false,
                })
            } else if (!isBase && headLine && headCodeCell) {
                cells.push({
                    cell: headCodeCell,
                    eventHandler: headCodeCell, // TODO(john): fix
                    line: headLine,
                    isAddition: false,
                    isDeletion: false,
                })
            }
        } else {
            const base = row.cells[0]
            const head = row.cells[1]
            const baseLine = getlineNumberForCell(base)
            const headLine = getlineNumberForCell(head)

            // find first cell that is not <th> and also has code in it
            const codeCell = Array.from(row.cells).find((c, idx) => c.tagName !== 'TH' && !c.matches('td.copy'))

            if (isBase && baseLine && codeCell) {
                cells.push({
                    cell: codeCell,
                    eventHandler: codeCell, // TODO(john): fix
                    line: baseLine,
                    isAddition: false,
                    isDeletion: false,
                })
            } else if (!isBase && headLine && codeCell) {
                cells.push({
                    cell: codeCell,
                    eventHandler: codeCell, // TODO(john): fix
                    line: headLine,
                    isAddition: false,
                    isDeletion: false,
                })
            }
        }
    }

    return cells
}

export function getlineNumberForCell(cell: Element): number | undefined {
    // Newer versions of Phabricator (end of 2017) rely on the data-n attribute for setting the line number.
    const lineString = cell.textContent || cell.getAttribute('data-n')
    return parseInt(lineString as string, 10)
}

export const PHAB_PAGE_LOAD_EVENT_NAME = 'phabPageLoaded'

/**
 * This hacks javelin Stratcom to ignore command + click actions on sg-clickable tokens.
 * Without this, two windows open when a user command + clicks on a token.
 */
export function metaClickOverride(): void {
    const JX = (window as any).JX
    if (JX.Stratcom._dispatchProxyPreMeta) {
        return
    }
    JX.Stratcom._dispatchProxyPreMeta = JX.Stratcom._dispatchProxy
    JX.Stratcom._dispatchProxy = proxyEvent => {
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

export function normalizeRepoPath(origin: string): string {
    let repoPath = origin
    repoPath = repoPath.replace('\\', '')
    if (origin.startsWith('git@')) {
        repoPath = origin.substr('git@'.length)
        repoPath = repoPath.replace(':', '/')
    } else if (origin.startsWith('git://')) {
        repoPath = origin.substr('git://'.length)
    } else if (origin.startsWith('https://')) {
        repoPath = origin.substr('https://'.length)
    } else if (origin.includes('@')) {
        // Assume the origin looks like `username@host:repo/path`
        const split = origin.split('@')
        repoPath = split[1]
        repoPath = repoPath.replace(':', '/')
    }
    return repoPath.replace(/.git$/, '')
}

export function getContainerForBlobAnnotation(): {
    file: HTMLElement | null
    diffusionButtonProps: {
        className: string
        iconStyle: any
        style: any
    }
} {
    const diffusionButtonProps = {
        className: 'button grey has-icon msl phui-header-action-link',
        iconStyle: { marginTop: '-1px', paddingRight: '4px', fontSize: '18px', height: '.8em', width: '.8em' },
        style: {},
    }
    const blobContainer = document.querySelector('.repository-crossreference')
    if (blobContainer && blobContainer.parentElement) {
        return { file: blobContainer.parentElement, diffusionButtonProps }
    }

    let file = document.querySelector('.phui-two-column-content.phui-two-column-footer') as HTMLElement
    if (file) {
        diffusionButtonProps.className = 'button button-grey has-icon has-text phui-button-default'
        return { file, diffusionButtonProps }
    }
    file = document.getElementsByClassName('phui-main-column')[0] as HTMLElement

    return { file, diffusionButtonProps }
}
