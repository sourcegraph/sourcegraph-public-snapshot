import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React, { useRef } from 'react'

import { PANEL_POSITIONS } from './constants'
import styles from './Panel.module.scss'
import { useResizablePanel } from './useResizablePanel'

export interface PanelProps {
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
                styles[`panel${upperFirst(position)}` as keyof typeof styles]
            )}
            ref={panelReference}
        >
            <div
                ref={handleReference}
                className={classNames(
                    styles.handle,
                    styles[`handle${upperFirst(position)}` as keyof typeof styles],
                    isResizing && styles.handleResizing
                )}
            />
            {children}
        </div>
    )
}
