import * as React from 'react'
import { startCase } from 'lodash'
import { AddFilterDropdown } from './AddFilterDropdown'
import { FilterTypes } from '../../../../../shared/src/search/interactive/util'

interface RowProps {
    /**
     * Whether we're currently on the search homepage.
     */
    isHomepage: boolean
    /**
     * Callback that adds a new filter to the SelectedFilterRow when one of the buttons are clicked.
     * */
    onAddNewFilter: (filter: FilterTypes) => void
}

// Filters that are shown as buttons, and not in the dropdown menu.
export const defaultFilterTypes = [FilterTypes.repo, FilterTypes.file]

/**
 * The row containing the buttons to add new filters in interactive mode.
 * */
export const AddFilterRow: React.FunctionComponent<RowProps> = ({ isHomepage, onAddNewFilter }) => {
    const buildOnAddFilterHandler = (filterType: FilterTypes) => (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault()

        onAddNewFilter(filterType)
    }

    return (
        <div className={`add-filter-row ${isHomepage ? 'add-filter-row--homepage' : ''} e2e-add-filter-row`}>
            {defaultFilterTypes.map(filterType => (
                <button
                    key={filterType}
                    type="button"
                    className={`add-filter-row__button btn btn-outline-primary e2e-add-filter-button-${filterType}`}
                    onClick={buildOnAddFilterHandler(filterType)}
                >
                    + {startCase(filterType)} filter
                </button>
            ))}
            <AddFilterDropdown onAddNewFilter={onAddNewFilter} />
        </div>
    )
}
