/* eslint-disable react/forbid-dom-props */
import React, { useMemo } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router'

import { Link, Text, Tooltip } from '@sourcegraph/wildcard'

import { formatNumber } from '../utils'

import styles from './index.module.scss'

interface ValueLegendItemProps {
    color?: string
    description: string
    value: number | string
    tooltip?: string
    className?: string
    filter?: { name: string; value: string }
}

export const ValueLegendItem: React.FunctionComponent<ValueLegendItemProps> = ({
    value,
    color = 'var(--body-color)',
    description,
    tooltip,
    className,
    filter,
}) => {
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
        <div className={classNames('d-flex flex-column align-items-center mr-4 justify-content-center', className)}>
            <Tooltip content={tooltipOnNumber}>
                {filter ? (
                    <Link to={`?${searchParams.toString()}`} style={{ color }} className={styles.count}>
                        {formattedNumber}
                    </Link>
                ) : (
                    <span style={{ color }} className={styles.count}>
                        {formattedNumber}
                    </span>
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
    <div className={classNames('d-flex justify-content-between', className)}>
        <div className="d-flex justify-content-left">
            {items
                .filter(item => item.position !== 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
        <div className="d-flex justify-content-right">
            {items
                .filter(item => item.position === 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
    </div>
)
