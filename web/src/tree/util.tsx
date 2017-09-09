export function getParentDir(path: string): string {
    const split = path.split('/')
    if (split.length === 1) {
        return ''
    }
    return split.splice(0, split.length - 1).join('/')
}

export function scrollIntoView(el: Element, scrollRoot: Element): void {
    if (!scrollRoot.getBoundingClientRect) {
        return el.scrollIntoView()
    }

    const rootRect = scrollRoot.getBoundingClientRect()
    const elRect = el.getBoundingClientRect()

    const elAbove = elRect.top <= rootRect.top
    const elBelow = elRect.bottom >= rootRect.bottom

    if (elAbove) {
        el.scrollIntoView(true)
    } else if (elBelow) {
        el.scrollIntoView(false)
    }
}
