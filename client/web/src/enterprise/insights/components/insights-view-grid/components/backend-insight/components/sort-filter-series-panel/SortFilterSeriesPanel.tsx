import type { ButtonHTMLAttributes, ChangeEventHandler, FC, FocusEventHandler, PropsWithChildren } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import {
    type SeriesSortOptionsInput,
    SeriesSortDirection,
    SeriesSortMode,
} from '../../../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SAMPLES, MAX_NUMBER_OF_SERIES } from '../../../../../../constants'
import type { InsightSeriesDisplayOptions } from '../../../../../../core/types/insight/common'
import type { DrillDownFiltersFormValues } from '../drill-down-filters-panel'

import styles from './SortFilterSeriesPanel.module.scss'

interface SortFilterSeriesPanelProps {
    value: InsightSeriesDisplayOptions
    isNumSamplesFilterAvailable: boolean
    onChange: (parameter: DrillDownFiltersFormValues['seriesDisplayOptions']) => void
}

export const SortFilterSeriesPanel: FC<SortFilterSeriesPanelProps> = props => {
    const { value, isNumSamplesFilterAvailable, onChange } = props

    const handleToggle = (sortOptions: SeriesSortOptionsInput): void => {
        onChange({ ...value, sortOptions })
    }

    const handleSeriesCountChange: ChangeEventHandler<HTMLInputElement> = event => {
        const inputValue = event.target.value

        // If a value is provided, clamp that value between 1 and maxLimit
        if (inputValue.length > 0) {
            const limit = Math.max(Math.min(parseInt(inputValue, 10), MAX_NUMBER_OF_SERIES), 1)
            onChange({ ...value, limit })
        } else {
            onChange({ ...value, limit: null })
        }
    }

    const handleNumCountChange: ChangeEventHandler<HTMLInputElement> = event => {
        const inputValue = event.target.value

        // If a value is provided, clamp that value between 1 and maxLimit
        if (inputValue.length > 0) {
            const numSamples = Math.max(Math.min(parseInt(inputValue, 10), MAX_NUMBER_OF_SAMPLES), 1)
            onChange({ ...value, numSamples })
        } else {
            onChange({ ...value, numSamples: null })
        }
    }

    const handleSeriesCountBlur: FocusEventHandler<HTMLInputElement> = event => {
        const limit = event.target.value

        if (limit === '') {
            onChange({ ...value, limit: null })
        }
    }

    const handleNumCountBlur: FocusEventHandler<HTMLInputElement> = event => {
        const limit = event.target.value

        if (limit === '') {
            onChange({ ...value, numSamples: null })
        }
    }

    return (
        <section>
            <section className={classNames(styles.togglesContainer)}>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by result count</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            aria-label="Sort by result count with descending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.RESULT_COUNT, direction: SeriesSortDirection.DESC }}
                            onToggle={handleToggle}
                        >
                            Highest
                        </ToggleButton>
                        <ToggleButton
                            aria-label="Sort by result count with ascending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.RESULT_COUNT, direction: SeriesSortDirection.ASC }}
                            onToggle={handleToggle}
                        >
                            Lowest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by name</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            aria-label="Sort by name with ascending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.LEXICOGRAPHICAL, direction: SeriesSortDirection.ASC }}
                            onToggle={handleToggle}
                        >
                            A-Z
                        </ToggleButton>
                        <ToggleButton
                            aria-label="Sort by name with descending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.LEXICOGRAPHICAL, direction: SeriesSortDirection.DESC }}
                            onToggle={handleToggle}
                        >
                            Z-A
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by date added</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            aria-label="Sort by date with descending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.DATE_ADDED, direction: SeriesSortDirection.DESC }}
                            onToggle={handleToggle}
                        >
                            Newest
                        </ToggleButton>
                        <ToggleButton
                            aria-label="Sort by date with ascending order"
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.DATE_ADDED, direction: SeriesSortDirection.ASC }}
                            onToggle={handleToggle}
                        >
                            Oldest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
            </section>
            <section className={styles.footer}>
                <span>
                    Max number of data series to display{' '}
                    <small className="text-muted">(max {MAX_NUMBER_OF_SERIES})</small>
                </span>
                <Input
                    type="number"
                    step="1"
                    min={1}
                    max={MAX_NUMBER_OF_SERIES}
                    placeholder={`${MAX_NUMBER_OF_SERIES}`}
                    value={value.limit ?? undefined}
                    onChange={handleSeriesCountChange}
                    onBlur={handleSeriesCountBlur}
                    variant="small"
                    aria-label="Number of data series"
                />
            </section>
            {isNumSamplesFilterAvailable && (
                <section className={styles.footer}>
                    <span>
                        Max number of series points to display <small className="text-muted">(max 90)</small>
                    </span>
                    <Input
                        type="number"
                        step="1"
                        min={1}
                        max={90}
                        value={value.numSamples ?? undefined}
                        placeholder="90"
                        onChange={handleNumCountChange}
                        onBlur={handleNumCountBlur}
                        variant="small"
                        aria-label="Number of data series"
                    />
                </section>
            )}
        </section>
    )
}

interface ToggleButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'value'> {
    selected: SeriesSortOptionsInput
    value: SeriesSortOptionsInput
    onToggle: (value: SeriesSortOptionsInput) => void
}

const ToggleButton: FC<PropsWithChildren<ToggleButtonProps>> = ({
    selected,
    value,
    children,
    onToggle,
    ...attributes
}) => {
    const isSelected = selected.mode === value.mode && selected.direction === value.direction

    return (
        <Button
            {...attributes}
            variant="secondary"
            size="sm"
            className={classNames({ [styles.selected]: isSelected, [styles.unselected]: !isSelected })}
            onClick={() => onToggle(value)}
        >
            {children}
        </Button>
    )
}
