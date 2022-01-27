import classNames from 'classnames'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useState } from 'react'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { useQuery } from '@sourcegraph/http-client'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import {
    RepoSpec,
    RevisionSpec,
    FileSpec,
    ResolvedRevisionSpec,
    toPositionOrRangeQueryParameter,
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
} from '@sourcegraph/shared/src/util/url'
import { Tab, TabList, TabPanel, TabPanels, Tabs, Link, LoadingSpinner, useLocalStorage } from '@sourcegraph/wildcard'

import { CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables, Maybe } from '../graphql-operations'
import { BlobProps } from '../repo/blob/Blob'

import styles from './GlobalCodeIntel.module.scss'
import { FETCH_REFERENCES_QUERY } from './GlobalCodeIntelQueries'

const SHOW_COOL_CODEINTEL = localStorage.getItem('coolCodeIntel') !== null

export const GlobalCodeIntel: React.FunctionComponent<
    {
        hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
        showPanel: boolean
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo'>
> = props => {
    if (!SHOW_COOL_CODEINTEL) {
        return null
    }

    if (props.showPanel) {
        return <CoolCodeIntelResizablePanel {...props} />
    }

    return null
}

export interface CoolCodeIntelPopoverTabProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo'> {
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references' | 'token' | 'definition'

interface CoolCodeIntelToolsTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<CoolCodeIntelPopoverTabProps>
}

export const TokenPanel: React.FunctionComponent<CoolCodeIntelPopoverTabProps> = props => (
    <>
        {props.hoveredToken ? (
            <code>
                Line: {props.hoveredToken.line}
                {'\n'}
                Character: {props.hoveredToken.character}
                {'\n'}
                Repo: {props.hoveredToken.repoName}
                {'\n'}
                Commit: {props.hoveredToken.commitID}
                {'\n'}
                Path: {props.hoveredToken.filePath}
                {'\n'}
            </code>
        ) : (
            <p>
                <i>No token</i>
            </p>
        )}
    </>
)

export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelPopoverTabProps> = props => (
    <div>{props.hoveredToken && <ReferencesList hoveredToken={props.hoveredToken} {...props} />}</div>
)

interface Location {
    __typename?: 'Location' | undefined
    resource: {
        __typename?: 'GitBlob' | undefined
        path: string
        content: string
        repository: {
            __typename?: 'Repository' | undefined
            name: string
        }
        commit: {
            __typename?: 'GitCommit' | undefined
            oid: string
        }
    }
    range: Maybe<{
        __typename?: 'Range' | undefined
        start: {
            __typename?: 'Position' | undefined
            line: number
            character: number
        }
        end: {
            __typename?: 'Position' | undefined
            line: number
            character: number
        }
    }>
}

interface RepoLocationGroup {
    repoName: string
    referenceGroups: LocationGroup[]
}

interface LocationGroup {
    repoName: string
    path: string
    references: Location[]
}

export const ReferencesList: React.FunctionComponent<
    {
        hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo'>
> = props => {
    const [activeReference, setActiveReference] = useState('')

    const { data, error, loading } = useQuery<CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables>(
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
    if (!data || !data.repository?.commit?.blob?.lsif) {
        return <>Nothing found</>
    }

    const references = data.repository.commit?.blob?.lsif?.references.nodes
    const definitions = data.repository.commit?.blob?.lsif?.definitions.nodes
    const hover = data.repository.commit?.blob?.lsif?.hover

    return (
        <div>
            <h3 className="card-header">Doc</h3>
            {hover && <div className="card-body" dangerouslySetInnerHTML={{ __html: hover.markdown.html }} />}
            <h3 className="card-header">Definitions</h3>
            {definitions.length > 0 ? (
                <LocationsList
                    locations={definitions}
                    activeReference={activeReference}
                    setActiveReference={setActiveReference}
                />
            ) : (
                <span>Nothing found</span>
            )}
            <h3 className="card-header">References</h3>
            {references.length > 0 ? (
                <LocationsList
                    locations={references}
                    activeReference={activeReference}
                    setActiveReference={setActiveReference}
                />
            ) : (
                <span>Nothing found</span>
            )}
        </div>
    )
}

const buildFileURL = (location: Location): string => {
    const path = `/${location.resource.repository.name}/-/blob/${location.resource.path}`

    if (location.range !== null) {
        return appendSubtreeQueryParameter(
            appendLineRangeQueryParameter(
                path,
                toPositionOrRangeQueryParameter({
                    range: {
                        // ATTENTION: Another off-by-one chaos in the making here
                        start: {
                            line: location.range.start.line + 1,
                            character: location.range.start.character + 1,
                        },
                        end: { line: location.range.end.line + 1, character: location.range.end.character + 1 },
                    },
                })
            )
        )
    }
    return path
}

const LocationsList: React.FunctionComponent<{
    locations: Location[]
    activeReference: string
    setActiveReference: (reference: string) => void
}> = ({ locations, activeReference, setActiveReference }) => {
    const byFile: Record<string, Location[]> = {}
    for (const location of locations) {
        if (byFile[location.resource.path] === undefined) {
            byFile[location.resource.path] = []
        }
        byFile[location.resource.path].push(location)
    }

    const locationGroups: LocationGroup[] = []
    Object.keys(byFile).map(path => {
        const references = byFile[path]
        const repoName = references[0].resource.repository.name
        locationGroups.push({ path, references, repoName })
    })

    const byRepo: Record<string, LocationGroup[]> = {}
    for (const group of locationGroups) {
        if (byRepo[group.repoName] === undefined) {
            byRepo[group.repoName] = []
        }
        byRepo[group.repoName].push(group)
    }
    const repoLocationGroups: RepoLocationGroup[] = []
    Object.keys(byRepo).map(repoName => {
        const referenceGroups = byRepo[repoName]
        repoLocationGroups.push({ repoName, referenceGroups })
    })

    const getLineContent = (location: Location): string => {
        const lines = location.resource.content.split(/\r?\n/)
        const range = location.range
        if (range !== null) {
            return lines[range.start?.line]
        }
        return ''
    }

    return (
        <>
            {repoLocationGroups.map(repoReferenceGroup => (
                <RepoReferenceGroup
                    key={repoReferenceGroup.repoName}
                    repoReferenceGroup={repoReferenceGroup}
                    activeReference={activeReference}
                    setActiveReference={setActiveReference}
                    getLineContent={getLineContent}
                    buildFileURL={buildFileURL}
                />
            ))}
        </>
    )
}

const RepoReferenceGroup: React.FunctionComponent<{
    repoReferenceGroup: RepoLocationGroup
    activeReference: string
    setActiveReference: (refeference: string) => void
    getLineContent: (reference: Location) => string
    buildFileURL: (reference: Location) => string
}> = ({ repoReferenceGroup, setActiveReference, getLineContent, buildFileURL, activeReference }) => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])

    return (
        <>
            <button
                aria-expanded={isOpen}
                type="button"
                onClick={handleOpen}
                className="bg-transparent border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                {isOpen ? (
                    <MenuUpIcon className={classNames('icon-inline', styles.chevron)} />
                ) : (
                    <MenuDownIcon className={classNames('icon-inline', styles.chevron)} />
                )}

                <span>
                    <Link to={`/${repoReferenceGroup.repoName}`}>{displayRepoName(repoReferenceGroup.repoName)}</Link>
                </span>
            </button>

            <Collapse id={repoReferenceGroup.repoName} isOpen={isOpen}>
                {repoReferenceGroup.referenceGroups.map(group => (
                    <ReferenceGroup
                        key={group.path + group.repoName}
                        group={group}
                        activeReference={activeReference}
                        setActiveReference={setActiveReference}
                        getLineContent={getLineContent}
                        buildFileURL={buildFileURL}
                    />
                ))}
            </Collapse>
        </>
    )
}

