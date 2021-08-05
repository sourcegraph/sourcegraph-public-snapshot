import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { CircleChevronLeftIcon } from '@sourcegraph/shared/src/components/icons'
import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'
import { ConnectionContainer, ConnectionForm } from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../components/FilteredConnection'
import {
    GitCommitAncestorFields,
    GitCommitAncestorsConnectionFields,
    GitRefConnectionFields,
    GitRefFields,
    RepositoryGitCommitResult,
    RepositoryGitCommitVariables,
} from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'
import { replaceRevisionInURL } from '../util/url'

import { GitReferenceNode, queryGitReferences } from './GitReference'
import { ReferencesRevisionsList } from './RevisionsPopoverLists'

const fetchRepositoryCommits = memoizeObservable(
    (
        args: RevisionSpec & { repo: Scalars['ID']; first?: number; query?: string }
    ): Observable<GitCommitAncestorsConnectionFields> =>
        requestGraphQL<RepositoryGitCommitResult, RepositoryGitCommitVariables>(
            gql`
                query RepositoryGitCommit($repo: ID!, $first: Int, $revision: String!, $query: String) {
                    node(id: $repo) {
                        __typename
                        ... on Repository {
                            commit(rev: $revision) {
                                ancestors(first: $first, query: $query) {
                                    ...GitCommitAncestorsConnectionFields
                                }
                            }
                        }
                    }
                }

                fragment GitCommitAncestorsConnectionFields on GitCommitConnection {
                    nodes {
                        ...GitCommitAncestorFields
                    }
                    pageInfo {
                        hasNextPage
                    }
                }

                fragment GitCommitAncestorFields on GitCommit {
                    id
                    oid
                    abbreviatedOID
                    author {
                        person {
                            name
                            avatarURL
                        }
                        date
                    }
                    subject
                }
            `,
            {
                first: args.first ?? null,
                query: args.query ?? null,
                repo: args.repo,
                revision: args.revision,
            }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error(`Repository ${args.repo} not found`)
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`Node is a ${node.__typename}, not a Repository`)
                }
                if (!node.commit?.ancestors) {
                    throw new Error(`Cannot load ancestors for repository ${args.repo}`)
                }
                return node.commit.ancestors
            })
        ),
    args => JSON.stringify(args)
)

interface GitReferencePopoverNodeProps {
    node: GitRefFields

    defaultBranch: string
    currentRevision: string | undefined

    location: H.Location

    getURLFromRevision: (href: string, revision: string) => string
}

export const GitReferencePopoverNode: React.FunctionComponent<GitReferencePopoverNodeProps> = ({
    node,
    defaultBranch,
    currentRevision,
    location,
    getURLFromRevision,
}) => {
    let isCurrent: boolean
    if (currentRevision) {
        isCurrent = node.name === currentRevision || node.abbrevName === currentRevision
    } else {
        isCurrent = node.name === `refs/heads/${defaultBranch}`
    }
    return (
        <GitReferenceNode
            node={node}
            url={getURLFromRevision(location.pathname + location.search + location.hash, node.abbrevName)}
            ancestorIsLink={false}
            className={classNames(
                'connection-popover__node-link',
                isCurrent && 'connection-popover__node-link--active'
            )}
        >
            {isCurrent && (
                <CircleChevronLeftIcon
                    className="icon-inline connection-popover__node-link-icon"
                    data-tooltip="Current"
                />
            )}
        </GitReferenceNode>
    )
}

interface GitCommitNodeProps {
    node: GitCommitAncestorFields

    currentCommitID: string | undefined

    location: H.Location

    getURLFromRevision: (href: string, revision: string) => string
}

const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({
    node,
    currentCommitID,
    location,
    getURLFromRevision,
}) => {
    const isCurrent = currentCommitID === node.oid
    return (
        <li key={node.oid} className="connection-popover__node revisions-popover-git-commit-node">
            <Link
                to={getURLFromRevision(location.pathname + location.search + location.hash, node.oid)}
                className={classNames(
                    'connection-popover__node-link',
                    isCurrent && 'connection-popover__node-link--active',
                    'revisions-popover-git-commit-node__link'
                )}
            >
                <code className="revisions-popover-git-commit-node__oid" title={node.oid}>
                    {node.abbreviatedOID}
                </code>
                {isCurrent && (
                    <CircleChevronLeftIcon
                        className="icon-inline connection-popover__node-link-icon"
                        data-tooltip="Current commit"
                    />
                )}
            </Link>
        </li>
    )
}

