import React, { useMemo, useState, useEffect } from 'react'

import classNames from 'classnames'

import { Card, Input, Text } from '@sourcegraph/wildcard'

import { formatNumber } from '../lib'

import styles from '../index.module.scss'

interface TimeSavedCalculatorGroupItem {
    label: string
    value: number
    minPerItem: number
    description: string
    percentage?: number
    hoursSaved?: number
}

interface TimeSavedCalculatorGroupProps {
    color: string
    value: number
    label: string
    description: string
    items: TimeSavedCalculatorGroupItem[]
}

const calculateHoursSaved = (
    items: TimeSavedCalculatorGroupItem[]
): (TimeSavedCalculatorGroupItem & { hoursSaved: number })[] =>
    items.map(item => ({
        ...item,
        hoursSaved: (item.minPerItem * item.value * (item.percentage ?? 100)) / (60 * 100),
    }))

export const TimeSavedCalculatorGroup: React.FunctionComponent<TimeSavedCalculatorGroupProps> = ({
    items,
    color,
    value,
    description,
    label,
}) => {
    const [memoizedItems, setMemoizedItems] = useState(calculateHoursSaved(items))

    useEffect(() => {
        if (!items.length) {
            return
        }

        setMemoizedItems(calculateHoursSaved(items))
    }, [items])

    const totalSavedHours = useMemo(() => memoizedItems.reduce((sum, item) => sum + item.hoursSaved, 0), [
        memoizedItems,
    ])

    const updateMinPerItem = (index: number, minPerItem: number): void => {
        const updatedItems = [...memoizedItems]
        updatedItems[index] = { ...memoizedItems[index], minPerItem }

        setMemoizedItems(calculateHoursSaved(updatedItems))
    }

    const updatePercentage = (index: number, percentage: number = 0): void => {
        if (!memoizedItems.length || percentage > 100 || percentage < 0) {
            return
        }

        const updatedItems = [...memoizedItems]

        if (index !== 0 || memoizedItems.length > 1) {
            let deltaPercentage = (memoizedItems[index].percentage ?? 100) - percentage

            for (
                let listIndex = index + 1;
                listIndex % memoizedItems.length !== index && deltaPercentage !== 0;
                listIndex++
            ) {
                const itemIndex = listIndex % memoizedItems.length

                const item = memoizedItems[itemIndex]

                const updatedPercentage = Math.min(Math.max((item.percentage ?? 100) + deltaPercentage, 0), 100)

                updatedItems[itemIndex] = {
                    ...item,
                    percentage: updatedPercentage,
                }

                deltaPercentage -= updatedPercentage - (item.percentage ?? 100)
            }
        }

        updatedItems[index] = { ...memoizedItems[index], percentage }

        setMemoizedItems(calculateHoursSaved(updatedItems))
    }

    return (
        <div>
            <Card className="mb-3 p-4 d-flex flex-row">
                <div className="d-flex flex-column align-items-center mr-5">
                    <Text as="span" style={{ color }} alignment="center" className={styles.count}>
                        {formatNumber(value)}
                    </Text>
                    <Text as="span" alignment="center" dangerouslySetInnerHTML={{ __html: label }} />
                </div>
                <div className="d-flex flex-column align-items-center mr-5">
                    <Text as="span" className={styles.count}>
                        {formatNumber(totalSavedHours)}
                    </Text>
                    <Text as="span" alignment="center">
                        Hours saved
                    </Text>
                </div>
                <div className="flex-1 d-flex flex-column m-0">
                    <Text as="span" weight="bold">
                        About this statistics
                    </Text>
                    <Text as="span" dangerouslySetInnerHTML={{ __html: description }} />
                </div>
            </Card>
            <div className={styles.calculatorList}>
                {memoizedItems.map(({ label, percentage, minPerItem, hoursSaved, value, description }, index) => (
                    <React.Fragment key={label}>
                        <Text
                            className="text-nowrap d-flex align-items-center"
                            dangerouslySetInnerHTML={{ __html: label }}
                        />
                        {!!percentage && percentage >= 0 ? (
                            <div className="d-flex flex-column align-items-center justify-content-center">
                                <Input
                                    type="number"
                                    value={percentage}
                                    className={classNames(styles.calculatorInput, 'mb-1')}
                                    onChange={event => updatePercentage(index, Number(event.target.value))}
                                />
                                <Text as="span">% of total</Text>
                            </div>
                        ) : (
                            <div className="d-flex flex-column align-items-center justify-content-center">
                                <Text as="span" weight="bold" className={styles.countBoxValue}>
                                    {formatNumber(value)}
                                </Text>
                                <Text as="span" alignment="center">
                                    Events
                                </Text>
                            </div>
                        )}
                        <div className="d-flex flex-column align-items-center justify-content-center">
                            <Input
                                type="number"
                                value={minPerItem}
                                className={classNames(styles.calculatorInput, 'mb-1')}
                                onChange={event => updateMinPerItem(index, Number(event.target.value))}
                            />
                            <Text as="span" className="text-nowrap">
                                Minutes per
                            </Text>
                        </div>
                        <div className="d-flex flex-column align-items-center justify-content-center">
                            <Text as="span" weight="bold" className={styles.countBoxValue}>
                                {formatNumber(hoursSaved)}
                            </Text>
                            <Text as="span" alignment="center">
                                Hours saved
                            </Text>
                        </div>
                        <Text dangerouslySetInnerHTML={{ __html: description }} className="d-flex align-items-center" />
                    </React.Fragment>
                ))}
            </div>
        </div>
    )
}
