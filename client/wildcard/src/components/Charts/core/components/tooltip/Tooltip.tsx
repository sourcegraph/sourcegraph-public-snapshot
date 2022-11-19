import React, { FunctionComponent, HTMLAttributes, LiHTMLAttributes, useEffect, useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

// In order to resolve cyclic deps in tests
// see https://github.com/sourcegraph/sourcegraph/pull/40209#pullrequestreview-1069334480
import { createRectangle, Popover, PopoverContent, PopoverTail, Position } from '../../../../Popover'

import styles from './Tooltip.module.scss'

const TOOLTIP_PADDING = createRectangle(0, 0, 10, 10)

interface TooltipProps {
    containerElement: Element
    activeElement?: HTMLElement
}

interface TooltipPosition {
    target: HTMLElement | null
    x: number
    y: number
}

export const Tooltip: React.FunctionComponent<React.PropsWithChildren<TooltipProps>> = props => {
    const { containerElement, activeElement = null, children } = props

    const [{ target, ...pinPoint }, setVirtualElement] = useState<TooltipPosition>({
        target: activeElement,
        x: 0,
        y: 0,
    })

    useLayoutEffect(() => {
        if (activeElement) {
            setVirtualElement(state => ({ ...state, target: activeElement }))
        }
    }, [activeElement])

    useEffect(() => {
        // We need this casting because Element type doesn't support
        // pointer or mouse events in pointermove handlers
        const element = containerElement as HTMLElement

        function handleMove(event: PointerEvent): void {
            setVirtualElement({
                target: null,
                x: event.clientX,
                // -10 because we want to visually compensate space for the
                // popover tail
                y: event.clientY - 10,
            })
        }

        element.addEventListener('pointermove', handleMove)

        return () => {
            element.removeEventListener('pointermove', handleMove)
        }
    }, [containerElement])

    const withRealTarget = !!target
    const tooltipPinPoint = withRealTarget ? null : pinPoint
    const tooltipTarget = withRealTarget ? target : null
    const tooltipPosition = withRealTarget ? Position.right : Position.rightStart

    return (
        <Popover isOpen={true} onOpenChange={noop}>
            <PopoverContent
                pin={tooltipPinPoint}
                target={tooltipTarget}
                position={tooltipPosition}
                autoFocus={false}
                focusLocked={false}
                returnTargetFocus={false}
                targetPadding={TOOLTIP_PADDING}
                className={styles.tooltip}
                // Hide Tooltip UI from screen reader otherwise Voice over in Safari will
                // catch this element and interrupt the original chart screen reader navigation flow
                aria-hidden={true}
                tabIndex={-1}
            >
                {children}
            </PopoverContent>

            <PopoverTail size='sm' tabIndex={-1}/>
        </Popover>
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
