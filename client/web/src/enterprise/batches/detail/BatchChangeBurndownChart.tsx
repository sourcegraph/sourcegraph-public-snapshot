import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import { getYear, parseISO } from 'date-fns'

// for polyfill
import 'events'

import {
    Area,
    ComposedChart,
    type LabelFormatter,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
    type TooltipPayload,
} from 'recharts'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Checkbox, Container, LoadingSpinner, Label, ErrorAlert } from '@sourcegraph/wildcard'

import type { ChangesetCountsOverTimeFields, Scalars } from '../../../graphql-operations'

import { useChangesetCountsOverTime } from './backend'

import styles from './BatchChangeBurndownChart.module.scss'

interface Props {
    batchChangeID: Scalars['ID']
    width?: string | number
}

// const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })
const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
const tooltipLabelFormatter = (date: number): string => tooltipLabelFormat.format(date)

const toLocaleString = (value: number): string => value.toLocaleString()

const tooltipStyle: React.CSSProperties = {
    color: 'var(--body-color)',
    border: 'none',
    background: 'var(--body-bg)',
}

const commonAreaProps = {
    isAnimationActive: false,
    strokeWidth: 0,
    stackId: 'stack',
    type: 'stepBefore',
} as const

interface StateDefinition {
    fill: string
    label: string
    sortOrder: number
}

type DisplayableChangesetCounts = Pick<
    ChangesetCountsOverTimeFields,
    'openPending' | 'openChangesRequested' | 'openApproved' | 'closed' | 'merged' | 'draft'
>

const states: Record<keyof DisplayableChangesetCounts, StateDefinition> = {
    draft: { fill: 'var(--text-muted)', label: 'Draft', sortOrder: 5 },
    openPending: { fill: 'var(--warning)', label: 'Awaiting review', sortOrder: 4 },
    openChangesRequested: { fill: 'var(--danger)', label: 'Changes requested', sortOrder: 3 },
    openApproved: { fill: 'var(--success)', label: 'Approved', sortOrder: 2 },
    closed: { fill: 'var(--secondary)', label: 'Closed', sortOrder: 1 },
    merged: { fill: 'var(--merged)', label: 'Merged', sortOrder: 0 },
}

const tooltipItemSorter = ({ dataKey }: TooltipPayload): number =>
    states[dataKey as keyof DisplayableChangesetCounts].sortOrder

/**
 * A burndown chart showing progress of the batch change's changesets.
 */
