import * as React from 'react'
import FilterInput from './FilterInput'
import { QueryState } from '../../helpers'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'

interface Props {
    /**
     * The filters currently added to the query.
     */
    filtersInQuery: FiltersToTypeAndValue

    /**
     * The query in the main query input.
     */
    navbarQuery: QueryState

    /**
     * Callback to handle the filter's value being updated.
     */
    onFilterEdited: (filterKey: string, value: string) => void

    /**
     * Callback to handle a filter being deleted from the selected filter row.
     */
    onFilterDeleted: (filterKey: string) => void

    /**
     * Callback to handle the editable state of a filter.
     */
    toggleFilterEditable: (filterKey: string) => void

    /**
     * Whether we're on the search homepage.
     */
    isHomepage: boolean
}

/**
 * The row displaying the filters that have been added to the query in interactive mode.
 */
export const SelectedFiltersRow: React.FunctionComponent<Props> = ({
    filtersInQuery,
    navbarQuery,
    onFilterEdited,
    onFilterDeleted,
    toggleFilterEditable,
    isHomepage,
}) => {
    const fieldValueKeys = Array.from(Object.keys(filtersInQuery))
    return (
        <>
            {fieldValueKeys.length > 0 && (
                <div className={`selected-filters-row ${isHomepage ? 'selected-filters-row--homepage' : ''}`}>
                    {filtersInQuery &&
                        fieldValueKeys.map(field => (
                            /** Replace this with new input component, which can be an input when editable, and button when non-editable */
                            <FilterInput
                                key={field}
                                mapKey={field}
                                filterType={filtersInQuery[field].type}
                                value={filtersInQuery[field].value}
                                editable={filtersInQuery[field].editable}
                                filtersInQuery={filtersInQuery}
                                navbarQuery={navbarQuery}
                                onFilterDeleted={onFilterDeleted}
                                onFilterEdited={onFilterEdited}
                                toggleFilterEditable={toggleFilterEditable}
                            />
                        ))}
                </div>
            )}
        </>
    )
}
