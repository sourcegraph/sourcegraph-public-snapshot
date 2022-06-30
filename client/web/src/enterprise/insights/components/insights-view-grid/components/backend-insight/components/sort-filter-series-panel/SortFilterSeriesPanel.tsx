import classNames from 'classnames'

import { Button, ButtonGroup, Input } from '@sourcegraph/wildcard'

import { SeriesSortOptionsInput, SeriesSortDirection, SeriesSortMode } from '../../../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SERIES } from '../../../../../../core/backend/gql-backend/methods/get-backend-insight-data/deserializators'
import { SeriesDisplayOptionsInputRequired } from '../../../../../../core/types/insight/common'

import styles from './SortFilterSeriesPanel.module.scss'

interface SortFilterSeriesPanelProps {
    value: {
        limit: number
        sortOptions: SeriesSortOptionsInput
    }
    seriesCount: number
    onChange: (parameter: SeriesDisplayOptionsInputRequired) => void
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
        if (event.target.value.length === 0) {
            return
        }
        const count = Math.min(parseInt(event.target.value, 10), maxLimit)
        onChange({ ...value, limit: count })
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
                    max={maxLimit}
                    value={value.limit}
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

const ToggleButton: React.FunctionComponent<ToggleButtonProps> = ({ selected, value, children, onClick }) => {
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
