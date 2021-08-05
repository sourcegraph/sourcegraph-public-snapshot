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
import { RevisionCommitsTab } from './RevisionsPopoverCommits'
import { RevisionReferencesTab } from './RevisionsPopoverReferences'

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

    useEffect(() => {
        eventLogger.logViewEvent('RevisionsPopover')
    }, [])

    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

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
                {tabs.map(tab => (
                    <TabPanel key={tab.id}>
                        {tab.type ? (
                            <RevisionReferencesTab
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                type={tab.type}
                                currentRev={props.currentRev}
                                getURLFromRevision={getURLFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repo}
                            />
                        ) : (
                            <RevisionCommitsTab
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                currentRev={props.currentRev}
                                getURLFromRevision={getURLFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repo}
                                currentCommitID={props.currentCommitID}
                            />
                        )}
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
}