interface RevisionsPopoverProps {
    repo: Scalars['ID']
    repoName: string
    defaultBranch: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    history: H.History
    location: H.Location

    /* Callback to dismiss the parent popover wrapper */
    togglePopover: () => void

    getURLFromRevision?: (href: string, revision: string) => string

    tabs?: RevisionsPopoverTab[]
}

type RevisionsPopoverTabID = 'branches' | 'tags' | 'commits'

interface RevisionsPopoverTab {
    id: RevisionsPopoverTabID
    label: string
    noun: string
    pluralNoun: string
    type?: GitRefType
}

const LAST_TAB_STORAGE_KEY = 'RevisionsPopover.lastTab'

export const BRANCHES_TAB: RevisionsPopoverTab = {
    id: 'branches',
    label: 'Branches',
    noun: 'branch',
    pluralNoun: 'branches',
    type: GitRefType.GIT_BRANCH,
}
export const TAGS_TAB: RevisionsPopoverTab = {
    id: 'tags',
    label: 'Tags',
    noun: 'tag',
    pluralNoun: 'tags',
    type: GitRefType.GIT_TAG,
}
export const COMMITS_TAB: RevisionsPopoverTab = {
    id: 'commits',
    label: 'Commits',
    noun: 'commit',
    pluralNoun: 'commits',
}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const RevisionsPopover: React.FunctionComponent<RevisionsPopoverProps> = props => {
    const { getURLFromRevision = replaceRevisionInURL, tabs = [BRANCHES_TAB, TAGS_TAB, COMMITS_TAB] } = props
    const [searchValue, setSearchValue] = useState('')
    const debouncedSearchValue = useDebounce(searchValue, 200)

    useEffect(() => {
        eventLogger.logViewEvent('RevisionsPopover')
    }, [])

    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    const queryGitBranches = (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> =>
        queryGitReferences({ ...args, repo: props.repo, type: GitRefType.GIT_BRANCH, withBehindAhead: false })

    const queryGitTags = (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> =>
        queryGitReferences({ ...args, repo: props.repo, type: GitRefType.GIT_TAG, withBehindAhead: false })

    const queryRepositoryCommits = (
        args: FilteredConnectionQueryArguments
    ): Observable<GitCommitAncestorsConnectionFields> =>
        fetchRepositoryCommits({
            ...args,
            repo: props.repo,
            revision: props.currentRev || props.defaultBranch,
        })

    return (
        <Tabs defaultIndex={tabIndex} className="revisions-popover connection-popover" onChange={handleTabsChange}>
            <div className="tablist-wrapper revisions-popover__tabs">
                <TabList>
                    {tabs.map(({ label, id }) => (
                        <Tab key={id} data-tab-content={id}>
                            <span className="tablist-wrapper--tab-label">{label}</span>
                        </Tab>
                    ))}
                </TabList>
                <button
                    onClick={props.togglePopover}
                    type="button"
                    className="btn btn-icon revisions-popover__tabs-close"
                    aria-label="Close"
                >
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
            <TabPanels>
                <ConnectionContainer compact={true} className="connection-popover__content">
                    <ConnectionForm
                        inputValue={searchValue}
                        onInputChange={event => setSearchValue(event.target.value)}
                        autoFocus={true}
                        inputPlaceholder="Find..."
                        inputClassName="connection-popover__input"
                    />
                    {tabs.map(tab => (
                        <TabPanel key={tab.id}>
                            {tab.type ? (
                                <ReferencesRevisionsList
                                    query={debouncedSearchValue}
                                    noun={tab.noun}
                                    pluralNoun={tab.pluralNoun}
                                    type={tab.type}
                                    currentRev={props.currentRev}
                                    getURLFromRevision={getURLFromRevision}
                                    defaultBranch={props.defaultBranch}
                                    repo={props.repo}
                                />
                            ) : null}
                        </TabPanel>
                    ))}
                </ConnectionContainer>
            </TabPanels>
        </Tabs>
    )
}
