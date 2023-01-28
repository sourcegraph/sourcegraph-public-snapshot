import { ChangeEvent, FC, ReactElement, useId, useState } from 'react';

import classNames from 'classnames';
import { upperFirst, identity } from 'lodash';

import { BackfillQueueOrderBy, InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations';
import {
    H3,
    H4,
    Label,
    Input,
    Select,
    Button,
    MultiCombobox,
    MultiComboboxEmptyList,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxPopover,
    PageHeader
} from '@sourcegraph/wildcard';

import { PageTitle } from '../../../components/PageTitle';

import styles from './CodeInsightsJobs.module.scss'

export const CodeInsightsJobs: FC = props => {

    const [orderBy, setOrderBy] = useState<BackfillQueueOrderBy>(BackfillQueueOrderBy.STATE)
    const [selectedFilters, setFilters] = useState<InsightQueueItemState[]>([])
    const [search, setSearch] = useState<string>('')

    return (
        <div>
            <PageTitle title="Code Insights jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code Insights jobs' }]}
                description="List of actionable Code Insights queued jobs"
                className="mb-3"
            />

            <header className={styles.header}>

                <CodeInsightsJobActions/>

                <CodeInsightsFiltersPicker
                    selectedFilters={selectedFilters}
                    className={styles.statusFilter}
                    onFiltersChange={setFilters}
                />

                <CodeInsightsOrderPicker
                    order={orderBy}
                    className={styles.orderBy}
                    onOrderChange={setOrderBy}
                />

                <Input
                    placeholder='Search jobs by title or series label'
                    value={search}
                    className={styles.search}
                    onChange={event => setSearch(event.target.value)}
                />
            </header>
        </div>
    )
}

interface CodeInsightsJobActionsProps {
    className?: string
}

function CodeInsightsJobActions(props: CodeInsightsJobActionsProps): ReactElement {
    const { className } = props

    return (
        <div className={classNames(className, styles.actions)}>
            <Button variant='primary'>Retry</Button>
            <Button variant='primary'>Front of queue</Button>
            <Button variant='primary'>Back of queue</Button>
        </div>
    )
}

interface CodeInsightsOrderPickerProps {
    order: BackfillQueueOrderBy
    className?: string
    onOrderChange: (order: BackfillQueueOrderBy) => void
}

function CodeInsightsOrderPicker(props: CodeInsightsOrderPickerProps): ReactElement {
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
            className={classNames(className, 'm-0')}
            labelClassName='flex-shrink-0'
            onChange={handleSelect}
        >
            <option value={BackfillQueueOrderBy.STATE}>State</option>
            <option value={BackfillQueueOrderBy.QUEUE_POSITION}>Queue position</option>
        </Select>
    )
}

const CODE_INSIGHTS_JOBS_FILTERS = [
    InsightQueueItemState.COMPLETED,
    InsightQueueItemState.FAILED,
    InsightQueueItemState.NEW,
    InsightQueueItemState.PROCESSING,
    InsightQueueItemState.QUEUED,
    InsightQueueItemState.UNKNOWN,
]

const formatFilter = (filter: InsightQueueItemState): string => upperFirst(filter.toLowerCase())

interface CodeInsightsFiltersPickerProps {
    selectedFilters: InsightQueueItemState[]
    className?: string
    onFiltersChange: (filters: InsightQueueItemState[]) => void
}

function CodeInsightsFiltersPicker(props: CodeInsightsFiltersPickerProps): ReactElement {
    const { selectedFilters, className, onFiltersChange } = props

    const filterInputId = useId()
    const [filterInput, setFilterInput] = useState<string>('')

    // Render only non-selected filters and filters that match with search term value
    const suggestions = CODE_INSIGHTS_JOBS_FILTERS.filter(
        filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(filterInput.toLowerCase())
    )

    return (
        <Label htmlFor={filterInputId} className={className}>
            <H4 as={H3} className="mb-0">
                Status:
            </H4>

            <MultiCombobox
                selectedItems={selectedFilters}
                getItemName={formatFilter}
                getItemKey={identity}
                onSelectedItemsChange={onFiltersChange}
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
    )
}
