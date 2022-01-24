import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback } from 'react'
import { UncontrolledPopover } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { gql, useQuery } from '@sourcegraph/http-client'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button, LoadingSpinner, useLocalStorage } from '@sourcegraph/wildcard'

import { CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables } from '../graphql-operations'

import styles from './GlobalCodeIntel.module.scss'

const SHOW_COOL_CODEINTEL = localStorage.getItem('coolCodeIntel') !== null

export const GlobalCodeIntel: React.FunctionComponent<{
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}> = props =>
    SHOW_COOL_CODEINTEL ? (
        <ul className={classNames('nav', styles.globalCodeintel)}>
            <li className="nav-item">
                <CoolCodeIntelPopover {...props} />
            </li>
        </ul>
    ) : null

/** A button that toggles the visibility of the ExtensionDevTools element in a popover. */
export const CoolCodeIntelPopover = React.memo<{
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}>(props => (
    <>
        <Button id="extension-status-popover" className="text-decoration-none px-2" variant="link">
            <span className="text-muted">Cool Code Intel</span> <MenuUpIcon className="icon-inline" />
        </Button>
        <UncontrolledPopover
            placement="bottom"
            target="extension-status-popover"
            hideArrow={true}
            popperClassName="border-0 rounded-0"
        >
            <CoolCodeIntel {...props} />
        </UncontrolledPopover>
    </>
))

const CoolCodeIntel: React.FunctionComponent<{
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}> = props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs
            defaultIndex={tabIndex}
            className={classNames('card border-0 rounded-0', styles.coolCodeIntelStatus)}
            onChange={handleTabsChange}
        >
            <div className="tablist-wrapper w-100 align-items-center">
                <TabList>
                    {TABS.map(({ label, id }) => (
                        <Tab className="d-flex flex-1 justify-content-around" key={id} data-tab-content={id}>
                            {label}
                        </Tab>
                    ))}
                </TabList>
            </div>

            <TabPanels>
                {TABS.map(tab => (
                    <TabPanel key={tab.id}>
                        <tab.component hoveredToken={props.hoveredToken} />
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
}

export interface CoolCodeIntelPopoverTabProps {
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references'

interface CoolCodeIntelToolsTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<CoolCodeIntelPopoverTabProps>
}

const FETCH_REFERENCES_QUERY = gql`
    query CoolCodeIntelReferences(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $after: String
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        references(line: $line, character: $character, after: $after) {
                            nodes {
                                resource {
                                    path
                                    repository {
                                        name
                                    }
                                    commit {
                                        oid
                                    }
                                }
                                range {
                                    start {
                                        line
                                        character
                                    }
                                    end {
                                        line
                                        character
                                    }
                                }
                            }
                            pageInfo {
                                endCursor
                            }
                        }
                    }
                }
            }
        }
    }
`
export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelPopoverTabProps> = props => (
    <>
        <div className="card-header">
            References{' '}
            <small>
                Check out this <i>intelligence</i>
            </small>
        </div>
        <div className="card-body border-top">
            {props.hoveredToken && (
                <>
                    <h4>
                        <b>Token under cursor</b>
                    </h4>
                    <p>Line: {props.hoveredToken.line}</p>
                    <p>Character: {props.hoveredToken.character}</p>
                    <p>RepoName: {props.hoveredToken.repoName}</p>
                    <p>CommitID: {props.hoveredToken.commitID}</p>
                    <p>FilePath: {props.hoveredToken.filePath}</p>
                </>
            )}
            {props.hoveredToken && <ReferencesList hoveredToken={props.hoveredToken} />}
        </div>
    </>
)

export const ReferencesList: React.FunctionComponent<{
    hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}> = props => {
    const { data, error, loading, refetch } = useQuery<CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables>(
        FETCH_REFERENCES_QUERY,
        {
            variables: {
                repository: props.hoveredToken.repoName,
                commit: props.hoveredToken.commitID,
                path: props.hoveredToken.filePath,
                // ATTENTION: Off by one ahead!!!!
                line: props.hoveredToken.line - 1,
                character: props.hoveredToken.character - 1,
                after: null,
            },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            // pollInterval: 5000,
            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <>
                <LoadingSpinner className="mx-auto my-4" />
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !data) {
        throw new Error(error.message)
    }
    // If there weren't any errors and we just didn't receive any data
    if (!data || !data.repository) {
        return <>Nothing found</>
    }

    const references = data.repository.commit?.blob?.lsif?.references.nodes

    console.log(references)

    return (
        <>
            <ul>
                {references?.map((reference, index) => (
                    <li key={index}>{reference.resource.path}</li>
                ))}
            </ul>
        </>
    )
}

const TABS: CoolCodeIntelToolsTab[] = [{ id: 'references', label: 'References', component: ReferencesPanel }]
