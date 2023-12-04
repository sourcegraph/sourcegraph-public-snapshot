import type React from 'react'
import { useEffect, useLayoutEffect, useRef, useState } from 'react'

import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged } from 'rxjs/operators'

import type { PANEL_POSITIONS } from './constants'

const STORAGE_KEY_PREFIX = 'ResizePanel:'

const savePanelSize = (storageKey: UseResizablePanelParameters['storageKey'], size: number): void => {
    if (storageKey) {
        localStorage.setItem(`${STORAGE_KEY_PREFIX}${storageKey}`, String(size))
    }
}

const getCachedPanelSize = (
    storageKey: string | undefined | null,
    defautlSize: number,
    maxSize: number | undefined,
    minSize: number | undefined
): number => {
    if (!storageKey) {
        return defautlSize
    }

    const cachedSize = localStorage.getItem(`${STORAGE_KEY_PREFIX}${storageKey}`)

    if (cachedSize !== null) {
        const sizeNumber = parseInt(cachedSize, 10)

        if (sizeNumber >= 0) {
            if (
                sizeNumber >= 0 &&
                isLessThanOrEqualMax(sizeNumber, maxSize) &&
                isGreaterThanOrEqualMin(sizeNumber, minSize)
            ) {
                return sizeNumber
            }
        }
    }

    return defautlSize
}

const isLessThanOrEqualMax = (size: number, maxSize?: number): boolean => {
    if (!maxSize) {
        return true
    }
    return size <= maxSize
}

const isGreaterThanOrEqualMin = (size: number, minSize?: number): boolean => {
    if (!minSize) {
        return true
    }
    return size >= minSize
}

export interface PanelResizerState {
    panelSize: number
    isResizing: boolean
}

export interface UseResizablePanelParameters {
    /**
     * Where the panel is (which also determines the axis along which the panel can be resized).
     */
    position?: typeof PANEL_POSITIONS[number]

    /**
     * Persist and restore the size of the panel using this key.
     */
    storageKey?: string | null

    /**
     * The default size for the panel.
     */
    defaultSize?: number

    /**
     * The minimum size for the panel.
     */
    minSize?: number

    /**
     * The maximum size for the panel.
     */
    maxSize?: number

    panelRef: React.MutableRefObject<HTMLDivElement | null>

    handleRef: React.MutableRefObject<HTMLDivElement | null>

    /**
     * Callback when the size has changed
     */
    onResize?: (size: number) => void
}

export const useResizablePanel = ({
    defaultSize = 250,
    handleRef,
    panelRef,
    position,
    storageKey,
    maxSize,
    minSize,
    onResize,
}: UseResizablePanelParameters): PanelResizerState => {
    const sizeUpdates = useRef(new Subject<number>())
    const subscriptions = useRef(new Subscription())

    const [isResizing, setResizing] = useState(false)
    const [panelSize, setPanelSize] = useState(defaultSize)

    useLayoutEffect(() => {
        const size = getCachedPanelSize(storageKey, defaultSize, maxSize, minSize)
        onResize?.(size)
        setPanelSize(size)
    }, [storageKey, defaultSize, maxSize, minSize, onResize])

    useEffect(() => {
        const currentSubscriptions = subscriptions.current

        currentSubscriptions.add(
            sizeUpdates.current
                .pipe(distinctUntilChanged(), debounceTime(250))
                .subscribe(size => savePanelSize(storageKey, size))
        )

        return () => {
            currentSubscriptions.unsubscribe()
        }
    }, [storageKey])

    useEffect(() => {
        const currentHandle = handleRef.current
        const currentPanel = panelRef.current

        const onMouseMove = (event: MouseEvent): void => {
            event.preventDefault()

            let size =
                position !== 'bottom'
                    ? position === 'left'
                        ? event.pageX - currentPanel!.getBoundingClientRect().left
                        : currentPanel!.getBoundingClientRect().right - event.pageX
                    : currentPanel!.getBoundingClientRect().bottom - event.pageY

            if (event.shiftKey) {
                size = Math.ceil(size / 20) * 20
            }

            if (isLessThanOrEqualMax(size, maxSize) && isGreaterThanOrEqualMin(size, minSize)) {
                onResize?.(size)
                setPanelSize(size)
                sizeUpdates.current.next(size)
            }
        }

        const onMouseUp = (event: Event): void => {
            event.preventDefault()

            setResizing(false)
            document.removeEventListener('mouseup', onMouseUp)
            document.removeEventListener('mousemove', onMouseMove)
        }

        const onMouseDown = (event: MouseEvent): void => {
            event.preventDefault()

            setResizing(true)
            document.addEventListener('mousemove', onMouseMove)
            document.addEventListener('mouseup', onMouseUp)
        }

        currentHandle?.addEventListener('mousedown', onMouseDown)

        return () => {
            currentHandle?.removeEventListener('mousedown', onMouseDown)
        }
    }, [panelRef, handleRef, position, storageKey, maxSize, minSize, onResize])

    return { panelSize, isResizing }
}
