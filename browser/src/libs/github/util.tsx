import { DiffResolvedRevSpec } from '../../shared/repo'
import { RepoSpec, RevSpec, FileSpec } from '../../../../shared/src/util/url'

/**
 * getFileContainers returns the elements on the page which should be marked
 * up with tooltips & links:
 *
 * 1. blob view: a single file
 * 2. commit view: one or more file diffs
 * 3. PR conversation view: snippets with inline comments
 * 4. PR unified/split view: one or more file diffs
 */
export function getFileContainers(): HTMLCollectionOf<HTMLElement> {
    return document.getElementsByClassName('file') as HTMLCollectionOf<HTMLElement>
}

/**
 * getDeltaFileName returns the path of the file container. It assumes
 * the file container is for a diff (i.e. a commit or pull request view).
 */
export function getDeltaFileName(container: HTMLElement): { headFilePath: string; baseFilePath: string | null } {
    const info = container.querySelector('.file-info') as HTMLElement

    if (info.title) {
        // for PR conversation snippets
        return getPathNamesFromElement(info)
    }
    const link = info.querySelector('a') as HTMLElement
    return getPathNamesFromElement(link)
}

function getPathNamesFromElement(element: HTMLElement): { headFilePath: string; baseFilePath: string | null } {
    const elements = element.title.split(' → ')
    if (elements.length > 1) {
        return { headFilePath: elements[1], baseFilePath: elements[0] }
    }
    return { headFilePath: elements[0], baseFilePath: elements[0] }
}

/**
 * getDiffResolvedRev returns the base and head revision SHA, or null for non-diff views.
 */
export function getDiffResolvedRev(codeView: HTMLElement): DiffResolvedRevSpec | null {
    const { pageType } = parseURL()
    if (!isDeltaPageType(pageType)) {
        return null
    }

    let baseCommitID = ''
    let headCommitID = ''
    const fetchContainers = document.getElementsByClassName(
        'js-socket-channel js-updatable-content js-pull-refresh-on-pjax'
    )
    const isCommentedSnippet = codeView.classList.contains('js-comment-container')
    if (pageType === 'pull') {
        if (fetchContainers && fetchContainers.length === 1) {
            // tslint:disable-next-line
            for (let i = 0; i < fetchContainers.length; ++i) {
                // for conversation view of pull request
                const el = fetchContainers[i] as HTMLElement
                const url = el.getAttribute('data-url')
                if (!url) {
                    continue
                }

                const urlSplit = url.split('?')
                const query = urlSplit[1]
                const querySplit = query.split('&')
                for (const kv of querySplit) {
                    const kvSplit = kv.split('=')
                    const k = kvSplit[0]
                    const v = kvSplit[1]
                    if (k === 'base_commit_oid') {
                        baseCommitID = v
                    }
                    if (k === 'end_commit_oid') {
                        headCommitID = v
                    }
                }
            }
        } else if (isCommentedSnippet) {
            const resolvedDiffSpec = getResolvedDiffFromCommentedSnippet(codeView)
            if (resolvedDiffSpec) {
                return resolvedDiffSpec
            }
        } else {
            // Last-ditch: look for inline comment form input which has base/head on it.
            const baseInput = document.querySelector(`input[name="comparison_start_oid"]`)
            if (baseInput) {
                baseCommitID = (baseInput as HTMLInputElement).value
            }
            const headInput = document.querySelector(`input[name="comparison_end_oid"]`)
            if (headInput) {
                headCommitID = (headInput as HTMLInputElement).value
            }
        }
    } else if (pageType === 'commit') {
        // Refined GitHub adds a `.patch-diff-links` element
        const shaContainers = document.querySelectorAll('.sha-block:not(.patch-diff-links)')
        if (shaContainers && shaContainers.length === 2) {
            const baseShaEl = shaContainers[0].querySelector('a')
            if (baseShaEl) {
                // e.g "https://github.com/gorilla/mux/commit/0b13a922203ebdbfd236c818efcd5ed46097d690"
                baseCommitID = baseShaEl.href.split('/').slice(-1)[0]
            }
            const headShaEl = shaContainers[1].querySelector('span.sha') as HTMLElement
            if (headShaEl) {
                headCommitID = headShaEl.innerHTML
            }
        }
    } else if (pageType === 'compare') {
        const resolvedDiffSpec = getResolvedDiffForCompare()
        if (resolvedDiffSpec) {
            return resolvedDiffSpec
        }
    }

    if (baseCommitID === '' || headCommitID === '') {
        return getDiffResolvedRevFromPageSource(document.documentElement.innerHTML, pageType === 'pull')
    }
    return { baseCommitID, headCommitID }
}

