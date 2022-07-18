import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import { SeriesSortOptionsInput, SeriesSortDirection, SeriesSortMode } from '../../../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SERIES } from '../../../../../../core/backend/gql-backend/methods/get-backend-insight-data/deserializators'
import { DrillDownFiltersFormValues } from '../drill-down-filters-panel'

import styles from './SortFilterSeriesPanel.module.scss'

interface SortFilterSeriesPanelProps {
    value: {
        limit: string
        sortOptions: SeriesSortOptionsInput
    }
    seriesCount: number
    onChange: (parameter: DrillDownFiltersFormValues['seriesDisplayOptions']) => void
}

export const SortFilterSeriesPanel: React.FunctionComponent<SortFilterSeriesPanelProps> = ({
    value,
    seriesCount,
    onChange,
}) => {
    // It is possible to have N number of series, but we need to have maximum to render in UI
    // or else it gets too cluttered to view
    const maxLimit = Math.min(seriesCount, MAX_NUMBER_OF_SERIES)

    const handleToggle = (sortOptions: SeriesSortOptionsInput): void => {
        onChange({ ...value, sortOptions })
    }

    const handleChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const inputValue = event.target.value
        let limit = inputValue

        // If a value is provided, clamp that value between 1 and maxLimit
        if (inputValue.length > 0) {
            limit = Math.max(Math.min(parseInt(inputValue, 10), maxLimit), 1).toString()
        }
        onChange({ ...value, limit })
    }

    const handleBlur: React.FocusEventHandler<HTMLInputElement> = event => {
        const limit = event.target.value
        if (limit === '') {
            onChange({ ...value, limit: `${maxLimit}` })
        }
    }

    return (
        <section>
            <section className={classNames(styles.togglesContainer)}>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by result count</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.RESULT_COUNT, direction: SeriesSortDirection.DESC }}
                            onClick={handleToggle}
                        >
                            Highest
                        </ToggleButton>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.RESULT_COUNT, direction: SeriesSortDirection.ASC }}
                            onClick={handleToggle}
                        >
                            Lowest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by name</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.LEXICOGRAPHICAL, direction: SeriesSortDirection.ASC }}
                            onClick={handleToggle}
                        >
                            A-Z
                        </ToggleButton>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.LEXICOGRAPHICAL, direction: SeriesSortDirection.DESC }}
                            onClick={handleToggle}
                        >
                            Z-A
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by date added</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.DATE_ADDED, direction: SeriesSortDirection.DESC }}
                            onClick={handleToggle}
                        >
                            Newest
                        </ToggleButton>
                        <ToggleButton
                            selected={value.sortOptions}
                            value={{ mode: SeriesSortMode.DATE_ADDED, direction: SeriesSortDirection.ASC }}
                            onClick={handleToggle}
                        >
                            Oldest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
            </section>
            <footer className={styles.footer}>
                <span>
                    Number of data series <small className="text-muted">(max {maxLimit})</small>
                </span>
                <Input
                    type="number"
                    step="1"
                    min={1}
                    max={maxLimit}
                    value={value.limit}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    variant="small"
                />
            </footer>
        </section>
    )
}

interface ToggleButtonProps {
    selected: SeriesSortOptionsInput
    value: SeriesSortOptionsInput
    onClick: (value: SeriesSortOptionsInput) => void
}

const ToggleButton: React.FunctionComponent<React.PropsWithChildren<ToggleButtonProps>> = ({
    selected,
    value,
    children,
    onClick,
}) => {
    const isSelected = selected.mode === value.mode && selected.direction === value.direction

    return (
        <Button
            variant="secondary"
            size="sm"
            className={classNames({ [styles.selected]: isSelected, [styles.unselected]: !isSelected })}
            onClick={() => onClick(value)}
        >
            {children}
        </Button>
    )
}
