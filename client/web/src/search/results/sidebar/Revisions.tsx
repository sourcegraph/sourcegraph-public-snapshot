import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import React from 'react'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { GitRefType } from '@sourcegraph/shared/src/graphql/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { useConnection } from '../../../components/FilteredConnection/hooks/useConnection'
import { SyntaxHighlightedSearchQuery } from '../../../components/SyntaxHighlightedSearchQuery'
import {
    SearchSidebarGitRefsResult,
    SearchSidebarGitRefsVariables,
    SearchSidebarGitRefFields,
} from '../../../graphql-operations'
import { useTemporarySetting } from '../../../settings/temporary/useTemporarySetting'
import { QueryUpdate } from '../../../stores/navbarSearchQueryState'

import { FilterLink } from './FilterLink'
import styles from './SearchSidebarSection.module.scss'

const DEFAULT_FIRST = 10
export const GIT_REVS_QUERY = gql`
    query SearchSidebarGitRefs($repo: String, $first: Int, $query: String, $type: GitRefType) {
        repository(name: $repo) {
            ... on Repository {
                __typename
                id
                gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
                    __typename
                    nodes {
                        ...SearchSidebarGitRefFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        }
    }

    fragment SearchSidebarGitRefFields on GitRef {
        __typename
        id
        name
        displayName
    }
`

const revisionLabel = (value: string): React.ReactElement => <SyntaxHighlightedSearchQuery query={`rev:${value}`} />

interface RevisionListProps {
    repoName: string
    type: GitRefType
    onFilterClick: (value: string) => void
    pluralNoun: string
    query: string
}

const RevisionList: React.FunctionComponent<RevisionListProps> = ({
    repoName,
    type,
    onFilterClick,
    pluralNoun,
    query,
}) => {
    const { connection, fetchMore, hasNextPage, loading, error } = useConnection<
        SearchSidebarGitRefsResult,
        SearchSidebarGitRefsVariables,
        SearchSidebarGitRefFields
    >({
        query: GIT_REVS_QUERY,
        variables: {
            first: DEFAULT_FIRST,
            repo: repoName,
            query,
            type,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data?.repository?.gitRefs) {
                throw new Error('Unable to fetch repo revisions.')
            }
            return data?.repository?.gitRefs
        },
    })

    if (loading) {
        return (
            <div className={classNames('d-flex justify-content-center mt-4', styles.sidebarSectionNoResults)}>
                <LoadingSpinner />
            </div>
        )
    }

    if (error || !connection || connection.error) {
        return (
            <p className={classNames('text-muted', styles.sidebarSectionNoResults)}>
                <span className="text-muted">Unable to fetch repository revisions.</span>
            </p>
        )
    }

    if (connection.nodes.length === 0) {
        return (
            <p className={classNames('text-muted', styles.sidebarSectionNoResults)}>
                {query
                    ? `None of the ${pluralNoun} in this repository match this filter.`
                    : `This repository doesn't have any ${pluralNoun}.`}
            </p>
        )
    }

    return (
        <>
            <ul className={styles.sidebarSectionList}>
                {connection?.nodes.map(node => (
                    <FilterLink
                        key={node.name}
                        label={node.displayName}
                        value={node.name}
                        labelConverter={revisionLabel}
                        onFilterChosen={onFilterClick}
                    />
                ))}
            </ul>
            {(connection.totalCount ?? 0) > DEFAULT_FIRST ? (
                <p className={classNames('text-muted d-flex', styles.sidebarSectionFooter)}>
                    <small className="flex-1" data-testid="summary">
                        {connection?.nodes.length} of {connection?.totalCount} {pluralNoun}
                    </small>
                    {hasNextPage ? (
                        <Button className={styles.sidebarSectionButtonLink} onClick={fetchMore} variant="link">
                            Show more
                        </Button>
                    ) : null}
                </p>
            ) : null}
        </>
    )
}

export enum TabIndex {
    BRANCHES,
    TAGS,
}

export interface RevisionsProps {
    repoName: string
    onFilterClick: (updates: QueryUpdate[]) => void
    query: string
    /**
     * This property is only exposed for storybook tests.
     */
    _initialTab?: TabIndex
}

export const Revisions: React.FunctionComponent<RevisionsProps> = React.memo(
    ({ repoName, onFilterClick, query, _initialTab }) => {
        const [selectedTab, setSelectedTab] = useTemporarySetting('search.sidebar.revisions.tab')
        const onRevisionFilterClick = (value: string): void =>
            onFilterClick([
                { type: 'updateOrAppendFilter', field: FilterType.rev, value },
                { type: 'appendFilter', field: FilterType.repo, value: `^${repoName}$`, unique: true },
            ])
        return (
            <Tabs index={_initialTab ?? selectedTab ?? 0} onChange={setSelectedTab}>
                <TabList className={styles.sidebarSectionTabsHeader}>
                    <Tab index={TabIndex.BRANCHES}>Branches</Tab>
                    <Tab index={TabIndex.TAGS}>Tags</Tab>
                </TabList>
                <TabPanels>
                    <TabPanel>
                        <RevisionList
                            pluralNoun="branches"
                            repoName={repoName}
                            type={GitRefType.GIT_BRANCH}
                            onFilterClick={onRevisionFilterClick}
                            query={query}
                        />
                    </TabPanel>
                    <TabPanel>
                        <RevisionList
                            pluralNoun="tags"
                            repoName={repoName}
                            type={GitRefType.GIT_TAG}
                            onFilterClick={onRevisionFilterClick}
                            query={query}
                        />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        )
    }
)

export const getRevisions = (props: Omit<RevisionsProps, 'query'>) => (query: string) => (
    <Revisions {...props} query={query} />
)
