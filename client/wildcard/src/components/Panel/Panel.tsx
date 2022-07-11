import React, { useRef } from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'
import { useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../types'

import { useResizablePanel, UseResizablePanelParameters } from './useResizablePanel'
import { getDisplayStyle, getPositionStyle } from './utils'

import styles from './Panel.module.scss'

export interface PanelProps extends Omit<UseResizablePanelParameters, 'panelRef' | 'handleRef'> {
    /**
     * If true, panel moves over elements on resize
     *
     * @default false
     */
    isFloating?: boolean
    /**
     * CSS class applied to the resize handle
     */
    handleClassName?: string
    className?: string
}

export const Panel = React.forwardRef(function Panel(
    {
        children,
        className,
        defaultSize = 200,
        storageKey,
        position = 'bottom',
        isFloating = false,
        handleClassName,
        minSize,
        maxSize,
    },
    reference
) {
    const handleReference = useRef<HTMLDivElement | null>(null)
    const panelReference = useRef<HTMLDivElement | null>(null)
    const mergedPanelReference = useMergeRefs([reference, panelReference])

    const { panelSize, isResizing } = useResizablePanel({
        position,
        panelRef: panelReference,
        handleRef: handleReference,
        storageKey,
        defaultSize,
        minSize,
        maxSize,
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
            ref={mergedPanelReference}
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
}) as ForwardReferenceComponent<'div', React.PropsWithChildren<PanelProps>>
