import { AbsoluteRepoFile } from '../../shared/repo'
import { MaybeDiffSpec } from '../../shared/repo'
import { getlineNumberForCell } from './util'

export enum PhabricatorMode {
    Diffusion = 1,
    Differential,
    Revision,
    Change,
}

export interface DiffusionState extends AbsoluteRepoFile {
    mode: PhabricatorMode
}

export interface DifferentialState {
    mode: PhabricatorMode
    differentialID: number
    leftDiffID?: number
    diffID?: number
    baseRev: string
    baseRepoPath: string
    headRev: string
    headRepoPath: string
}

/**
 * Refers to a URL like http://phabricator.aws.sgdev.org/differential/changeset/?ref=10%2F4&whitespace=ignore-most,
 * which a user gets to from a differential page by clicking "View Standalone" on a single file.
 * It is a diff view for a single file.
 * TODO(john): this is unimplemented, I'm not sure how.
 */
export interface ChangesetState {
    mode: PhabricatorMode
    differentialID: number
    leftDiffID?: number
    diffID: number
    repoPath: number
    filePath: number
}

export interface RevisionState {
    mode: PhabricatorMode
    repoPath: string
    baseCommitID: string
    headCommitID: string
}

/**
 * Refers to a URL like http://phabricator.aws.sgdev.org/source/nzap/change/master/checked_message_bench_test.go,
 * which a user gets to by clicking "Show Last Change" on a differential page.
 */
export interface ChangeState {
    mode: PhabricatorMode
    repoPath: string
    filePath: string
    commitID: string
}

export function convertSpacesToTabs(realLineContent: string, domContent: string): boolean {
    return !!realLineContent && !!domContent && realLineContent.startsWith('\t') && !domContent.startsWith('\t')
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
    blobContentLines: string[]
): (target: HTMLElement, opt: MaybeDiffSpec) => { line: number; character: number; word: string } | undefined {
    return (target: HTMLElement, opt: MaybeDiffSpec) => {
        let isDelta = opt.isDelta

        const origTarget = target
        while (target && target.tagName !== 'TD' && target.tagName !== 'TR' && target.tagName !== 'BODY') {
            // Find ancestor which wraps the whole line of code, not just the target token.
            target = target.parentNode as HTMLElement
        }
        if (!target || target.tagName !== 'TD') {
            // Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
            return undefined
        }
        let line = 0
        if (!isDelta) {
            let innerText = target.previousElementSibling!.textContent!
            if (!innerText && target.previousElementSibling!.firstElementChild) {
                innerText = getComputedStyle(
                    target.previousElementSibling!.firstElementChild!,
                    ':before'
                ).getPropertyValue('content')
            }
            line = parseInt(innerText.replace(`"`, ''), 10)
        } else {
            let lineEl = target.previousElementSibling
            let seenFirstHeader = false
            while (lineEl) {
                if (lineEl.tagName === 'TH') {
                    if (opt.isBase && !opt.isSplitDiff && !seenFirstHeader) {
                        // for unified diff, skip over to get the leftmost column
                        seenFirstHeader = true
                        continue
                    }
                    const lineNumber = getlineNumberForCell(lineEl)
                    if (lineNumber) {
                        line = lineNumber
                        break
                    }
                }
                lineEl = lineEl.previousElementSibling
            }
        }

        if (!line) {
            return undefined
        }

        const realLineContent = blobContentLines[line - 1]
        const convertSpaces = convertSpacesToTabs(realLineContent, target.textContent!.substr(isDelta ? 1 : 0))

        isDelta = true && !opt.isBase && !(opt.isDelta && !opt.isSplitDiff) // strip first character for diffusion and right side of split diff only
        let seenNonWhitespace = false

        let character = 1
        // Iterate recursively over the current target's children until we find the original target;
        // count characters along the way. Return true if the original target is found.
        function findOrigTarget(root: HTMLElement): boolean {
            // tslint:disable-next-line:prefer-for-of
            for (let i = 0; i < root.childNodes.length; ++i) {
                const child = root.childNodes[i] as HTMLElement
                if (child.id === 'scroll_target') {
                    // Scroll target appears on lines like http://phabricator.aws.sgdev.org/source/nzap/browse/master/keyvalue.go;9bee3bc2cd3068dd97dfa87068c4431c5d6093ef$29#L29:10
                    continue
                }
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
                        character -= 1 // make sure not to count weird diff prefix
                        if (convertSpaces) {
                            character -= spacesToTabsAdjustment(child.textContent!.substr(1))
                        }
                        seenNonWhitespace = !isOnlyWhitespace(child.textContent!.substr(1))
                        isDelta = false
                    } else {
                        if (!seenNonWhitespace && convertSpaces) {
                            character -= spacesToTabsAdjustment(child.textContent!)
                            seenNonWhitespace = !isOnlyWhitespace(child.textContent!)
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
}

export function spacesToTabsAdjustment(text: string): number {
    let suffix = text
    let adjustment = 0

    while (suffix.length >= 2 && suffix.startsWith('  ')) {
        ++adjustment
        suffix = suffix.substr(2)
    }
    return adjustment
}

const onlyWhitespaceRegex = /^\s+$/

function isOnlyWhitespace(text: string): boolean {
    if (text.length === 0) {
        return true
    }
    return onlyWhitespaceRegex.test(text)
}

/**
 * Returns the <span> (descendent of a <td> containing code) which contains text beginning
 * at the specified character offset (1-indexed)
 * @param cell the <td> containing syntax highlighted code
 * @param offset character offset
 */
export function findElementWithOffset(
    blobContentLines: string[]
): (cell: HTMLElement, line: number, offset: number, opt: MaybeDiffSpec) => HTMLElement | undefined {
    return (cell: HTMLElement, line: number, offset: number, opt: MaybeDiffSpec) => {
        let ignoreFirstCharacter = true && !opt.isBase // for some reason, base doesn't have leading character
        let seenNonWhitespace = false
        const realLineContent = blobContentLines[line - 1]
        const convertSpaces = convertSpacesToTabs(
            realLineContent,
            cell.textContent!.substr(ignoreFirstCharacter ? 1 : 0)
        )

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
                        let adjustment = 0
                        if (!seenNonWhitespace && convertSpaces) {
                            // TODO(john): this requires the first ignore character to be on the same span as whitespace
                            // not sure if that's ok
                            if (doIgnore) {
                                adjustment = spacesToTabsAdjustment(child.textContent!.substr(1))
                            } else {
                                adjustment = spacesToTabsAdjustment(child.textContent!)
                            }
                        }
                        if (!seenNonWhitespace) {
                            seenNonWhitespace = doIgnore
                                ? !isOnlyWhitespace(child.textContent!.substr(1))
                                : isOnlyWhitespace(child.textContent!)
                        }
                        if (currOffset + (child.textContent!.length - adjustment) >= offset) {
                            return currNode
                        }
                        currOffset += child.textContent!.length - adjustment
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
}
