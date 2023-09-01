import { FC, useState } from 'react'

import { upperFirst } from 'lodash'
import FeatureSearchOutlineIcon from 'mdi-react/MapSearchOutlineIcon'

import { SearchJobsOrderBy, SearchJobState } from '@sourcegraph/shared/src/graphql-operations'
import {
    PageHeader,
    Link,
    Container,
    Input,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
    Select,
    FeedbackBadge,
} from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'

import { UsersPicker, User } from './UsersPicker'

import styles from './SearchJobsPage.module.scss'

const SEARCH_JOB_STATES = [
    SearchJobState.COMPLETED,
    SearchJobState.ERRORED,
    SearchJobState.FAILED,
    SearchJobState.QUEUED,
    SearchJobState.PROCESSING,
]

export const SearchJobsPage: FC = props => {
    const [searchStateTerm, setSearchStateTerm] = useState('')
    const [selectedStates, setStates] = useState<SearchJobState[]>([])
    const [selectedUsers, setUsers] = useState<User[]>([])
    const [sortBy, setSortBy] = useState<SearchJobsOrderBy>(SearchJobsOrderBy.CREATED_DATE)

    // Render only non-selected filters and filters that match with search term value
    const suggestions = SEARCH_JOB_STATES.filter(
        filter => !selectedStates.includes(filter) && filter.toLowerCase().includes(searchStateTerm.toLowerCase())
    )

    return (
        <Page>
            <PageHeader
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                path={[{ icon: FeatureSearchOutlineIcon, text: 'Search Jobs' }]}
                description={
                    <>
                        Run search queries over all repositories, branches, commit and revisions.{' '}
                        <Link to="">Learn more</Link> about search jobs.
                    </>
                }
            />

            <Container className="mt-4">
                <header className={styles.header}>
                    <Input
                        placeholder="Search jobs by query..."
                        className={styles.search}
                        inputClassName={styles.searchInput}
                    />

                    <MultiCombobox
                        selectedItems={selectedStates}
                        getItemKey={formatJobState}
                        getItemName={formatJobState}
                        onSelectedItemsChange={setStates}
                        className={styles.filters}
                    >
                        <MultiComboboxInput
                            placeholder="Filter by search status..."
                            value={searchStateTerm}
                            autoCorrect="false"
                            autoComplete="off"
                            onChange={event => setSearchStateTerm(event.target.value)}
                        />

                        <MultiComboboxPopover>
                            <MultiComboboxList items={suggestions}>
                                {items =>
                                    items.map((item, index) => (
                                        <MultiComboboxOption
                                            key={formatJobState(item)}
                                            value={formatJobState(item)}
                                            index={index}
                                        />
                                    ))
                                }
                            </MultiComboboxList>
                        </MultiComboboxPopover>
                    </MultiCombobox>

                    <UsersPicker value={selectedUsers} onChange={setUsers} />

                    <Select
                        aria-label="Filter by search job status"
                        value={sortBy}
                        onChange={event => setSortBy(event.target.value as SearchJobsOrderBy)}
                        isCustomStyle={true}
                        className={styles.sort}
                        selectClassName={styles.sortSelect}
                    >
                        <option value={SearchJobsOrderBy.CREATED_DATE}>Sort by Created date</option>
                        <option value={SearchJobsOrderBy.QUERY}>Sort by Query</option>
                        <option value={SearchJobsOrderBy.STATE}>Sort by Status</option>
                    </Select>
                </header>
            </Container>
        </Page>
    )
}

const formatJobState = (state: SearchJobState): string => upperFirst(state.toLowerCase())
