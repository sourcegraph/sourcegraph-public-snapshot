import React, { useRef } from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { PANEL_POSITIONS } from './constants'
import { useResizablePanel } from './useResizablePanel'
import { getDisplayStyle, getPositionStyle } from './utils'

import styles from './Panel.module.scss'

export interface PanelProps {
    /**
     * If true, panel moves over elements on resize, defaults to true
     */
    isFloating?: boolean
    /**
     * CSS class applied to the resize handle
     */
    handleClassName?: string
    className?: string
    storageKey?: string
    defaultSize?: number
    position?: typeof PANEL_POSITIONS[number]
}

export const Panel: React.FunctionComponent<PanelProps> = ({
    children,
    className,
    defaultSize = 200,
    storageKey,
    position = 'bottom',
    isFloating = true,
    handleClassName,
}) => {
    const handleReference = useRef<HTMLDivElement | null>(null)
    const panelReference = useRef<HTMLDivElement | null>(null)

    const { panelSize, isResizing } = useResizablePanel({
        position,
        panelRef: panelReference,
        handleRef: handleReference,
        storageKey,
        defaultSize,
    })

    return (
        <div
            // eslint-disable-next-line react/forbid-dom-props
            style={{ [position === 'bottom' ? 'height' : 'width']: `${panelSize}px` }}
            className={classNames(
                className,
                styles.panel,
                getPositionStyle({ position }),
                getDisplayStyle({ isFloating })
            )}
            ref={panelReference}
        >
            <div
                ref={handleReference}
                role="presentation"
                className={classNames(
                    styles.handle,
                    styles[`handle${upperFirst(position)}` as keyof typeof styles],
                    handleClassName,
                    isResizing && styles.handleResizing
                )}
            />
            {children}
        </div>
    )
}
