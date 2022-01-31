import classNames from 'classnames'
import { createMemoryHistory } from 'history'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { useQuery } from '@sourcegraph/http-client'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
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

import { CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables } from '../graphql-operations'
import { Blob, BlobProps } from '../repo/blob/Blob'

import styles from './GlobalCodeIntel.module.scss'
import { FETCH_REFERENCES_QUERY } from './GlobalCodeIntelQueries'

const SHOW_COOL_CODEINTEL = localStorage.getItem('coolCodeIntel') !== null

type CoolHoveredToken = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'> {
    hoveredToken?: CoolHoveredToken
}

export const GlobalCodeIntel: React.FunctionComponent<
    {
        showPanel: boolean
    } & CoolCodeIntelProps
> = props => {
    if (!SHOW_COOL_CODEINTEL) {
        return null
    }

    if (props.showPanel) {
        return <CoolCodeIntelResizablePanel {...props} />
    }

    return null
}

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references' | 'token' | 'definition'

interface CoolCodeIntelTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<CoolCodeIntelProps>
}

export const TokenPanel: React.FunctionComponent<CoolCodeIntelProps> = props => (
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

export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    if (!props.hoveredToken) {
        return null
    }

    return <ReferencesList hoveredToken={props.hoveredToken} {...props} />
}

interface Location {
    resource: {
        path: string
        content: string
        repository: {
            name: string
        }
        commit: {
            oid: string
        }
        highlight: {
            html: string
        }
    }
    range?: {
        start: {
            line: number
            character: number
        }
        end: {
            line: number
            character: number
        }
    }

    url: string
    lines: string[]
}

interface RepoLocationGroup {
    repoName: string
    referenceGroups: LocationGroup[]
}

interface LocationGroup {
    repoName: string
    path: string
    locations: Location[]
}

export const ReferencesList: React.FunctionComponent<
    {
        hoveredToken: CoolHoveredToken
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const [activeLocation, setActiveLocation] = useState<Location | undefined>(undefined)

    useEffect(() => {
        setActiveLocation(undefined)
    }, [props.hoveredToken])

    const history = useMemo(() => createMemoryHistory(), [])

    const onReferenceClick = (location: Location | undefined): void => {
        if (location) {
            history.push(location.url)
        }
        setActiveLocation(location)
    }

    return (
        <div className={classNames('align-items-stretch', styles.referencesList)}>
            <div className={classNames('px-0', styles.sideReferences)}>
                <SideReferences {...props} activeLocation={activeLocation} setActiveLocation={onReferenceClick} />
            </div>
            {activeLocation !== undefined && (
                <div className={classNames('px-0 border-left', styles.sideBlob)}>
                    <SideBlob
                        {...props}
                        history={history}
                        location={history.location}
                        activeLocation={activeLocation}
                    />
                </div>
            )}
        </div>
    )
}

export const SideReferences: React.FunctionComponent<
    {
        hoveredToken: CoolHoveredToken
        setActiveLocation: (location: Location | undefined) => void
        activeLocation: Location | undefined
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
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
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>Loading references ...</i>
                </p>
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

    const references: Location[] = []
    for (const node of data.repository.commit?.blob?.lsif?.references.nodes) {
        const location: Location = {
            resource: {
                repository: { name: node.resource.repository.name },
                content: node.resource.content,
                path: node.resource.path,
                commit: node.resource.commit,
                highlight: { html: node.resource.highlight.html },
            },
            url: '',
            lines: [],
        }
        if (node.range !== null) {
            location.range = node.range
        }
        location.url = buildFileURL(location)
        location.lines = location.resource.content.split(/\r?\n/)
        references.push(location)
    }

    const definitions: Location[] = []
    for (const node of data.repository.commit?.blob?.lsif?.definitions.nodes) {
        const location: Location = {
            resource: {
                repository: { name: node.resource.repository.name },
                content: node.resource.content,
                path: node.resource.path,
                commit: node.resource.commit,
                highlight: { html: node.resource.highlight.html },
            },
            url: '',
            lines: [],
        }
        if (node.range !== null) {
            location.range = node.range
        }
        location.url = buildFileURL(location)
        location.lines = location.resource.content.split(/\r?\n/)
        definitions.push(location)
    }

    const hover = data.repository.commit?.blob?.lsif?.hover

    return (
        <>
            {hover && (
                <Markdown
                    className={classNames('mb-0 card-body text-small', styles.hoverMarkdown)}
                    dangerousInnerHTML={renderMarkdown(hover.markdown.text)}
                />
            )}
            <h4 className="card-header py-1">Definitions</h4>
            {definitions.length > 0 ? (
                <LocationsList
                    locations={definitions}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                />
            ) : (
                <p className="text-muted pl-2">
                    <i>No definitions found</i>
                </p>
            )}
            <h4 className="card-header py-1">References</h4>
            {references.length > 0 ? (
                <LocationsList
                    locations={references}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                />
            ) : (
                <p className="text-muted pl-2">
                    <i>No references found</i>
                </p>
            )}
        </>
    )
}