export const BatchChangeBurndownChart: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    batchChangeID,
    width = '100%',
}) => {
    const [includeArchived, setIncludeArchived] = useState<boolean>(false)
    const toggleIncludeArchived = useCallback((): void => setIncludeArchived(previousValue => !previousValue), [])
    const [hiddenStates, setHiddenStates] = useState<Set<keyof DisplayableChangesetCounts>>(new Set())

    const { loading, data, error } = useChangesetCountsOverTime(batchChangeID, includeArchived)

    const dateTickFormatter = useMemo(() => {
        let dateTickFormat = new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric' })
        if (!(data?.node?.__typename === 'BatchChange')) {
            return (timestamp: number): string => dateTickFormat.format(timestamp)
        }
        const changesetCountsOverTime = data.node.changesetCountsOverTime
        if (changesetCountsOverTime.length > 1) {
            const start = parseISO(changesetCountsOverTime[0].date)
            const end = parseISO(changesetCountsOverTime.at(-1)!.date)
            // If the range spans multiple years, we want to display the year as well.
            if (getYear(start) !== getYear(end)) {
                dateTickFormat = new Intl.DateTimeFormat(undefined, {
                    month: 'short',
                    day: 'numeric',
                    year: '2-digit',
                })
            }
        }
        return (timestamp: number): string => dateTickFormat.format(timestamp)
    }, [data])

    if (!loading && error) {
        return <ErrorAlert error={error} />
    }

    if (loading && !data) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }

    if (!data) {
        // Shouldn't happen.
        return null
    }

    if (!data.node) {
        return <ErrorAlert error={new Error(`BatchChange with ID ${batchChangeID} does not exist`)} />
    }

    if (data.node.__typename !== 'BatchChange') {
        return <ErrorAlert error={new Error(`The given ID is a ${data.node.__typename}, not a BatchChange`)} />
    }

    const changesetCountsOverTime = data.node.changesetCountsOverTime

    return (
        <Container>
            <div className={classNames(styles.batchChangeBurndownChartContainer, 'd-flex align-items-center')}>
                <ResponsiveContainer width={width} height={300} className="test-batches-chart">
                    <ComposedChart
                        data={changesetCountsOverTime.map(snapshot => ({
                            ...snapshot,
                            date: Date.parse(snapshot.date),
                        }))}
                    >
                        <XAxis
                            dataKey="date"
                            domain={[changesetCountsOverTime[0].date, changesetCountsOverTime.at(-1)!.date]}
                            name="Time"
                            tickFormatter={dateTickFormatter}
                            type="number"
                            stroke="var(--text-muted)"
                            scale="time"
                        />
                        <YAxis
                            tickFormatter={toLocaleString}
                            stroke="var(--text-muted)"
                            type="number"
                            allowDecimals={false}
                            domain={[0, 'dataMax']}
                        />
                        <Tooltip
                            labelFormatter={tooltipLabelFormatter as LabelFormatter}
                            isAnimationActive={false}
                            wrapperStyle={{ border: '1px solid var(--border-color)' }}
                            contentStyle={tooltipStyle}
                            labelStyle={{ fontWeight: 'bold' }}
                            itemStyle={tooltipStyle}
                            itemSorter={tooltipItemSorter}
                        />

                        {Object.entries(states)
                            .sort(([, a], [, b]) => b.sortOrder - a.sortOrder)
                            .filter(([dataKey]) => !hiddenStates.has(dataKey as keyof DisplayableChangesetCounts))
                            .map(([dataKey, state]) => (
                                <Area
                                    key={state.sortOrder}
                                    dataKey={dataKey}
                                    name={state.label}
                                    fill={state.fill}
                                    // The stroke is used to color the legend, which we
                                    // want to match the fill color for each area.
                                    stroke={state.fill}
                                    {...commonAreaProps}
                                />
                            ))}
                    </ComposedChart>
                </ResponsiveContainer>
                <div className="flex-grow-0 ml-2">
                    {Object.entries(states).map(([key, state]) => (
                        <LegendLabel
                            key={key}
                            stateKey={key as keyof DisplayableChangesetCounts}
                            label={state.label}
                            fill={state.fill}
                            hiddenStates={hiddenStates}
                            setHiddenStates={setHiddenStates}
                        />
                    ))}
                    <hr className="flex-grow-1" />
                    <IncludeArchivedToggle includeArchived={includeArchived} onToggle={toggleIncludeArchived} />
                </div>
            </div>
        </Container>
    )
}

const LegendLabel: React.FunctionComponent<
    React.PropsWithChildren<{
        stateKey: keyof DisplayableChangesetCounts
        label: string
        fill: string
        hiddenStates: Set<keyof DisplayableChangesetCounts>
        setHiddenStates: (
            setter: (currentValue: Set<keyof DisplayableChangesetCounts>) => Set<keyof DisplayableChangesetCounts>
        ) => void
    }>
> = ({ stateKey, label, fill, hiddenStates, setHiddenStates }) => {
    const onChangeCheckbox = useCallback(() => {
        setHiddenStates(current => {
            if (current.has(stateKey)) {
                const newSet = new Set<keyof DisplayableChangesetCounts>(current)
                newSet.delete(stateKey)
                return newSet
            }
            return new Set<keyof DisplayableChangesetCounts>(current).add(stateKey)
        })
    }, [setHiddenStates, stateKey])
    const checked = useMemo(() => !hiddenStates.has(stateKey), [hiddenStates, stateKey])
    return (
        <div className="d-flex align-items-center text-nowrap p-2">
            <div
                // We want to set the fill based on the state config.
                // eslint-disable-next-line react/forbid-dom-props
                style={{
                    backgroundColor: fill,
                }}
                className={classNames(styles.batchChangeBurndownChartLegendColorBox, 'mr-2')}
            />
            <Checkbox id={stateKey} checked={checked} onChange={onChangeCheckbox} label={label} />
        </div>
    )
}

const IncludeArchivedToggle: React.FunctionComponent<
    React.PropsWithChildren<{
        includeArchived: boolean
        onToggle: () => void
    }>
> = ({ includeArchived, onToggle }) => (
    <div className="d-flex align-items-center justify-content-between text-nowrap mb-2 pt-1">
        <Label htmlFor="include-archived" className="mb-0">
            Include archived
        </Label>
        <Toggle
            id="include-archived"
            value={includeArchived}
            onToggle={onToggle}
            title="Include archived changesets"
            className="ml-2"
            display="inline"
        />
    </div>
)
