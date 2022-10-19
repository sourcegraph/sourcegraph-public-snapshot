import { ButtonHTMLAttributes, ChangeEventHandler, FC, FocusEventHandler, PropsWithChildren } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import { SeriesSortOptionsInput, SeriesSortDirection, SeriesSortMode } from '../../../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SERIES } from '../../../../../../constants'
import { DrillDownFiltersFormValues } from '../drill-down-filters-panel'

import styles from './SortFilterSeriesPanel.module.scss'

interface SortFilterSeriesPanelProps {
    value: {
        limit: string
        sortOptions: SeriesSortOptionsInput
    }
    onChange: (parameter: DrillDownFiltersFormValues['seriesDisplayOptions']) => void
}

export const SortFilterSeriesPanel: FC<SortFilterSeriesPanelProps> = ({ value, onChange }) => {
    const handleToggle = (sortOptions: SeriesSortOptionsInput): void => {
        onChange({ ...value, sortOptions })
    }

    const handleChange: ChangeEventHandler<HTMLInputElement> = event => {
        const inputValue = event.target.value
        let limit = inputValue

        // If a value is provided, clamp that value between 1 and maxLimit
        if (inputValue.length > 0) {
            limit = Math.max(Math.min(parseInt(inputValue, 10), MAX_NUMBER_OF_SERIES), 1).toString()
        }
        onChange({ ...value, limit })
    }

    const handleBlur: FocusEventHandler<HTMLInputElement> = event => {
        const limit = event.target.value
        if (limit === '') {
            onChange({ ...value, limit: `${MAX_NUMBER_OF_SERIES}` })
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
            <footer className={styles.footer}>
                <span>
                    Number of data series <small className="text-muted">(max {MAX_NUMBER_OF_SERIES})</small>
                </span>
                <Input
                    type="number"
                    step="1"
                    min={1}
                    max={MAX_NUMBER_OF_SERIES}
                    value={value.limit}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    variant="small"
                    aria-label="Number of data series"
                />
            </footer>
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
