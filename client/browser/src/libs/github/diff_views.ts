import { Observable } from 'rxjs'
import { distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators'
import { observeMutations } from '../../shared/util/dom'

export function observeDiffViewVisibleRanges(elem: HTMLElement): Observable<boolean> {
    return observeMutations(elem, { attributes: true }).pipe(
        switchMap(mutations => mutations),
        filter(
            mutation =>
                mutation.type === 'attributes' &&
                mutation.attributeName === 'class' &&
                mutation.attributeNamespace === null
        ),
        map(() => getDiffViewCollapsed(elem)),
        distinctUntilChanged()
    )
}

export function setDiffViewCollapsed(elem: HTMLElement, collapsed: boolean): void {
    const toggleButton = elem.querySelector<HTMLElement>('button[aria-label="Toggle diff contents"]')
    if (!toggleButton) {
        throw new Error('Unable to set diff view visible range (no toggle button found)')
    }

    // const isCollapsed = toggleButton.getAttribute('aria-expanded') !== 'true'
    if (getDiffViewCollapsed(elem) !== collapsed) {
        // Only toggle if the want state != current state.
        toggleButton.click()
    }
}

function getDiffViewCollapsed(elem: HTMLElement): boolean {
    // On GitHub: `.file.open` means open (non-collapsed), `.file:not(.open)` means
    // collapsed.
    return !elem.classList.contains('open')
}
