import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import React from 'react'
import { from, Observable } from 'rxjs'
import { useHistory, useLocation } from 'react-router'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { getDocumentNode, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { GitRefType } from '@sourcegraph/shared/src/graphql/schema'

import {
    SearchSidebarGitRefsResult,
    SearchSidebarGitRefsVariables,
    SearchSidebarGitRefsConnectionFields,
    SearchSidebarGitRefFields,
} from '../../../graphql-operations'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { SyntaxHighlightedSearchQuery } from '../../../components/SyntaxHighlightedSearchQuery'
import { client } from '../../../backend/graphql'

import styles from './SearchSidebarSection.module.scss'
import { FilterLink } from './FilterLink'

const queryGitBranches = memoizeObservable(
    (args: {
        repo: string
        type: GitRefType
        first?: number
        query?: string
    }): Observable<SearchSidebarGitRefsConnectionFields> =>
        from(
            client.query<SearchSidebarGitRefsResult, SearchSidebarGitRefsVariables>({
                query: getDocumentNode(gql`
                    query SearchSidebarGitRefs($repo: String, $first: Int, $query: String, $type: GitRefType) {
                        repository(name: $repo) {
                            ... on Repository {
                                __typename
                                id
                                gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
                                    ...SearchSidebarGitRefsConnectionFields
                                }
                            }
                        }
                    }

                    fragment SearchSidebarGitRefsConnectionFields on GitRefConnection {
                        nodes {
                            ...SearchSidebarGitRefFields
                        }
                        totalCount
                        pageInfo {
                            hasNextPage
                        }
                    }

                    fragment SearchSidebarGitRefFields on GitRef {
                        __typename
                        id
                        name
                        displayName
                    }
                `),
                variables: {
                    query: args.query ?? null,
                    first: args.first ?? null,
                    repo: args.repo,
                    type: args.type,
                },
            })
        ).pipe(
            map(({ data, errors }) => {
                if (!data?.repository?.gitRefs) {
                    throw createAggregateError(errors)
                }
                return data.repository.gitRefs
            })
        ),
    args => `${args.repo}:${String(args.first)}:${String(args.query)}:${String(args.type)}`
)

const revLabel = (value: string) => <SyntaxHighlightedSearchQuery query={`rev:${value}`} />

interface RevisionListProps {
    repoName: string
    type: GitRefType
    onFilterClick: (value: string) => void
    noun: string
    pluralNoun: string
}

const RevisionList: React.FunctionComponent<RevisionListProps> = ({
    repoName,
    type,
    onFilterClick,
    noun,
    pluralNoun,
}) => {
    const history = useHistory()
    const location = useLocation()
    const query = (args: FilteredConnectionQueryArguments): Observable<SearchSidebarGitRefsConnectionFields> =>
        queryGitBranches({ ...args, repo: repoName, type })

    return (
        <FilteredConnection<SearchSidebarGitRefFields, any>
            history={history}
            location={location}
            defaultFirst={10}
            compact={true}
            inputClassName="form-control-sm"
            noun={noun}
            pluralNoun={pluralNoun}
            queryConnection={query}
            nodeComponent={({ node, onFilterClick }) => {
                return (
                    <FilterLink
                        label={node.displayName}
                        value={node.name}
                        labelConverter={revLabel}
                        onFilterChosen={onFilterClick}
                    />
                )
            }}
            nodeComponentProps={{
                onFilterClick,
            }}
            useURLQuery={false}
        />
    )
}

const REVISION_TAB_KEY = 'SearchProduct.Sidebar.Revisions.Tab'

interface RevisionsProps {
    repoName: string
    onFilterClick: (filter: string, value: string) => void
}

export const Revisions: React.FunctionComponent<RevisionsProps> = ({ repoName, onFilterClick }) => {
    const [selectedTab, setSelectedTab] = useLocalStorage(REVISION_TAB_KEY, 0)
    const onRevFilterClick = (value: string) => onFilterClick('rev', value)
    return (
        <Tabs index={selectedTab} onChange={setSelectedTab}>
            <TabList className={styles.sidebarSectionTabsHeader}>
                <Tab>Branches</Tab>
                <Tab>Tags</Tab>
            </TabList>
            <TabPanels>
                <TabPanel>
                    <RevisionList
                        noun="branche"
                        pluralNoun="branches"
                        repoName={repoName}
                        type={GitRefType.GIT_BRANCH}
                        onFilterClick={onRevFilterClick}
                    />
                </TabPanel>
                <TabPanel>
                    <RevisionList
                        noun="tag"
                        pluralNoun="tags"
                        repoName={repoName}
                        type={GitRefType.GIT_TAG}
                        onFilterClick={onRevFilterClick}
                    />
                </TabPanel>
            </TabPanels>
        </Tabs>
    )
}
