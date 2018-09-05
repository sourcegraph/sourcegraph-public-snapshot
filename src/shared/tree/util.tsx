export function getParentDir(path: string): string {
    const split = path.split('/')
    if (split.length === 1) {
        return ''
    }
    return split.splice(0, split.length - 1).join('/')
}

/**
 * Returns true iff path is equal to candidate OR if candidate is an ancestor
 * directory of path.
 */
export function isEqualOrAncestor(path: string, candidate: string): boolean {
    return path === candidate || candidate === '' || path.startsWith(candidate + '/')
}

export function scrollIntoView(el: Element, scrollRoot: Element): void {
    if (!scrollRoot.getBoundingClientRect) {
        return el.scrollIntoView()
    }

    const rootRect = scrollRoot.getBoundingClientRect()
    const elRect = el.getBoundingClientRect()

    const elAbove = elRect.top <= rootRect.top + 30
    const elBelow = elRect.bottom >= rootRect.bottom

    if (elAbove) {
        el.scrollIntoView(true)
    } else if (elBelow) {
        el.scrollIntoView(false)
    }
}
