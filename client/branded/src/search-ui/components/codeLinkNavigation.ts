import type { MouseEvent, KeyboardEvent } from 'react'
import type React from 'react'

import type { NavigateFunction } from 'react-router-dom'

/**
 * A helper function to replicate browser behavior when clicking on links.
 * A very common interaction is to open links in a new in the _background_ via
 * CTRL/CMD + click or middle click.
 * Unfortunately `window.open` doesn't give us much control over how the new
 * window/tab should be opened, and the behavior is inconcistent between
 * browsers.
 * In order to replicate the standard behvior as much as possible this function
 * dynamically creates an `<a>` element and triggers a click event on it.
 */
function openLinkInNewTab(
    url: string,
    event: Pick<MouseEvent, 'ctrlKey' | 'altKey' | 'shiftKey' | 'metaKey'>,
    button: 'primary' | 'middle'
): void {
    const link = document.createElement('a')
    link.href = url
    link.style.display = 'none'
    link.target = '_blank'
    link.rel = 'noopener noreferrer'
    const clickEvent = new window.MouseEvent('click', {
        bubbles: false,
        altKey: event.altKey,
        shiftKey: event.shiftKey,
        // Regarding middle click: Setting "button: 1:" doesn't seem to suffice:
        // Firefox doesn't react to the event at all, Chromium opens the tab in
        // the foreground. So in order to simulate a middle click, we set
        // ctrlKey and metaKey to `true` instead.
        ctrlKey: button === 'middle' ? true : event.ctrlKey,
        metaKey: button === 'middle' ? true : event.metaKey,
        view: window,
    })

    // It looks the link has to be part of the document, otherwise Firefox won't
    // trigger the default behavior (it works without appending in Chromium).
    document.body.append(link)
    link.dispatchEvent(clickEvent)
    link.remove()
}

/**
 * Since we are not using a real link anymore, we have to simulate opening
 * the file in a new tab when the search result is clicked on with the
 * middle mouse button.
 * This handler is bound to the `mouseup` event because the `auxclick`
 * (https://w3c.github.io/uievents/#event-type-auxclick) event is not
 * support by all browsers yet (https://caniuse.com/?search=auxclick)
 */
export function navigateToFileOnMiddleMouseButtonClick(event: React.MouseEvent<HTMLElement>): void {
    const href = event.currentTarget.getAttribute('data-href')
    if (href && event.button === 1) {
        openLinkInNewTab(href, event, 'middle')
    }
}

/**
 * This helper function determines whether a mouse/click event was triggered as
 * a result of selecting text in search results.
 * There are at least to ways to do this:
 *
 * - Tracking `mouseup`, `mousemove` and `mousedown` events. The occurrence of
 * a `mousemove` event would indicate a text selection. However, users
 * might slightly move the mouse while clicking, and solutions that would
 * take this into account seem fragile.
 * - (implemented here) Inspect the Selection object returned by
 * `window.getSelection()`.
 *
 * CAVEAT: Chromium and Firefox (and maybe other browsers) behave
 * differently when a search result is clicked *after* text selection was
 * made:
 *
 * - Firefox will clear the selection before executing the click event
 * handler, i.e. the search result will be opened.
 * - Chrome will only clear the selection if the click happens *outside*
 * of the selected text (in which case the search result will be
 * opened). If the click happens inside the selected text the selection
 * will be cleared only *after* executing the click event handler.
 */
function isTextSelectionEvent(event: MouseEvent<HTMLElement>): boolean {
    const selection = window.getSelection()

    // Text selections are always ranges. Should the type not be set, verify
    // that the selection is not empty.
    if (selection && (selection.type === 'Range' || selection.toString() !== '')) {
        // Firefox specific: Because our code excerpts are implemented as tables,
        // CTRL+click would select the table cell. Since users don't know that we
        // use tables, the most likely wanted to open the search results in a new
        // tab instead though.
        if ((event.ctrlKey || event.metaKey) && selection.anchorNode?.nodeName === 'TR') {
            // Ugly side effect: We don't want the table cell to be highlighted.
            // The focus style that Firefox uses doesn't seem to be affected by
            // CSS so instead we clear the selection.
            selection.empty()
            return false
        }

        return true
    }

    return false
}

/**
 * This handler implements the logic to simulate the click/keyboard
 * activation behavior of links, while also allowing the selection of text
 * inside the element.
 * Because a click event is dispatched in both cases (clicking the search
 * result to open it as well as selecting text within it), we have to be
 * able to distinguish between those two actions.
 * If we detect a text selection action, we don't have to do anything.
 *
 * CAVEATS:
 * - In Firefox, Shift+click will open the URL in a new tab instead of
 * a window (unlike Chromium which seems to show the same behavior as with
 * native links).
 * - Firefox will insert \t\n in between table rows, causing the copied
 * text to be different from what is in the file/search result.
 */
function onClickCodeExcerptHref(
    event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>,
    onClickHref: (href: string) => void
): void {
    // Testing for text selection is only necessary for mouse/click
    // events. Middle-click (event.button === 1) is already handled in the `onMouseUp` callback.
    if (
        (event.type === 'click' &&
            !isTextSelectionEvent(event as MouseEvent<HTMLElement>) &&
            (event as MouseEvent<HTMLElement>).button !== 1) ||
        (event as KeyboardEvent<HTMLElement>).key === 'Enter'
    ) {
        const href = event.currentTarget.getAttribute('data-href')
        if (!event.defaultPrevented && href) {
            event.preventDefault()
            onClickHref(href)
        }
    }
}

export function navigateToCodeExcerpt(
    event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>,
    openInNewTab: boolean,
    navigate: NavigateFunction
): void {
    onClickCodeExcerptHref(event, href => {
        if (openInNewTab || event.ctrlKey || event.metaKey || event.shiftKey) {
            openLinkInNewTab(href, event, 'primary')
        } else {
            navigate(href)
        }
    })
}
