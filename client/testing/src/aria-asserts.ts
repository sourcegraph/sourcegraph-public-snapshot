import { expect } from '@jest/globals'

// Since jest doesn't provide native matcher to check aria state of the element
// see https://github.com/testing-library/jest-dom/issues/144 for more details.
// We have to use our in-house assert utility for this. We can't use custom jest
// matcher `.toBeAriaDisabled` due to problems with TS global types problems,
// see https://github.com/sourcegraph/sourcegraph/pull/44461

/**
 * Checks element aria-disabled and disabled state. Fails if element is disabled.
 */
export function assertAriaEnabled(element: HTMLElement): void {
    expect(element.getAttribute('aria-disabled') !== 'true' && element.getAttribute('disabled') === null).toBe(true)
}

/**
 * Checks element aria-disabled and disabled state. Fails if element is active.
 */
export function assertAriaDisabled(element: HTMLElement): void {
    const nativeDisabled = (element as HTMLButtonElement).disabled
    const ariaDisabled = element.getAttribute('aria-disabled') === 'true'

    expect(ariaDisabled || nativeDisabled).toBe(true)
}
