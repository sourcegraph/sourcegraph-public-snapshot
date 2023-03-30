import React, { useMemo } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { Link, LoadingSpinner, Text, Tooltip } from '@sourcegraph/wildcard'

import { formatNumber } from '../utils'

import styles from './index.module.scss'

interface ValueLegendItemProps {
    color?: string
    description: string
    value: number | string
    tooltip?: string
    className?: string
    filter?: { name: string; value: string }
    onClick?: () => any
}

export const ValueLegendItem: React.FunctionComponent<ValueLegendItemProps> = ({
    value,
    color = 'var(--body-color)',
    description,
    tooltip,
    className,
    filter,
    onClick,
}) => {
    if (value === 'loading') {
        return <LoadingSpinner className={classNames(styles.count, 'my-auto')} />
    }

    const formattedNumber = useMemo(() => (typeof value === 'number' ? formatNumber(value) : value), [value])
    const unformattedNumber = `${value}`
    const location = useLocation()

    const searchParams = useMemo(() => {
        const search = new URLSearchParams(location.search)
        if (filter) {
            search.set(filter.name, filter.value)
        }
        return search
    }, [filter, location.search])

    const tooltipOnNumber =
        formattedNumber !== unformattedNumber
            ? isNaN(parseFloat(unformattedNumber))
                ? unformattedNumber
                : Intl.NumberFormat('en').format(parseFloat(unformattedNumber))
            : undefined

    return (
        <div className={classNames(styles.legendItem, className)}>
            <Tooltip content={tooltipOnNumber}>
                {filter ? (
                    <Link to={`?${searchParams.toString()}`} style={{ color }} className={styles.count}>
                        {formattedNumber}
                    </Link>
                ) : (
                    <Text
                        as="span"
                        alignment="center"
                        style={{ color }}
                        className={classNames(styles.count, 'cursor-pointer')}
                        onClick={onClick}
                    >
                        {formattedNumber}
                    </Text>
                )}
            </Tooltip>
            <Tooltip content={tooltip}>
                {filter ? (
                    <Link
                        to={`?${searchParams.toString()}`}
                        className={classNames(styles.textWrap, tooltip && 'cursor-pointer', 'text-muted')}
                    >
                        {description}
                        {tooltip && <span className={styles.linkColor}>*</span>}
                    </Link>
                ) : (
                    <Text
                        as="span"
                        alignment="center"
                        className={classNames(styles.textWrap, tooltip && 'cursor-pointer', 'text-muted')}
                        onClick={onClick}
                    >
                        {description}
                        {tooltip && <span className={styles.linkColor}>*</span>}
                    </Text>
                )}
            </Tooltip>
        </div>
    )
}

export interface ValueLegendListProps {
    className?: string
    items: (ValueLegendItemProps & { position?: 'left' | 'right' })[]
}

export const ValueLegendList: React.FunctionComponent<ValueLegendListProps> = ({ items, className }) => (
    <div className={classNames(styles.legend, className)}>
        <div className={styles.legendLeftPanel}>
            {items
                .filter(item => item.position !== 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
        <div className={styles.legendRightPanel}>
            {items
                .filter(item => item.position === 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
    </div>
)
