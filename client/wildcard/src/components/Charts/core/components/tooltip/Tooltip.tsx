import React, { type FunctionComponent, type HTMLAttributes, type LiHTMLAttributes } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

// In order to resolve cyclic deps in tests
// see https://github.com/sourcegraph/sourcegraph/pull/40209#pullrequestreview-1069334480
import { createRectangle, Popover, PopoverContent, PopoverTail, Position } from '../../../../Popover'

import styles from './Tooltip.module.scss'

const TOOLTIP_PADDING = createRectangle(0, 0, 5, 5)

interface TooltipProps {
    activeElement: HTMLElement
}

export const Tooltip: React.FunctionComponent<React.PropsWithChildren<TooltipProps>> = props => {
    const { activeElement, children } = props

    return (
        <Popover isOpen={true} onOpenChange={noop}>
            <PopoverContent
                target={activeElement}
                position={Position.right}
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

            <PopoverTail size="sm" tabIndex={-1} className={styles.tail} />
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
