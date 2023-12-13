import React, { useMemo, useState, useEffect } from 'react'

import classNames from 'classnames'

import type { TemporarySettingsSchema } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Card, Input, Text, H2 } from '@sourcegraph/wildcard'

import { AnalyticsDateRange } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { formatNumber } from '../utils'

import styles from './index.module.scss'

export const MIN_PER_ITEM_SAVED_KEY = 'minPerItemSaved'

interface TimeSavedCalculatorGroupItem {
    label: string
    value: number
    minPerItem: number
    onMinPerItemChange?: (minPerItem: number) => void
    description: string
    percentage?: number
    hoursSaved?: number
}

interface TimeSavedCalculatorGroupProps {
    page: string
    color: string
    value: number
    itemsLabel?: string
    label: string
    description: string
    dateRange: AnalyticsDateRange
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
    page,
    items,
    color,
    value,
    itemsLabel = 'Events',
    description,
    label,
    dateRange,
}) => {
    const [memoizedItems, setMemoizedItems] = useState(calculateHoursSaved(items))
    const [minutesInputChangeLogs, setMinutesInputChangeLogs] = useState<{ [index: number]: boolean }>({})
    const [percentageInputChangeLogs, setPercentageInputChangeLogs] = useState<{ [index: number]: boolean }>({})

    useEffect(() => {
        if (!items.length) {
            return
        }

        setMemoizedItems(calculateHoursSaved(items))
    }, [items])

    const totalSavedHours = useMemo(
        () => memoizedItems.reduce((sum, item) => sum + item.hoursSaved, 0),
        [memoizedItems]
    )

    const updateMinPerItem = (index: number, minPerItem: number): void => {
        const updatedItems = [...memoizedItems]
        updatedItems[index] = { ...memoizedItems[index], minPerItem }
        updatedItems[index].onMinPerItemChange?.(minPerItem)

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

    const projectedHoursSaved = useMemo(() => {
        if (dateRange === AnalyticsDateRange.LAST_WEEK) {
            return totalSavedHours * 52
        }
        if (dateRange === AnalyticsDateRange.LAST_MONTH) {
            return totalSavedHours * 12
        }
        if (dateRange === AnalyticsDateRange.LAST_THREE_MONTHS) {
            return (totalSavedHours * 12) / 3
        }
        return totalSavedHours
    }, [totalSavedHours, dateRange])

    return (
        <div>
            <Card className="mb-4 p-4">
                <div className="d-flex flex-row">
                    <div className="d-flex flex-column align-items-center mr-5">
                        <Text as="span" style={{ color }} alignment="center" className={styles.count}>
                            {formatNumber(value)}
                        </Text>
                        <Text
                            as="span"
                            alignment="center"
                            className="text-muted"
                            dangerouslySetInnerHTML={{ __html: label }}
                        />
                    </div>
                    <div className="d-flex flex-column align-items-center mr-5">
                        <Text as="span" className={styles.count}>
                            {formatNumber(totalSavedHours)}
                        </Text>
                        <Text as="span" alignment="center" className="text-muted">
                            Hours saved
                        </Text>
                    </div>
                    <div className="flex-1 d-flex flex-column m-0">
                        <Text as="span" weight="bold">
                            About this statistic
                        </Text>
                        <Text as="span" dangerouslySetInnerHTML={{ __html: description }} />
                    </div>
                </div>
                <div className="d-flex flex-column align-items-center mt-4">
                    <H2>
                        Annual projection:{' '}
                        <span className={styles.purpleColor}>{formatNumber(projectedHoursSaved)} hours</span> saved*
                    </H2>
                    <Text as="span" className="text-muted">
                        * Based on{' '}
                        {dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
                            ? 'last 3 months'
                            : dateRange === AnalyticsDateRange.LAST_MONTH
                            ? 'last month'
                            : 'last week'}{' '}
                        of data
                    </Text>
                </div>
            </Card>
            <div className={styles.calculatorList}>
                <div />
                {typeof memoizedItems[0]?.percentage === 'number' ? (
                    <Text as="span" className="text-muted">
                        % of total
                    </Text>
                ) : (
                    <Text as="span" alignment="center" className="text-muted">
                        {itemsLabel}
                    </Text>
                )}
                <Text as="span" className="text-nowrap text-muted">
                    Minutes per
                </Text>
                <Text as="span" alignment="center" className="text-muted">
                    Hours saved
                </Text>
                <div />
                {memoizedItems.map(({ label, percentage, minPerItem, hoursSaved, value, description }, index) => (
                    <React.Fragment key={label}>
                        <Text
                            className="text-nowrap d-flex align-items-center"
                            dangerouslySetInnerHTML={{ __html: label }}
                        />
                        {typeof percentage === 'number' ? (
                            <div className="d-flex flex-column align-items-center justify-content-center">
                                <Input
                                    type="number"
                                    value={percentage}
                                    className={classNames(styles.calculatorInput, 'mb-1')}
                                    onChange={event => {
                                        updatePercentage(index, Number(event.target.value))

                                        if (!percentageInputChangeLogs[index]) {
                                            setPercentageInputChangeLogs({
                                                ...percentageInputChangeLogs,
                                                [index]: true,
                                            })
                                            window.context.telemetryRecorder?.recordEvent(
                                                `AdminAnalytics${page}PercentageInput`,
                                                'edited'
                                            )
                                            eventLogger.log(`AdminAnalytics${page}PercentageInputEdited`)
                                        }
                                    }}
                                />
                            </div>
                        ) : (
                            <div className="d-flex flex-column align-items-center justify-content-center">
                                <Text as="span" weight="bold" className={styles.countBoxValue}>
                                    {formatNumber(value)}
                                </Text>
                            </div>
                        )}
                        <div className="d-flex flex-column align-items-center justify-content-center">
                            <Input
                                type="number"
                                value={minPerItem}
                                className={classNames(styles.calculatorInput, 'mb-1')}
                                onChange={event => {
                                    updateMinPerItem(index, Number(event.target.value))

                                    if (!minutesInputChangeLogs[index]) {
                                        setMinutesInputChangeLogs({
                                            ...minutesInputChangeLogs,
                                            [index]: true,
                                        })
                                        window.context.telemetryRecorder?.recordEvent(
                                            `AdminAnalytics${page}MinutesInput`,
                                            'edited'
                                        )
                                        eventLogger.log(`AdminAnalytics${page}MinutesInputEdited`)
                                    }
                                }}
                            />
                        </div>
                        <div className="d-flex flex-column align-items-center justify-content-center">
                            <Text as="span" weight="bold" className={styles.countBoxValue}>
                                {formatNumber(hoursSaved)}
                            </Text>
                        </div>
                        <Text dangerouslySetInnerHTML={{ __html: description }} className="d-flex align-items-center" />
                    </React.Fragment>
                ))}
            </div>
        </div>
    )
}

export interface TimeSavedCalculatorProps {
    page: string
    color: string
    label: string
    value: number
    defaultMinPerItem: number
    description: string
    percentage?: number
    dateRange: AnalyticsDateRange
    temporarySettingsKey: keyof TemporarySettingsSchema
}

export const TimeSavedCalculator: React.FunctionComponent<TimeSavedCalculatorProps> = ({
    page,
    color,
    label,
    value,
    defaultMinPerItem,
    description,
    percentage,
    dateRange,
    temporarySettingsKey,
}) => {
    const [minPerItemSavedSetting, setMinPerItemSaved] = useTemporarySetting(temporarySettingsKey, defaultMinPerItem)
    const minPerItemSaved = Number(minPerItemSavedSetting) || defaultMinPerItem
    const [inputChangeLogged, setInputChangeLogged] = useState(false)
    const hoursSaved = useMemo(
        () => (minPerItemSaved * value * (percentage ?? 100)) / (60 * 100),
        [value, minPerItemSaved, percentage]
    )

    const projectedHoursSaved = useMemo(() => {
        if (dateRange === AnalyticsDateRange.LAST_WEEK) {
            return hoursSaved * 52
        }
        if (dateRange === AnalyticsDateRange.LAST_MONTH) {
            return hoursSaved * 12
        }
        if (dateRange === AnalyticsDateRange.LAST_THREE_MONTHS) {
            return (hoursSaved * 12) / 3
        }
        return hoursSaved
    }, [hoursSaved, dateRange])

    const stringMinPerItemSaved = minPerItemSaved.toString()
    localStorage.setItem(MIN_PER_ITEM_SAVED_KEY, stringMinPerItemSaved)

    return (
        <Card className="mb-3 p-4">
            <div className="d-flex flex-row">
                <div className="flex-1 d-flex flex-row justify-content-between align-items-start">
                    <div className="d-flex flex-column align-items-center mr-5">
                        <Text as="span" style={{ color }} alignment="center" className={styles.count}>
                            {formatNumber(value)}
                        </Text>
                        <Text as="span" alignment="center" dangerouslySetInnerHTML={{ __html: label }} />
                    </div>
                    <div className="d-flex flex-column align-items-center justify-content-center">
                        <Input
                            type="number"
                            value={minPerItemSaved}
                            className={classNames(styles.calculatorInput, 'mb-1')}
                            onChange={event => {
                                setMinPerItemSaved(Number(event.target.value))
                                if (!inputChangeLogged) {
                                    setInputChangeLogged(true)
                                    window.context.telemetryRecorder?.recordEvent(
                                        `AdminAnalytics${page}MinutesInput`,
                                        'edited'
                                    )
                                    eventLogger.log(`AdminAnalytics${page}MinutesInputEdited`)
                                }
                            }}
                        />
                        <Text as="span" className="text-nowrap">
                            Minutes per
                        </Text>
                    </div>
                    <div className="d-flex flex-column align-items-center mr-5">
                        <Text as="span" weight="bold" className={styles.count}>
                            {formatNumber(hoursSaved)}
                        </Text>
                        <Text as="span" alignment="center">
                            Hours saved
                        </Text>
                    </div>
                </div>
                <div className="flex-1 d-flex flex-column m-0">
                    <Text as="span" weight="bold">
                        About this statistic
                    </Text>
                    <Text as="span" dangerouslySetInnerHTML={{ __html: description }} />
                </div>
            </div>
            <div className="d-flex flex-column align-items-center mt-4">
                <H2>
                    Annual projection:{' '}
                    <span className={styles.purpleColor}>{formatNumber(projectedHoursSaved)} hours</span> saved*
                </H2>
                <Text as="span" className="text-muted">
                    * Based on{' '}
                    {dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
                        ? 'last 3 months'
                        : dateRange === AnalyticsDateRange.LAST_MONTH
                        ? 'last month'
                        : 'last week'}{' '}
                    of data
                </Text>
            </div>
        </Card>
    )
}
