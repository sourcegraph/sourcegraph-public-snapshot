/* eslint-disable react/forbid-dom-props */
import React from 'react'

import classNames from 'classnames'

import { Text } from '@sourcegraph/wildcard'

import { formatNumber } from '../utils'

import styles from './index.module.scss'

interface ValueLegendItemProps {
    color: string
    description: string
    value: number
}

const ValueLegendItem: React.FunctionComponent<ValueLegendItemProps> = ({ value, color, description }) => (
    <div className="d-flex flex-column align-items-center mr-4 justify-content-center">
        <span style={{ color }} className={styles.count}>
            {formatNumber(value)}
        </span>
        <Text as="span" alignment="center" className={styles.textWrap}>
            {description}
        </Text>
    </div>
)

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
