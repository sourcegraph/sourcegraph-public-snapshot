import React, { useCallback } from 'react'
import { FilterTypes, filterTypeKeys } from '../../../../../shared/src/search/interactive/util'
import { defaultFilterTypes } from './AddFilterRow'
import { FilterTypesToProseNames } from './filters'

interface Props {
    onAddNewFilter: (filterType: FilterTypes) => void
}

export const AddFilterDropdown: React.FunctionComponent<Props> = ({ onAddNewFilter }) => {
    const addNewFilter = useCallback(
        (e: React.ChangeEvent<HTMLSelectElement>): void => {
            onAddNewFilter(e.target.value as FilterTypes)
        },
        [onAddNewFilter]
    )

    return (
        <select
            className="form-control add-filter-dropdown e2e-filter-dropdown"
            onChange={addNewFilter}
            value="default"
        >
            <option value="default" disabled={true}>
                Add filter…
            </option>
            {filterTypeKeys
                .filter(filter => !defaultFilterTypes.includes(filter) && filter !== FilterTypes.case)
                .map(filter => (
                    <option key={filter} value={filter} className={`e2e-filter-dropdown-option-${filter}`}>
                        {FilterTypesToProseNames[filter]}
                    </option>
                ))}
        </select>
    )
}
