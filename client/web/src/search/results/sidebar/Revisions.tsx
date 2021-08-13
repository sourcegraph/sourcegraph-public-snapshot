import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { GitRefType } from '@sourcegraph/shared/src/graphql/schema'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { useConnection } from '../../../components/FilteredConnection/hooks/useConnection'
import { SyntaxHighlightedSearchQuery } from '../../../components/SyntaxHighlightedSearchQuery'
import {
    SearchSidebarGitRefsResult,
    SearchSidebarGitRefsVariables,
    SearchSidebarGitRefFields,
} from '../../../graphql-operations'

import { FilterLink } from './FilterLink'
import styles from './SearchSidebarSection.module.scss'

const GIT_REVS_QUERY = gql`
    query SearchSidebarGitRefs($repo: String, $first: Int, $query: String, $type: GitRefType) {
        repository(name: $repo) {
            ... on Repository {
                __typename
                id
                gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
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
    const { connection, fetchMore, hasNextPage } = useConnection<
        SearchSidebarGitRefsResult,
        SearchSidebarGitRefsVariables,
        SearchSidebarGitRefFields
    >({
        query: GIT_REVS_QUERY,
        variables: {
            first: 10,
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

    if (!connection) {
        return (
            <div className={classNames('d-flex justify-content-center mt-4', styles.sidebarSectionNoResults)}>
                <LoadingSpinner className="icon-inline" />
            </div>
        )
    }

    if (connection?.error) {
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
            {hasNextPage || (connection.totalCount ?? 0) > 100 || connection.nodes.length > 100 ? (
                <p className={classNames('text-muted d-flex', styles.sidebarSectionFooter)}>
                    <small className="flex-1">
                        <span>
                            {connection?.nodes.length} of {connection?.totalCount} {pluralNoun}
                        </span>
                    </small>
                    {hasNextPage ? (
                        <button
                            type="button"
                            className={classNames('btn btn-link', styles.sidebarSectionButtonLink)}
                            onClick={fetchMore}
                        >
                            Show more
                        </button>
                    ) : null}
                </p>
            ) : null}
        </>
    )
}

const REVISION_TAB_KEY = 'SearchProduct.Sidebar.Revisions.Tab'

interface RevisionsProps {
    repoName: string
    onFilterClick: (filter: string, value: string) => void
    query: string
}

export const Revisions: React.FunctionComponent<RevisionsProps> = ({ repoName, onFilterClick, query }) => {
    const [selectedTab, setSelectedTab] = useLocalStorage(REVISION_TAB_KEY, 0)
    const onRevisionFilterClick = (value: string): void => onFilterClick('rev', value)
    return (
        <Tabs index={selectedTab} onChange={setSelectedTab}>
            <TabList className={styles.sidebarSectionTabsHeader}>
                <Tab>Branches</Tab>
                <Tab>Tags</Tab>
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

export const getRevisions = (props: Omit<RevisionsProps, 'query'>) => (query: string) => (
    <Revisions {...props} query={query} />
)
