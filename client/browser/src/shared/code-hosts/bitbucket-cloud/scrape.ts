const clean = (txt: string): string => txt.replace('//', '/')

/** The type of the change of a file in a diff */
const changeTypes = ['modified', 'renamed', 'removed', 'added'] as const
type ChangeType = typeof changeTypes[number]

export function parsePRFilePaths(codeView: HTMLElement, changeType: ChangeType): { base?: string; head?: string } {
    const breadcrumbs = codeView.querySelector<HTMLElement>('[data-qa="bk-filepath"]')?.cloneNode(true)
    const text = breadcrumbs?.textContent

    if (!(breadcrumbs instanceof HTMLElement) || !text) {
        throw new Error('Could not find breadcrumbs component for code view')
    }
    if (changeType === 'added') {
        return { head: text }
    }
    if (changeType === 'removed') {
        return { base: text }
    }
    if (changeType === 'modified') {
        return { base: text, head: text }
    }

    const SEPARATOR = '<>SEPARATOR<>'

    // For moved/renamed files.
    const arrowImageElements = breadcrumbs.querySelectorAll<HTMLElement>('[aria-label="arrow"]')
    for (const arrowImageElement of arrowImageElements) {
        arrowImageElement.textContent = SEPARATOR
    }

    const [start, changed, end] = breadcrumbs.textContent!.split(/[{}]/)
    const [first, second] = changed.split(SEPARATOR)
    return { base: clean(start + first + end), head: clean(start + second + end) }
}

export function parseCommitFilePaths(codeView: HTMLElement, changeType: ChangeType): { base?: string; head?: string } {
    // Null out irrelevant textContent (changeType, etc.)
    const breadcrumbs = codeView.querySelector<HTMLElement>('.filename')?.cloneNode(true)
    if (!breadcrumbs) {
        throw new Error('Could not find breadcrumbs component for code view')
    }

    for (const child of breadcrumbs.childNodes) {
        if (child instanceof HTMLSpanElement) {
            child.textContent = null
        }
    }

    const text = breadcrumbs.textContent!.trim()

    if (changeType === 'added') {
        return { head: text }
    }
    if (changeType === 'removed') {
        return { base: text }
    }
    if (changeType === 'modified') {
        return { base: text, head: text }
    }

    const SEPARATOR = '→'

    const [start, changed, end] = text.split(/[{}]/).map(text => text.trim())
    const [first, second] = changed.split(SEPARATOR).map(text => text.trim())
    return { base: clean(start + first + end), head: clean(start + second + end) }
}

export function determinePRChangeType(codeView: HTMLElement): ChangeType {
    // Moves and renames are rendered the same way.
    const spans = [...codeView.querySelectorAll<HTMLElement>('[data-qa="bk-file__header"] span')]
    for (const span of spans) {
        for (const changeType of changeTypes) {
            if (span.textContent?.toLowerCase().includes(changeType)) {
                return changeType
            }
        }
    }
    // Default change type
    return 'modified'
}

export function determineChangeTypeForCommit(codeView: HTMLElement): ChangeType {
    const changeTypeLozenge = codeView.querySelector<HTMLElement>('.diff-entry-lozenge')
    for (const changeType of changeTypes) {
        if (changeTypeLozenge?.textContent?.toLowerCase().includes(changeType)) {
            return changeType
        }
    }

    // Default change type
    return 'modified'
}

export function isPullRequestView({ pathname }: Pick<Location, 'pathname'>): boolean {
    return /\/pull-requests\/\d+/.test(pathname)
}

export function getRevisionsForPR(): { baseRevision: string; headRevision: string } {
    // Read from UI badges
    const branchesContainer = document.querySelector<HTMLElement>('[data-qa="pr-branches-and-state-styles"]')
    if (!branchesContainer) {
        throw new Error('Could not find branches element')
    }
    const branches = [...branchesContainer.querySelectorAll<HTMLElement>('[aria-hidden="true"]')]
    if (branches.length !== 2) {
        throw new Error('Expected source and destination branch elements')
    }
    // "Head" is the "source" branch.
    const headRevision = branches[0].textContent ?? ''
    // "Base" is the "destination" branch.
    const baseRevision = branches[1].textContent ?? ''

    return { baseRevision, headRevision }
}

/**
 * For single file code view
 */
export function getCommitIDFromPermalink(): string | null {
    const permalinkSelectors = ['a[type="button"]', 'a']

    // Try the narrower selector first, broaden if necessary
    for (const selector of permalinkSelectors) {
        const anchors = document.querySelectorAll<HTMLAnchorElement>(selector)

        for (const anchor of anchors) {
            const matches = anchor.href.match(/full-commit\/([\da-f]{40})\//)
            if (!matches) {
                continue
            }
            return matches[1]
        }
    }

    return null
}

export function getCommitIDsForCommit(): { baseCommitID: string; headCommitID: string } {
    const baseCommitMatch = document
        .querySelector<HTMLAnchorElement>('[title="Parent commit"]')
        ?.href.match(/\/commits\/(.*)/)
    if (!baseCommitMatch) {
        throw new Error('Could not determine base commit ID')
    }
    const baseCommitID = baseCommitMatch[1]

    const headCommitMatch = window.location.href.match(/\/commits\/(.*)/)
    if (!headCommitMatch) {
        throw new Error('Could not determine head commit ID')
    }
    const headCommitID = headCommitMatch[1]

    return { baseCommitID, headCommitID }
}
