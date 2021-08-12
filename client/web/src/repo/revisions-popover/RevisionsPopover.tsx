import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect } from 'react'

import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { eventLogger } from '../../tracking/eventLogger'
import { replaceRevisionInURL } from '../../util/url'

import { RevisionCommitsTab } from './RevisionsPopoverCommits'
import { RevisionReferencesTab } from './RevisionsPopoverReferences'

export interface RevisionsPopoverProps {
    repo: Scalars['ID']
    repoName: string
    defaultBranch: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    /* Callback to dismiss the parent popover wrapper */
    togglePopover: () => void

    getURLFromRevision?: (href: string, revision: string) => string

    allowSpeculativeSearch?: boolean
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

const TABS: RevisionsPopoverTab[] = [
    { id: 'branches', label: 'Branches', noun: 'branch', pluralNoun: 'branches', type: GitRefType.GIT_BRANCH },
    { id: 'tags', label: 'Tags', noun: 'tag', pluralNoun: 'tags', type: GitRefType.GIT_TAG },
    // { id: 'commits', label: 'Commits', noun: 'commit', pluralNoun: 'commits' },
]

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const RevisionsPopover: React.FunctionComponent<RevisionsPopoverProps> = props => {
    const { getURLFromRevision = replaceRevisionInURL } = props

    useEffect(() => {
        eventLogger.logViewEvent('RevisionsPopover')
    }, [])

    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs defaultIndex={tabIndex} className="revisions-popover connection-popover" onChange={handleTabsChange}>
            <div className="tablist-wrapper revisions-popover__tabs">
                <TabList>
                    {TABS.map(({ label, id }) => (
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
                {TABS.map(tab => (
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
                                repoName={props.repoName}
                                allowSpeculativeSearch={
                                    props.allowSpeculativeSearch && tab.type === GitRefType.GIT_BRANCH
                                }
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
