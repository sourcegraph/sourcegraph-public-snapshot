import { useState } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup } from '@sourcegraph/wildcard'

import styles from './SortFilterSeriesPanel.module.scss'

const getClasses = (selected: boolean): string =>
    classNames({ [styles.selected]: selected, [styles.unselected]: !selected })

export enum SortSeriesBy {
    CountAsc = 'CountAsc',
    CountDesc = 'CountDesc',
    AlphaAsc = 'AlphaAsc',
    AlphaDesc = 'AlphaDesc',
    DateAsc = 'DateAsc',
    DateDesc = 'DateDesc',
}

export interface SortFilterSeriesValue {
    selected: SortSeriesBy
    seriesCount: number
}

interface SortFilterSeriesPanelProps {
    value: SortFilterSeriesValue
    onChange: (parameter: SortFilterSeriesValue) => void
}

export const SortFilterSeriesPanel: React.FunctionComponent<SortFilterSeriesPanelProps> = ({ value, onChange }) => {
    const [selected, setSelected] = useState(value.selected)
    const [seriesCount, setSeriesCount] = useState(value.seriesCount)

    const handleToggle = (value: SortSeriesBy): void => {
        setSelected(value)
        onChange({ selected: value, seriesCount })
    }

    const handleChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const count = parseInt(event.target.value, 10)
        setSeriesCount(count)
        onChange({ selected, seriesCount: count })
    }

    return (
        <section>
            <section className={classNames(styles.togglesContainer)}>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by result count</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton selected={selected} value={SortSeriesBy.CountDesc} onClick={handleToggle}>
                            Highest
                        </ToggleButton>
                        <ToggleButton selected={selected} value={SortSeriesBy.CountAsc} onClick={handleToggle}>
                            Lowest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by name</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton selected={selected} value={SortSeriesBy.AlphaAsc} onClick={handleToggle}>
                            A-Z
                        </ToggleButton>
                        <ToggleButton selected={selected} value={SortSeriesBy.AlphaDesc} onClick={handleToggle}>
                            Z-A
                        </ToggleButton>
                    </ButtonGroup>
                </div>
                <div className="d-flex flex-column">
                    <small className={styles.label}>Sort by date added</small>
                    <ButtonGroup className={styles.toggleGroup}>
                        <ToggleButton selected={selected} value={SortSeriesBy.DateDesc} onClick={handleToggle}>
                            Latest
                        </ToggleButton>
                        <ToggleButton selected={selected} value={SortSeriesBy.DateAsc} onClick={handleToggle}>
                            Oldest
                        </ToggleButton>
                    </ButtonGroup>
                </div>
            </section>
            <footer className={styles.footer}>
                <span>Number of data series</span>
                <input
                    type="number"
                    step="1"
                    value={seriesCount}
                    className="form-control form-control-sm"
                    onChange={handleChange}
                />
            </footer>
        </section>
    )
}

interface ToggleButtonProps {
    selected: SortSeriesBy
    value: SortSeriesBy
    onClick: (value: SortSeriesBy) => void
}

const ToggleButton: React.FunctionComponent<ToggleButtonProps> = ({ selected, value, children, onClick }) => (
    <Button variant="secondary" size="sm" className={getClasses(selected === value)} onClick={() => onClick(value)}>
        {children}
    </Button>
)
