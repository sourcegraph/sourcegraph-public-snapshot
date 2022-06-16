import { useState } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import { SeriesSortOptionsInput, SeriesSortDirection, SeriesSortMode } from '../../../../../../../../graphql-operations'
import { SeriesDisplayOptionsInputRequired } from '../../../../../../core/types/insight/common'

import styles from './SortFilterSeriesPanel.module.scss'

const getClasses = (selected: SeriesSortOptionsInput, value: SeriesSortOptionsInput): string => {
    const isSelected = selected.mode === value.mode && selected.direction === value.direction
    return classNames({ [styles.selected]: isSelected, [styles.unselected]: !isSelected })
}

interface SortFilterSeriesPanelProps {
    selectedOption: SeriesSortOptionsInput
    limit: number
    seriesCount: number
    onChange: (parameter: SeriesDisplayOptionsInputRequired) => void
}

export const SortFilterSeriesPanel: React.FunctionComponent<SortFilterSeriesPanelProps> = ({
    selectedOption,
    limit,
    seriesCount: maxLimit,
    onChange,
}) => {
    const [selected, setSelected] = useState(selectedOption)
    const [seriesCount, setSeriesCount] = useState(Math.min(limit, maxLimit))

    const handleToggle = (value: SeriesSortOptionsInput): void => {
        setSelected(value)
        onChange({ limit: seriesCount, sortOptions: value })
    }

    const handleChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const count = Math.min(parseInt(event.target.value, 10) || 1, maxLimit)
        setSeriesCount(count)
        onChange({ limit: count, sortOptions: selected })
    }

    return (
        <section>
            <section className={classNames(styles.togglesContainer)}>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by result count</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton
                            selected={selectedOption}
                            value={{ mode: SeriesSortMode.RESULT_COUNT, direction: SeriesSortDirection.DESC }}
                            onClick={handleToggle}
                        >
                            Highest
                        </ToggleButton>
                        <ToggleButton
                            selected={selectedOption}
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
                            selected={selectedOption}
                            value={{ mode: SeriesSortMode.LEXICOGRAPHICAL, direction: SeriesSortDirection.ASC }}
                            onClick={handleToggle}
                        >
                            A-Z
                        </ToggleButton>
                        <ToggleButton
                            selected={selectedOption}
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
                            selected={selectedOption}
                            value={{ mode: SeriesSortMode.DATE_ADDED, direction: SeriesSortDirection.DESC }}
                            onClick={handleToggle}
                        >
                            Newest
                        </ToggleButton>
                        <ToggleButton
                            selected={selectedOption}
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
                    max={maxLimit}
                    value={seriesCount}
                    onChange={handleChange}
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

const ToggleButton: React.FunctionComponent<ToggleButtonProps> = ({ selected, value, children, onClick }) => (
    <Button variant="secondary" size="sm" className={getClasses(selected, value)} onClick={() => onClick(value)}>
        {children}
    </Button>
)
