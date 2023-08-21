import { useCallback, useLayoutEffect, useState } from 'react'

import { mapValues } from 'lodash'

type ClientRect = Omit<DOMRect, 'toJSON'>

function roundValues(rect: ClientRect): ClientRect {
    return mapValues(rect, Math.round)
}

function shallowDiff(prev: any = {}, next: any = {}) {
    if (prev !== null && next !== null) {
        for (const key of Object.keys(next)) {
            if (prev[key] !== next[key]) {
                return true
            }
        }
    } else if (prev !== next) {
        return true
    }
    return false
}

interface TextSelectionState {
    clientRect?: ClientRect
    isCollapsed?: boolean
    textContent?: string
    selection?: Selection
}

const DEFAULT_STATE: TextSelectionState = {
    clientRect: {
        height: 0,
        width: 0,
        x: 0,
        y: 0,
        bottom: 0,
        left: 0,
        right: 0,
        top: 0,
    },
    isCollapsed: true,
    textContent: '',
}

export function useTextSelection(target?: HTMLElement): TextSelectionState {
    const [{ clientRect, isCollapsed, textContent, selection }, setState] = useState<TextSelectionState>(DEFAULT_STATE)

    const handler = useCallback(() => {
        let newRect: ClientRect
        const selection = window.getSelection()
        const newState: TextSelectionState = {}

        if (!selection?.rangeCount) {
            setState(newState)
            return
        }

        const range = selection.getRangeAt(0)

        if (!target?.contains(range.commonAncestorContainer)) {
            setState(newState)
            return
        }

        if (!range) {
            setState(newState)
            return
        }

        // Restrict popover to only code content.
        // Hack because Cody's dangerouslySetInnerHTML forces us to use a ref on code block's wrapper text
        if (selection?.getRangeAt(0)?.commonAncestorContainer?.nodeName !== 'CODE') {
            return
        }

        const contents = range.cloneContents()

        if (!!contents.textContent) {
            newState.textContent = contents.textContent
        }

        const rects = range.getClientRects()

        if (rects.length === 0 && range.commonAncestorContainer !== null) {
            const element = range.commonAncestorContainer as HTMLElement
            newRect = roundValues(element.getBoundingClientRect().toJSON())
        } else {
            if (rects.length < 1) {
                return
            }
            newRect = roundValues(rects[0].toJSON())
        }
        if (shallowDiff(clientRect, newRect)) {
            newState.clientRect = newRect
        }
        newState.isCollapsed = range.collapsed

        setState(newState)
    }, [target, clientRect])

    useLayoutEffect(() => {
        document.addEventListener('selectionchange', handler)
        document.addEventListener('keydown', handler)
        document.addEventListener('keyup', handler)
        window.addEventListener('resize', handler)

        return () => {
            document.removeEventListener('selectionchange', handler)
            document.removeEventListener('keydown', handler)
            document.removeEventListener('keyup', handler)
            window.removeEventListener('resize', handler)
        }
    }, [target, handler])

    return {
        clientRect,
        isCollapsed,
        textContent,
        selection,
    }
}
