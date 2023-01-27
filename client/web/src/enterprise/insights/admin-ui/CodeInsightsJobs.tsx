import { FC, useId, useState } from 'react';

import { upperFirst, identity } from 'lodash';

import { InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations';
import {
    H3,
    H4,
    Label,
    MultiCombobox,
    MultiComboboxEmptyList,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxPopover,
    PageHeader
} from '@sourcegraph/wildcard';

import { PageTitle } from '../../../components/PageTitle';

const FILTERS = [
    InsightQueueItemState.COMPLETED,
    InsightQueueItemState.ERRORED,
    InsightQueueItemState.FAILED,
    InsightQueueItemState.NEW,
    InsightQueueItemState.PROCESSING,
    InsightQueueItemState.QUEUED
]
const formatFilter = (filter: InsightQueueItemState): string => upperFirst(filter.toLowerCase())

export const CodeInsightsJobs: FC = props => {
    const filterInputId = useId()
    const [filterInput, setFilterInput] = useState<string>('')
    const [selectedFilters, setFilters] = useState<InsightQueueItemState[]>([])

    // Render only non-selected filters and filters that match with search term value
    const suggestions = FILTERS.filter(
        filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(filterInput.toLowerCase())
    )

    return (
        <div>
            <PageTitle title="Code Insights jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code Insights jobs' }]}
                description="List of actionable Code Insights queued jobs"
                className="mb-3"
            />

            <header>
                <Label htmlFor={filterInputId}>
                    <H4 as={H3} className="mb-0 mr-2">
                        Status
                    </H4>

                    <MultiCombobox
                        selectedItems={selectedFilters}
                        getItemName={formatFilter}
                        getItemKey={identity}
                        onSelectedItemsChange={setFilters}
                    >
                        <MultiComboboxInput
                            id={filterInputId}
                            autoCorrect="false"
                            autoComplete="off"
                            placeholder="Select filter..."
                            value={filterInput}
                            onChange={event => setFilterInput(event.target.value)}
                        />

                        <MultiComboboxPopover>
                            <MultiComboboxList items={suggestions}>
                                {filters =>
                                    filters.map((filter, index) => (
                                        <MultiComboboxOption key={filter.toString()} value={formatFilter(filter)} index={index} />
                                    ))
                                }
                            </MultiComboboxList>

                            {suggestions.length === 0 && (
                                <MultiComboboxEmptyList>
                                    {!filterInput ? <>All filters are selected</> : <>No options</>}
                                </MultiComboboxEmptyList>
                            )}
                        </MultiComboboxPopover>
                    </MultiCombobox>
                </Label>
            </header>
        </div>
    )
}

