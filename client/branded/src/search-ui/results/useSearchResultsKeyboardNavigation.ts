import { useCallback, useEffect, useState } from 'react'

import { isMacPlatform } from '@sourcegraph/common'

export function useSearchResultsKeyboardNavigation(
    root: HTMLElement | null,
    enableKeyboardNavigation: boolean | undefined
): [boolean, (isVisible: boolean, index: number) => void] {
    const [showFocusInputMessage, setShowFocusInputMessage] = useState(false)
    const [focusInputMessageTimeoutId, setFocusInputMessageTimeoutId] = useState<NodeJS.Timeout>()

    useEffect(() => {
        if (!root) {
            return
        }

        const onKeyDown = (event: KeyboardEvent): void => {
            if (!enableKeyboardNavigation) {
                return
            }
            // Do not run keyboard shortcuts if any modifier key is pressed to avoid hijacking default browser actions.
            if (event.ctrlKey || event.metaKey || event.altKey || event.shiftKey) {
                return
            }

            const selectableResults = Array.from(
                root.querySelectorAll<HTMLElement>('[data-selectable-search-result="true"]')
            )
            const selectedResult = getSelectedResult(root)

            switch (event.key) {
                case 'ArrowUp':
                case 'ArrowDown':
                case 'j':
                case 'k': {
                    const direction = event.key === 'ArrowUp' || event.key === 'k' ? 'up' : 'down'
                    const hasSelectedNextResult = selectNextResult(selectableResults, selectedResult, direction)
                    if (hasSelectedNextResult) {
                        event.stopPropagation()
                        event.preventDefault()
                    }
                    break
                }
                case 'ArrowLeft':
                case 'h': {
                    if (selectedResult) {
                        selectedResult.dispatchEvent(
                            new CustomEvent('collapseSearchResultsGroup', { bubbles: true, cancelable: true })
                        )
                        event.preventDefault()
                    }
                    break
                }
                case 'ArrowRight':
                case 'l': {
                    if (selectedResult) {
                        selectedResult.dispatchEvent(
                            new CustomEvent('expandSearchResultsGroup', { bubbles: true, cancelable: true })
                        )
                        event.preventDefault()
                    }
                    break
                }
                case '/': {
                    setShowFocusInputMessage(false)
                    if (focusInputMessageTimeoutId) {
                        clearTimeout(focusInputMessageTimeoutId)
                    }
                    break
                }
                default: {
                    // If the user tries typing alphanumeric characters while focused on the search results, we
                    // show him a message on how to focus the search input.
                    if (
                        event.key.length === 1 &&
                        /[\dA-Za-z]/.test(event.key) &&
                        !event.altKey &&
                        !event.ctrlKey &&
                        !event.shiftKey &&
                        !event.metaKey
                    ) {
                        setShowFocusInputMessage(true)
                        if (focusInputMessageTimeoutId) {
                            clearTimeout(focusInputMessageTimeoutId)
                        }
                        const timeoutId = setTimeout(() => setShowFocusInputMessage(false), 3000)
                        setFocusInputMessageTimeoutId(timeoutId)
                    }
                    break
                }
            }
        }

        // When a user collapses a group of results (e.g., file matches) and hides the selected
        // result we select the last visible result in the group instead.
        const onCollapse = (): void => {
            // Get the currently selected results group.
            const group = getSelectedResult(root)?.closest('[data-selectable-search-results-group="true"]')
            if (!group) {
                return
            }

            // Hack: use setTimeout to find out what results are visible after the collapsed group is rendered.
            setTimeout(() => {
                // Exit early if there is still a result selected after the group is collapsed.
                if (getSelectedResult(root)) {
                    return
                }
                // Otherwise, find the last visible result in the group and select it.
                const groupSelectables = group.querySelectorAll<HTMLElement>('[data-selectable-search-result="true"]')
                // eslint-disable-next-line unicorn/prefer-at
                selectElement(groupSelectables[groupSelectables.length - 1])
            }, 0)
        }

        const onDocumentKeyDown = (event: KeyboardEvent): void => {
            if (!enableKeyboardNavigation) {
                return
            }

            if ((event.key === 'ArrowDown' || event.key === 'j') && isMetaKey(event, isMacPlatform())) {
                selectFirstResult(root)
                event.preventDefault()
            }
        }

        document.addEventListener('keydown', onDocumentKeyDown)
        root.addEventListener('keydown', onKeyDown)
        root.addEventListener('collapseSearchResultsGroup', onCollapse)
        root.addEventListener('toggleSearchResultsGroup', onCollapse)
        return () => {
            document.removeEventListener('keydown', onDocumentKeyDown)
            root.removeEventListener('keydown', onKeyDown)
            root.removeEventListener('collapseSearchResultsGroup', onCollapse)
            root.removeEventListener('toggleSearchResultsGroup', onCollapse)
        }
    }, [
        root,
        enableKeyboardNavigation,
        focusInputMessageTimeoutId,
        setFocusInputMessageTimeoutId,
        setShowFocusInputMessage,
    ])

    const onVisibilityChange = useCallback(
        (isVisible: boolean, index: number) => {
            if (index === 0 && isVisible && root && enableKeyboardNavigation) {
                selectFirstResult(root)
            }
        },
        [root, enableKeyboardNavigation]
    )

    return [showFocusInputMessage, onVisibilityChange]
}

function selectFirstResult(root: HTMLElement): void {
    selectElement(root.querySelector<HTMLElement>('[data-selectable-search-result="true"]'))
}

function selectNextResult(
    selectableResults: HTMLElement[],
    selectedResult: HTMLElement | null,
    direction: 'up' | 'down'
): boolean {
    if (!selectedResult || selectableResults.length === 0) {
        return false
    }

    const currentIndex = selectableResults.findIndex(selectable => selectable.isEqualNode(selectedResult))
    const nextIndex = direction === 'down' ? currentIndex + 1 : currentIndex - 1
    const nextSelected = nextIndex >= 0 && nextIndex < selectableResults.length ? selectableResults[nextIndex] : null

    return selectElement(nextSelected)
}

function selectElement(resultElement: HTMLElement | undefined | null): boolean {
    if (!resultElement) {
        return false
    }
    resultElement?.focus()
    resultElement?.scrollIntoView({ behavior: 'auto', block: 'nearest' })
    return true
}

// The selected element is the currently focused element (activeElement) with
// the `data-selectable-search-result=true` attribute, and is contained in the root element subtree.
function getSelectedResult(root: HTMLElement): HTMLElement | null {
    return document.activeElement &&
        document.activeElement instanceof HTMLElement &&
        document.activeElement.dataset.selectableSearchResult === 'true' &&
        root.contains(document.activeElement)
        ? document.activeElement
        : null
}

function isMetaKey(event: KeyboardEvent, isMacPlatform: boolean): boolean {
    return (isMacPlatform && event.metaKey) || (!isMacPlatform && event.ctrlKey)
}