// ".../files/(BASE..)?HEAD#diff-DIFF"
// https://github.com/sourcegraph/codeintellify/pull/77/files/e8ffee0c59e951d29bcc7cff7d58caff1c5c97c2..ce472adbfc6ac8ccf1bf7afbe71f18505ca994ec#diff-8a128e9e8a5a8bb9767f5f5392391217
// https://github.com/lguychard/sourcegraph-configurable-references/pull/1/files/fa32ce95d666d73cf4cb3e13b547993374eb158d#diff-45327f86d4438556066de133327f4ca2
const COMMENTED_SNIPPET_DIFF_REGEX = /\/files\/((\w+)\.\.)?(\w+)#diff-\w+$/

function getResolvedDiffFromCommentedSnippet(codeView: HTMLElement): DiffResolvedRevSpec | null {
    // For commented snippets, try to get the HEAD commit ID from the file header,
    // as it will always be the most accurate (for example in the case of outdated snippets).
    const linkToFile: HTMLLinkElement | null = codeView.querySelector('.file-header a')
    if (!linkToFile) {
        return null
    }
    const match = linkToFile.href.match(COMMENTED_SNIPPET_DIFF_REGEX)
    if (!match) {
        return null
    }
    const headCommitID = match[3]
    // The file header may not contain the base commit ID, so we get it from the page source.
    const resolvedRevFromPageSource = getDiffResolvedRevFromPageSource(document.documentElement.innerHTML, true)
    return headCommitID && resolvedRevFromPageSource
        ? {
              ...resolvedRevFromPageSource,
              headCommitID,
          }
        : null
}

function getResolvedDiffForCompare(): DiffResolvedRevSpec | undefined {
    const branchElements = document.querySelectorAll<HTMLElement>('.commitish-suggester .select-menu-button span')
    if (branchElements && branchElements.length === 2) {
        return { baseCommitID: branchElements[0].innerText, headCommitID: branchElements[1].innerText }
    }
    return undefined
}

function getDiffResolvedRevFromPageSource(pageSource: string, isPullRequest: boolean): DiffResolvedRevSpec | null {
    if (!isPullRequest) {
        return null
    }
    const baseShaComment = '<!-- base sha1: &quot;'
    const baseIndex = pageSource.indexOf(baseShaComment)

    if (baseIndex === -1) {
        return null
    }

    const headShaComment = '<!-- head sha1: &quot;'
    const headIndex = pageSource.indexOf(headShaComment, baseIndex)
    if (headIndex === -1) {
        return null
    }

    const baseCommitID = pageSource.substr(baseIndex + baseShaComment.length, 40)
    const headCommitID = pageSource.substr(headIndex + headShaComment.length, 40)
    return {
        baseCommitID,
        headCommitID,
    }
}

/**
 * Scrapes the branch name from the branch select menu element
 * present on GitHub blob pages.
 */
export function getBranchName(): string {
    const branchSelectMenu = document.querySelector('.branch-select-menu') as HTMLElement
    if (!branchSelectMenu) {
        throw new Error('Could not find .branch-select-menu')
    }
    const cssTruncateTarget = branchSelectMenu.querySelector('span.css-truncate-target') as HTMLElement
    if (!cssTruncateTarget) {
        throw new Error('Could not find span.css-truncate-target')
    }
    if (!cssTruncateTarget.innerText.endsWith('…')) {
        // The branch name is not truncated
        return cssTruncateTarget.innerText
    }
    // When the branch name is truncated, it is stored in full
    // in the select menu button's title attribute.
    const selectButton = cssTruncateTarget.closest('.select-menu-button') as HTMLElement
    if (!selectButton) {
        throw new Error('Could not find .select-menu-button')
    }
    return selectButton.title
}

type GitHubURL =
    | ({ pageType: 'tree' | 'commit' | 'pull' | 'compare' | 'other' } & RepoSpec)
    | ({ pageType: 'blob'; revAndFilePath: string } & RepoSpec)

export function isDeltaPageType(pageType: GitHubURL['pageType']): boolean {
    switch (pageType) {
        case 'commit':
        case 'pull':
        case 'compare':
            return true
        default:
            return false
    }
}

export function parseURL(loc: Pick<Location, 'host' | 'pathname'> = window.location): GitHubURL {
    const { host, pathname } = loc
    const [user, ghRepoName, pageType, ...rest] = pathname.slice(1).split('/')
    if (!user || !ghRepoName) {
        throw new Error(`Could not parse repoName from GitHub url: ${window.location}`)
    }
    const repoName = `${host}/${user}/${ghRepoName}`
    switch (pageType) {
        case 'blob':
            return {
                pageType,
                repoName,
                revAndFilePath: rest.join('/'),
            }
        case 'tree':
        case 'pull':
        case 'commit':
        case 'compare':
            return {
                pageType,
                repoName,
            }
        default:
            return { pageType: 'other', repoName }
    }
}
