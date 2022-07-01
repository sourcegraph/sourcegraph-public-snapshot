import React, { FunctionComponent, HTMLAttributes, LiHTMLAttributes, useEffect, useState } from 'react'

import classNames from 'classnames'

import { PopoverContent, Position, Point as PopoverPoint, createRectangle } from '@sourcegraph/wildcard'

import styles from './Tooltip.module.scss'

const TOOLTIP_PADDING = createRectangle(0, 0, 10, 10)

export const Tooltip: React.FunctionComponent = props => {
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

interface TooltipListProps extends HTMLAttributes<HTMLUListElement> {}

export const TooltipList: FunctionComponent<TooltipListProps> = props => (
    <ul {...props} className={classNames(styles.tooltipList, props.className)} />
)

export const TooltipListBlankItem: FunctionComponent<LiHTMLAttributes<HTMLLIElement>> = props => (
    <li {...props} className={classNames(styles.item, props.className)} />
)

interface TooltipListItemProps extends LiHTMLAttributes<HTMLLIElement> {
    isActive: boolean
    name: string
    value: number | string
    color: string
    stackedValue?: number | string | null
}

export const TooltipListItem: FunctionComponent<TooltipListItemProps> = props => {
    const { name, value, stackedValue, color, className, isActive, ...attributes } = props

    /* eslint-disable react/forbid-dom-props */
    return (
        <TooltipListBlankItem {...attributes} className={classNames(className, { [styles.itemActive]: isActive })}>
            <div style={{ backgroundColor: color }} className={styles.mark} />

            <span className={styles.legendText}>{name}</span>

            {stackedValue && (
                <span className={styles.legendStackedValue}>
                    {stackedValue}
                    {'\u00A0—\u00A0'}
                </span>
            )}

            <span>{value}</span>
        </TooltipListBlankItem>
    )
}
