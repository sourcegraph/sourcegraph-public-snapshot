import React, { useEffect, useState } from 'react'

import { PopoverContent, Position, Point as PopoverPoint, createRectangle } from '@sourcegraph/wildcard'

import styles from './Tooltip.module.scss'

const TOOLTIP_PADDING = createRectangle(0, 0, 10, 10)

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
