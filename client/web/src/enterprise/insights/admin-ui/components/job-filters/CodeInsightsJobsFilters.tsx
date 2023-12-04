import { type ChangeEvent, useId, useState, type FC } from 'react'

import classNames from 'classnames'
import { identity, upperFirst } from 'lodash'

import { BackfillQueueOrderBy, InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations'
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
    Select,
} from '@sourcegraph/wildcard'

import styles from './CodeInsightsJobsFilters.module.scss'

const CODE_INSIGHTS_JOBS_FILTERS = [
    InsightQueueItemState.COMPLETED,
    InsightQueueItemState.FAILED,
    InsightQueueItemState.NEW,
    InsightQueueItemState.PROCESSING,
    InsightQueueItemState.QUEUED,
    InsightQueueItemState.UNKNOWN,
]

export const formatFilter = (filter: InsightQueueItemState): string => upperFirst(filter.toLowerCase())

interface CodeInsightsFiltersPickerProps {
    selectedFilters: InsightQueueItemState[]
    className?: string
    onFiltersChange: (filters: InsightQueueItemState[]) => void
}

export const CodeInsightsJobStatusPicker: FC<CodeInsightsFiltersPickerProps> = props => {
    const { selectedFilters, className, onFiltersChange } = props

    const filterInputId = useId()
    const [filterInput, setFilterInput] = useState<string>('')

    // Render only non-selected filters and filters that match with search term value
    const suggestions = CODE_INSIGHTS_JOBS_FILTERS.filter(
        filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(filterInput.toLowerCase())
    )

    return (
        <Label htmlFor={filterInputId} className={classNames(styles.statusFilter, className)}>
            <H4 as={H3} className="mb-0">
                Status:
            </H4>

            <MultiCombobox
                selectedItems={selectedFilters}
                getItemName={formatFilter}
                getItemKey={identity}
                onSelectedItemsChange={onFiltersChange}
                className={styles.statusFilterField}
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
                                <MultiComboboxOption
                                    key={filter.toString()}
                                    value={formatFilter(filter)}
                                    index={index}
                                />
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
    )
}

interface CodeInsightsOrderPickerProps {
    order: BackfillQueueOrderBy
    className?: string
    onOrderChange: (order: BackfillQueueOrderBy) => void
}

export const CodeInsightsJobsOrderPicker: FC<CodeInsightsOrderPickerProps> = props => {
    const { order, className, onOrderChange } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value as BackfillQueueOrderBy

        onOrderChange(nextValue)
    }

    return (
        <Select
            id={useId()}
            label="Order jobs by:"
            value={order}
            isCustomStyle={true}
            className={classNames(className, styles.orderBy)}
            labelClassName="flex-shrink-0"
            onChange={handleSelect}
        >
            <option value={BackfillQueueOrderBy.STATE}>State</option>
            <option value={BackfillQueueOrderBy.QUEUE_POSITION}>Queue position</option>
        </Select>
    )
}
