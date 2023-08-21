import { useCallback, useLayoutEffect, useState } from 'react'

type ClientRect = Record<keyof Omit<DOMRect, 'toJSON'>, number>

function roundValues(_rect: ClientRect): ClientRect {
    const rect: ClientRect = {
        ..._rect,
    }
    for (const key of Object.keys(rect) as any as keyof ClientRect) {
        rect[key as any as keyof ClientRect] = Math.round(rect[key as any as keyof ClientRect])
    }
    return rect
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
}

const defaultState: TextSelectionState = {
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
    const [{ clientRect, isCollapsed, textContent }, setState] = useState<TextSelectionState>(defaultState)

    const handler = useCallback(() => {
        let newRect: ClientRect
        const selection = window.getSelection()
        const newState: TextSelectionState = {}

        if (selection === null || !selection.rangeCount) {
            setState(newState)
            return
        }

        const range = selection.getRangeAt(0)

        if (target !== null && !target?.contains(range.commonAncestorContainer)) {
            setState(newState)
            return
        }

        if (range === null) {
            setState(newState)
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
    }
}
