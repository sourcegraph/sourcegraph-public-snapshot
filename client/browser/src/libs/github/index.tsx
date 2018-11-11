import { MaybeDiffSpec, ParsedRepoURI } from '../../shared/repo'

export interface GitHubURL extends ParsedRepoURI {
    user?: string
    repoName?: string
    isDelta?: boolean
    isPullRequest?: boolean
    isCommit?: boolean
    isCodePage?: boolean
    isCompare?: boolean
}

export interface GitHubBlobUrl {
    mode: GitHubMode
    owner: string
    repoName: string
    revAndPath: string
    lineNumber: string | undefined
    rev: string
    filePath: string
}

export interface GitHubPullUrl {
    mode: GitHubMode
    owner: string
    repoName: string
    view: string
    rev: string
    id: number
    filePath?: string
}

export interface GitHubRepositoryUrl {
    mode: GitHubMode
    owner: string
    repoName: string
    rev?: string
    filePath?: string
}

export enum GitHubMode {
    Blob,
    Commit,
    PullRequest,
    Repository,
}

/**
 * getTargetLineAndOffset determines the line and character offset for some source code, identified by its HTMLElement wrapper.
 * It works by traversing the DOM until the HTMLElement's TD ancestor. Once the ancestor is found, we traverse the DOM again
 * (this time the opposite direction) counting characters until the original target is found.
 * Returns undefined if line/char cannot be determined for the provided target.
 * @param target The element to compute line & character offset for.
 * @param isDelta Whether to ignore the first character on a line when computing character offset.
 */
export function getTargetLineAndOffset(
    target: HTMLElement,
    opt: MaybeDiffSpec
): { line: number; character: number; word: string } | undefined {
    const origTarget = target
    let isDelta = opt.isDelta
    while (target && target.tagName !== 'TD' && target.tagName !== 'BODY') {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = target.parentNode as HTMLElement
    }
    if (!target || target.tagName !== 'TD') {
        // Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
        return undefined
    }
    if (!target.classList.contains('blob-code') || target.classList.contains('blob-code-hunk')) {
        // we've hovered over something else, like line number or line expander
        return undefined
    }
    let line: number
    if (target.classList.contains('blob-num')) {
        line = parseInt(target.getAttribute('data-line-number')!, 10)
    } else {
        if (opt.isSplitDiff) {
            const lineEl = target.previousElementSibling!
            line = parseInt(lineEl.getAttribute('data-line-number')!, 10) || parseInt(lineEl.textContent!, 10)
        } else {
            const lineEl = (isDelta
                ? opt.isBase
                    ? target.previousElementSibling!.previousElementSibling
                    : target.previousElementSibling
                : target.previousElementSibling) as HTMLElement
            line = parseInt(lineEl.getAttribute('data-line-number')!, 10) || parseInt(lineEl.textContent!, 10)
        }
    }

    if (!target) {
        return undefined
    }

    if (isDelta && !target.classList.contains('blob-code-inner')) {
        target = target.querySelector('.blob-code-inner') as HTMLElement
    }

    if (!target) {
        return undefined
    }

    let character = 1
    // Iterate recursively over the current target's children until we find the original target;
    // count characters along the way. Return true if the original target is found.
    function findOrigTarget(root: HTMLElement): boolean {
        // tslint:disable-next-line prefer-for-of
        for (let i = 0; i < root.childNodes.length; ++i) {
            const child = root.childNodes[i] as HTMLElement
            if (child === origTarget) {
                return true
            }
            if (child.children === undefined) {
                character += child.textContent!.length
                continue
            }
            if (child.children.length > 0 && findOrigTarget(child)) {
                // Walk over nested children, then short-circuit the loop to avoid double counting children.
                return true
            }
            if (child.children.length === 0) {
                // Child is not the original target, but has no chidren to recurse on. Add to character offset.
                character += (child.textContent as string).length // TODO(john): I think this needs to be escaped before we add its length...
                if (isDelta) {
                    isDelta = false
                    const blobInner = root.closest('.blob-code-inner') as HTMLElement
                    if (
                        !['deletion', 'context', 'addition'].some(name =>
                            blobInner.classList.contains('blob-code-marker-' + name)
                        ) &&
                        !isRefinedGitHubExtensionRemoveDiffSignsEnabled(blobInner)
                    ) {
                        character -= 1
                    }
                }
            }
        }
        return false
    }

    // Start recursion.
    if (findOrigTarget(target)) {
        return { line, character, word: origTarget.innerText }
    }
}

/**
 * If the Refined GitHub extension is installed, it changes the DOM of a commit/PR diff (as part of the extension's
 * "No Whitespace" functionality). It removes the diff marker ('+'/'-'/' '), which means we need to add 1 to the
 * character to compensate. Otherwise, hovers and definitions would all be off-by-one and would usually return no
 * results.
 */
function isRefinedGitHubExtensionRemoveDiffSignsEnabled(target: HTMLElement): boolean {
    try {
        const tr = target.parentElement!.parentElement!
        return tr.classList.contains('refined-github-diff-signs')
    } catch (err) {
        // noop
    }
    return false
}

/**
 * Returns the <span> (descendent of a <td> containing code) which contains text beginning
 * at the specified character offset (1-indexed)
 * @param cell the <td> containing syntax highlighted code
 * @param offset character offset
 */
export function findElementWithOffset(
    cell: HTMLElement,
    line: number,
    offset: number,
    opt: MaybeDiffSpec
): HTMLElement | undefined {
    let ignoreFirstCharacter = opt.isDelta
    let currOffset = 0
    const walkNode = (currNode: HTMLElement): HTMLElement | undefined => {
        const numChildNodes = currNode.childNodes.length
        for (let i = 0; i < numChildNodes; ++i) {
            const child = currNode.childNodes[i]
            switch (child.nodeType) {
                case Node.TEXT_NODE:
                    const doIgnore = ignoreFirstCharacter && currOffset === 0 && child.textContent!.length > 0
                    if (doIgnore) {
                        ignoreFirstCharacter = false
                        currOffset -= 1
                    }
                    if (currOffset + child.textContent!.length >= offset) {
                        return currNode
                    }
                    currOffset += child.textContent!.length
                    continue

                case Node.ELEMENT_NODE:
                    const found = walkNode(child as HTMLElement)
                    if (found) {
                        return found
                    }
                    continue
            }
        }
        return undefined
    }
    return walkNode(cell)
}