export const SideBlob: React.FunctionComponent<
    {
        hoveredToken: CoolHoveredToken
        activeLocation: Location
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
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
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>Loading references ...</i>
                </p>
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
    return (
        <Blob
            {...props}
            onHoverToken={(token: CoolHoveredToken) => {
                console.log('sideblob token', token)
                props.onHoverToken(token)
            }}
            disableStatusBar={true}
            wrapCode={true}
            className={styles.sideBlob}
            blobInfo={{
                content: props.activeLocation.resource.content,
                html: props.activeLocation.resource.highlight.html,
                filePath: props.activeLocation.resource.path,
                repoName: props.activeLocation.resource.repository.name,
                commitID: props.activeLocation.resource.commit.oid,
                revision: props.activeLocation.resource.commit.oid,
                mode: 'lspmode',
            }}
        />
    )
}

const buildFileURL = (location: Location): string => {
    const path = `/${location.resource.repository.name}/-/blob/${location.resource.path}`
    const range = location.range

    if (range !== undefined) {
        return appendSubtreeQueryParameter(
            appendLineRangeQueryParameter(
                path,
                toPositionOrRangeQueryParameter({
                    range: {
                        // ATTENTION: Another off-by-one chaos in the making here
                        start: {
                            line: range.start.line + 1,
                            character: range.start.character + 1,
                        },
                        end: { line: range.end.line + 1, character: range.end.character + 1 },
                    },
                })
            )
        )
    }
    return path
}

const LocationsList: React.FunctionComponent<{
    locations: Location[]
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
}> = ({ locations, activeLocation, setActiveLocation }) => {
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
        locationGroups.push({ path, locations: references, repoName })
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
        const range = location.range
        if (range !== undefined) {
            return location.lines[range.start?.line].trim()
        }
        return ''
    }

    return (
        <>
            {repoLocationGroups.map(repoReferenceGroup => (
                <RepoReferenceGroup
                    key={repoReferenceGroup.repoName}
                    repoReferenceGroup={repoReferenceGroup}
                    activeLocation={activeLocation}
                    setActiveLocation={setActiveLocation}
                    getLineContent={getLineContent}
                />
            ))}
        </>
    )
}

const RepoReferenceGroup: React.FunctionComponent<{
    repoReferenceGroup: RepoLocationGroup
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    getLineContent: (location: Location) => string
}> = ({ repoReferenceGroup, setActiveLocation, getLineContent, activeLocation }) => {
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
                        activeLocation={activeLocation}
                        setActiveLocation={setActiveLocation}
                        getLineContent={getLineContent}
                    />
                ))}
            </Collapse>
        </>
    )
}

const ReferenceGroup: React.FunctionComponent<{
    group: LocationGroup
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    getLineContent: (reference: Location) => string
}> = ({ group, setActiveLocation: setActiveLocation, getLineContent, activeLocation }) => {
    const [fileBase, fileName] = splitPath(group.path)

    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])

    return (
        <div className="ml-4">
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

                <span className={styles.coolCodeIntelReferenceFilename}>
                    {fileBase ? `${fileBase}/` : null}
                    {fileName} ({group.locations.length} references)
                </span>
            </button>

            <Collapse id={group.repoName + group.path} isOpen={isOpen} className="ml-2">
                <ul className="list-unstyled pl-3 py-1 mb-0">
                    {group.locations.map(reference => {
                        const className =
                            activeLocation && activeLocation.url === reference.url
                                ? styles.coolCodeIntelReferenceActive
                                : ''

                        return (
                            <li key={reference.url} className={classNames('border-0 rounded-0', className)}>
                                <div>
                                    <Link
                                        onClick={event => {
                                            event.preventDefault()
                                            setActiveLocation(reference)
                                        }}
                                        to={reference.url}
                                        className={styles.referenceLink}
                                    >
                                        <span>
                                            {reference.range?.start?.line}
                                            {': '}
                                        </span>
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

const TABS: CoolCodeIntelTab[] = [
    { id: 'token', label: 'Token', component: TokenPanel },
    { id: 'references', label: 'References', component: ReferencesPanel },
]

export const CoolCodeIntelPanel = React.memo<CoolCodeIntelProps>(props => {
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
                    <TabPanel key={tab.id}>
                        <tab.component {...props} />
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
})

export const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    if (!props.hoveredToken) {
        return null
    }

    return (
        <Resizable
            className={styles.resizablePanel}
            handlePosition="top"
            defaultSize={350}
            storageKey="panel-size"
            element={<CoolCodeIntelPanel {...props} />}
        />
    )
}
