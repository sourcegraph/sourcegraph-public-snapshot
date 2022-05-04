import React, { useEffect, useState } from 'react'

import { PopoverContent, Position, Point as PopoverPoint, createRectangle } from '@sourcegraph/wildcard'

import { Series } from '../../../../types'

import styles from './Tooltip.module.scss'

const TOOLTIP_PADDING = createRectangle(0, 0, 10, 10)

/**
 * Default value for line color in case if we didn't get color for line from content config.
 */
export const DEFAULT_LINE_STROKE = 'var(--gray-07)'

export function getLineStroke<Datum>(line: Series<Datum>): string {
    return line?.color ?? DEFAULT_LINE_STROKE
}

interface TooltipProps {
    reference?: HTMLElement
}

export const Tooltip: React.FunctionComponent<React.PropsWithChildren<TooltipProps>> = props => {
    const [virtualElement, setVirtualElement] = useState<PopoverPoint | null>(null)

    useEffect(() => {
        function handleMove(event: PointerEvent): void {
            setVirtualElement({
                x: event.clientX,
                y: event.clientY,
            })
        }

        window.addEventListener('pointermove', handleMove)
        window.addEventListener('pointerleave', () => setVirtualElement(null))

        return () => {
            window.removeEventListener('pointermove', handleMove)
        }
    }, [])

    return (
        virtualElement && (
            <PopoverContent
                isOpen={true}
                pin={virtualElement}
                targetPadding={TOOLTIP_PADDING}
                position={Position.rightStart}
                className={styles.tooltip}
            >
                {props.children}
            </PopoverContent>
        )
    )
}