const ReferenceGroup: React.FunctionComponent<{
    group: LocationGroup
    activeReference: string
    setActiveReference: (refeference: string) => void
    getLineContent: (reference: Location) => string
    buildFileURL: (reference: Location) => string
}> = ({ group, setActiveReference, getLineContent, buildFileURL, activeReference }) => {
    const [fileBase, fileName] = splitPath(group.path)

    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])

    return (
        <div className="ml-2">
            <button
                aria-expanded={isOpen}
                type="button"
                onClick={handleOpen}
                className="bg-transparent border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                {isOpen ? (
                    <MenuUpIcon className={classNames('icon-inline', styles.chevron)} />
                ) : (
                    <MenuDownIcon className={classNames('icon-inline', styles.chevron)} />
                )}

                <span>
                    {fileBase ? `${fileBase}/` : null}
                    {fileName} ({group.references.length} references)
                </span>
            </button>

            <Collapse id={group.repoName + group.path} isOpen={isOpen}>
                <ul className="list-unstyled pl-3 py-1 mb-0">
                    {group.references.map(reference => {
                        const fileURL = buildFileURL(reference)
                        const className = activeReference === fileURL ? styles.coolCodeIntelReferenceActive : ''

                        return (
                            <li key={fileURL} className={classNames('border-0 rounded-0', className)}>
                                <div>
                                    <Link
                                        onClick={() => setActiveReference(fileURL)}
                                        to={fileURL}
                                        className={styles.referenceLink}
                                    >
                                        {reference.range?.start?.line}
                                        {': '}
                                        <code>{getLineContent(reference)}</code>
                                    </Link>
                                </div>
                            </li>
                        )
                    })}
                </ul>
            </Collapse>
        </div>
    )
}

const TABS: CoolCodeIntelToolsTab[] = [
    { id: 'token', label: 'Token', component: TokenPanel },
    { id: 'references', label: 'References', component: ReferencesPanel },
]

interface CoolCodeIntelPanelProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo'> {
    hoveredToken?: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
}

export const CoolCodeIntelPanel = React.memo<CoolCodeIntelPanelProps>(props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    return (
        <Tabs size="medium" className={styles.panel} index={tabIndex} onChange={handleTabsChange}>
            <div className={classNames('tablist-wrapper d-flex justify-content-between sticky-top', styles.header)}>
                <TabList>
                    <div className="d-flex w-100">
                        {TABS.map(({ label, id }) => (
                            <Tab key={id}>
                                <span className="tablist-wrapper--tab-label" role="none">
                                    {label}
                                </span>
                            </Tab>
                        ))}
                    </div>
                </TabList>
            </div>
            <TabPanels>
                {TABS.map(tab => (
                    <TabPanel key={tab.id} data-testid="panel-tabs-content">
                        <tab.component {...props} />
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
})

export const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelPanelProps> = props => (
    <Resizable
        className={styles.resizablePanel}
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<CoolCodeIntelPanel {...props} />}
    />
)
