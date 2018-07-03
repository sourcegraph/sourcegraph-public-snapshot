/**
 * Returned when only the line is known.
 *
 * 1-indexed
 */
interface Line {
    line: number
}

export interface HoveredToken {
    /** 1-indexed */
    line: number
    /** 1-indexed */
    character: number
    word: string
    part?: 'old' | 'new'
}

/**
 * Determines the line and character offset for some source code, identified by its HTMLElement wrapper.
 * It works by traversing the DOM until the HTMLElement's TD ancestor. Once the ancestor is found, we traverse the DOM again
 * (this time the opposite direction) counting characters until the original target is found.
 * Returns undefined if line/char cannot be determined for the provided target.
 * @param target The element to compute line & character offset for.
 * @param ignoreFirstChar Whether to ignore the first character on a line when computing character offset.
 */
export function locateTarget(
    target: HTMLElement,
    boundary: HTMLElement,
    ignoreFirstChar = false
): Line | HoveredToken | undefined {
    const origTarget = target
    while (target && target.tagName !== 'TD' && target.tagName !== 'BODY' && target !== boundary) {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = target.parentNode as HTMLElement
    }
    if (!target || target.tagName !== 'TD' || target === boundary) {
        // Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
        return undefined
    }

    let lineElement: HTMLElement
    if (target.classList.contains('line')) {
        lineElement = target
    } else if (target.previousElementSibling && (target.previousElementSibling as HTMLElement).dataset.line) {
        lineElement = target.previousElementSibling as HTMLTableDataCellElement
    } else if (
        target.previousElementSibling &&
        target.previousElementSibling.previousElementSibling &&
        (target.previousElementSibling.previousElementSibling as HTMLElement).dataset.line
    ) {
        lineElement = target.previousElementSibling.previousElementSibling as HTMLTableDataCellElement
    } else {
        lineElement = target.parentElement as HTMLTableRowElement
    }
    if (!lineElement || !lineElement.dataset.line) {
        return undefined
    }
    const line = parseInt(lineElement.dataset.line!, 10)
    const part = lineElement.dataset.part as 'old' | 'new' | undefined

    let character = 1
    // Iterate recursively over the current target's children until we find the original target;
    // count characters along the way. Return true if the original target is found.
    function findOrigTarget(root: HTMLElement): boolean {
        // tslint:disable-next-line
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
                if (ignoreFirstChar) {
                    character -= 1 // make sure not to count weird diff prefix
                    ignoreFirstChar = false
                }
            }
        }
        return false
    }
    // Start recursion.
    if (findOrigTarget(target)) {
        return { line, character, word: origTarget.innerText, part }
    }
    return { line }
}
