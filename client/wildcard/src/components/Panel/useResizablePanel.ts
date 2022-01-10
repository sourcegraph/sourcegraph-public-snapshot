import React, { useEffect, useRef, useState } from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged } from 'rxjs/operators'

import { PANEL_POSITIONS } from './constants'

const STORAGE_KEY_PREFIX = 'ResizePanel:'

const savePanelSize = (storageKey: UseResizablePanelParameters['storageKey'], size: number): void => {
    if (storageKey) {
        localStorage.setItem(`${STORAGE_KEY_PREFIX}${storageKey}`, String(size))
    }
}

const getCachedPanelSize = (storageKey: string | undefined | null, defautlSize: number): number => {
    if (!storageKey) {
        return defautlSize
    }

    const cachedSize = localStorage.getItem(`${STORAGE_KEY_PREFIX}${storageKey}`)

    if (cachedSize !== null) {
        const sizeNumber = parseInt(cachedSize, 10)

        if (sizeNumber >= 0) {
            return sizeNumber
        }
    }

    return defautlSize
}

export interface PanelResizerState {
    panelSize: number
    isResizing: boolean
}

export interface UseResizablePanelParameters {
    position: typeof PANEL_POSITIONS[number]
    panelRef: React.MutableRefObject<HTMLDivElement | null>
    handleRef: React.MutableRefObject<HTMLDivElement | null>
    storageKey: string | undefined | null
    defaultSize: number
}

export const useResizablePanel = ({
    defaultSize,
    handleRef,
    panelRef,
    position,
    storageKey,
}: UseResizablePanelParameters): PanelResizerState => {
    const sizeUpdates = useRef(new Subject<number>())
    const subscriptions = useRef(new Subscription())

    const [isResizing, setResizing] = useState(false)
    const [panelSize, setPanelSize] = useState(defaultSize)

    useEffect(() => {
        setPanelSize(getCachedPanelSize(storageKey, defaultSize))
    }, [storageKey, defaultSize])

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

            setPanelSize(size)
            sizeUpdates.current.next(size)
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
    }, [panelRef, handleRef, position, storageKey])

    return { panelSize, isResizing }
}
