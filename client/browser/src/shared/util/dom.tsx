import { Observable } from 'rxjs'

/**
 * commitIDFromPermalink finds the permalink element on the page and extracts
 * the 40 character commit ID from it. This will throw if the link doesn't exist
 * or doesn't match the provided regex.
 */
export function commitIDFromPermalink({ selector, hrefRegex }: { selector: string; hrefRegex: RegExp }): string {
    const permalinkElement = document.querySelector<HTMLAnchorElement>(selector)
    if (!permalinkElement) {
        throw new Error(
            `Unable to determine the commit ID (40 character hash) you're on because the permalink shortcut element (query selector ${selector}) which contains the commit ID does not exist in the DOM.`
        )
    }
    const href = permalinkElement.getAttribute('href')
    if (!href) {
        throw new Error(
            `Unable to determine the commit ID (40 character hash) you're on because the permalink shortcut element (query selector ${selector}) which contains the commit ID does not have an href attribute.`
        )
    }
    const commitIDMatch = hrefRegex.exec(href)
    if (!commitIDMatch || !commitIDMatch[1]) {
        throw new Error(
            `Unable to determine the commit ID (40 character hash) you're on because the permalink shortcut element's (query selector ${selector}) href is ${href}, which doesn't match the regex /${hrefRegex}/.`
        )
    }
    return commitIDMatch[1]
}

/**
 * Compatible with MutationRecord, but synthesizable.
 */
export interface MutationRecordLike {
    addedNodes: Iterable<Node>
    removedNodes: Iterable<Node>
}

/**
 * An Observable wrapper around `MutationObserver`.
 *
 * Instructs the user agent to observe a given `target` (a node) and report any mutations based on
 * the criteria given by `options` (an object).
 *
 * The `options` argument allows for setting mutation
 * observation options via object members. These are the object members that
 * can be used:
 * - `childList`
 *   Set to true if mutations to target's children are to be observed.
 * - `attributes`
 *   Set to true if mutations to target's attributes are to be observed. Can be omitted if attributeOldValue or attributeFilter is
 *   specified.
 * - `characterData`
 *   Set to true if mutations to target's data are to be observed. Can be omitted if characterDataOldValue is specified.
 * - `subtree`
 *   Set to true if mutations to not just target, but
 *   also target's descendants are to be
 *   observed.
 * - `attributeOldValue`
 *   Set to true if attributes is true or omitted
 *   and target's attribute value before the mutation
 *   needs to be recorded.
 * - `characterDataOldValue`
 *   Set to true if characterData is set to true or omitted and target's data before the mutation
 *   needs to be recorded.
 * - `attributeFilter`
 *   Set to a list of attribute local names (without namespace) if not all attribute mutations need to be
 *   observed and attributes is true
 *   or omitted.
 */
export const observeMutations = (target: Node, options?: MutationObserverInit): Observable<MutationRecord[]> =>
    new Observable<MutationRecord[]>(subscriber => {
        const mutationObserver = new MutationObserver(mutations => {
            subscriber.next(mutations)
        })
        mutationObserver.observe(target, options)
        return () => mutationObserver.disconnect()
    })

/**
 * Like `element.querySelectorAll()`, but will return (only) the element itself if it matches the selector.
 */
export function querySelectorAllOrSelf<K extends keyof HTMLElementTagNameMap>(
    element: Element,
    selectors: K
): Iterable<HTMLElementTagNameMap[K]>
export function querySelectorAllOrSelf<K extends keyof SVGElementTagNameMap>(
    element: Element,
    selectors: K
): Iterable<SVGElementTagNameMap[K]>
export function querySelectorAllOrSelf(element: Element, selectors: string): Iterable<Element>
export function querySelectorAllOrSelf(element: Element, selectors: string): Iterable<Element> {
    return element.matches(selectors) ? [element] : element.querySelectorAll(selectors)
}
