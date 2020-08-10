import * as React from 'react'
import { startCase } from 'lodash'
import { AddFilterDropdown } from './AddFilterDropdown'
import { FilterType } from '../../../../../shared/src/search/interactive/util'

interface RowProps {
    /**
     * Whether we're currently on the search homepage.
     */
    isHomepage: boolean
    /**
     * Callback that adds a new filter to the SelectedFilterRow when one of the buttons are clicked.
     * */
    onAddNewFilter: (filter: FilterType) => void
}

// Filters that are shown as buttons, and not in the dropdown menu.
export const defaultFilterTypes = [FilterType.repo, FilterType.file]

/**
 * The row containing the buttons to add new filters in interactive mode.
 * */
export const AddFilterRow: React.FunctionComponent<RowProps> = ({ isHomepage, onAddNewFilter }) => {
    const buildOnAddFilterHandler = (filterType: FilterType) => (event: React.MouseEvent<HTMLButtonElement>) => {
        event.preventDefault()

        onAddNewFilter(filterType)
    }

    return (
        <div className={`add-filter-row ${isHomepage ? 'add-filter-row--homepage' : ''} test-add-filter-row`}>
            {defaultFilterTypes.map(filterType => (
                <button
                    key={filterType}
                    type="button"
                    className={`add-filter-row__button btn btn-outline-primary test-add-filter-button-${filterType}`}
                    onClick={buildOnAddFilterHandler(filterType)}
                >
                    + {startCase(filterType)} filter
                </button>
            ))}
            <AddFilterDropdown onAddNewFilter={onAddNewFilter} />
        </div>
    )
}
