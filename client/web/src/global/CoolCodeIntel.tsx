import classNames from 'classnames'
import { createMemoryHistory } from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
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
import {
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    Link,
    LoadingSpinner,
    useLocalStorage,
    CardHeader,
    useDebounce,
    Button,
} from '@sourcegraph/wildcard'

import {
    CoolCodeIntelHighlightedBlobResult,
    CoolCodeIntelHighlightedBlobVariables,
    CoolCodeIntelReferencesResult,
    CoolCodeIntelReferencesVariables,
    LocationFields,
    Maybe,
} from '../graphql-operations'
import { Blob, BlobProps } from '../repo/blob/Blob'

import styles from './CoolCodeIntel.module.scss'
import { FETCH_HIGHLIGHTED_BLOB, FETCH_REFERENCES_QUERY } from './CoolCodeIntelQueries'

export interface GlobalCoolCodeIntelProps {
    coolCodeIntelEnabled: boolean
    onTokenClick?: (clickedToken: CoolClickedToken) => void
}

export type CoolClickedToken = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'> {
    clickedToken?: CoolClickedToken
}

export const isCoolCodeIntelEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.coolCodeIntel !== false

export const CoolCodeIntel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    if (!props.coolCodeIntelEnabled) {
        return null
    }

    return <CoolCodeIntelResizablePanel {...props} />
}

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references' | 'token' | 'definition'

interface CoolCodeIntelTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<CoolCodeIntelProps>
}

export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    if (!props.clickedToken) {
        return null
    }

    return <ReferencesList clickedToken={props.clickedToken} {...props} />
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

const buildLocation = (node: LocationFields): Location => {
    const location: Location = {
        resource: {
            repository: { name: node.resource.repository.name },
            content: node.resource.content,
            path: node.resource.path,
            commit: node.resource.commit,
        },
        url: '',
        lines: [],
    }
    if (node.range !== null) {
        location.range = node.range
    }
    location.url = buildFileURL(location)
    location.lines = location.resource.content.split(/\r?\n/)
    return location
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
        clickedToken: CoolClickedToken
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const [activeLocation, setActiveLocation] = useState<Location | undefined>(undefined)
    const [filter, setFilter] = useState<string | undefined>(undefined)
    const debouncedFilter = useDebounce(filter, 150)

    useEffect(() => {
        setActiveLocation(undefined)
        setFilter(undefined)
    }, [props.clickedToken])

    const history = useMemo(() => createMemoryHistory(), [])

    const onReferenceClick = (location: Location | undefined): void => {
        if (location) {
            history.push(location.url)
        }
        setActiveLocation(location)
    }

    return (
        <>
            <input
                className={classNames('form-control px-2', styles.referencesFilter)}
                type="text"
                placeholder="Filter by filename..."
                value={filter === undefined ? '' : filter}
                onChange={event => setFilter(event.target.value)}
            />
            <div className={classNames('align-items-stretch', styles.referencesList)}>
                <div className={classNames('px-0', styles.referencesSideReferences)}>
                    <SideReferences
                        {...props}
                        activeLocation={activeLocation}
                        setActiveLocation={onReferenceClick}
                        filter={debouncedFilter}
                    />
                </div>
                {activeLocation !== undefined && (
                    <div className={classNames('px-0 border-left', styles.referencesSideBlob)}>
                        <SideBlob
                            {...props}
                            history={history}
                            location={history.location}
                            activeLocation={activeLocation}
                        />
                    </div>
                )}
            </div>
        </>
    )
}

export const SideReferences: React.FunctionComponent<
    {
        clickedToken: CoolClickedToken
        setActiveLocation: (location: Location | undefined) => void
        activeLocation: Location | undefined
        filter: string | undefined
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const { data, error, loading } = useQuery<CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables>(
        FETCH_REFERENCES_QUERY,
        {
            variables: {
                repository: props.clickedToken.repoName,
                commit: props.clickedToken.commitID,
                path: props.clickedToken.filePath,
                // ATTENTION: Off by one ahead!!!!
                line: props.clickedToken.line - 1,
                character: props.clickedToken.character - 1,
                after: null,
                filter: props.filter || null,
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

    const lsif = data.repository?.commit?.blob?.lsif

    return (
        <SideReferencesLists
            {...props}
            references={lsif.references}
            definitions={lsif.definitions}
            implementations={lsif.implementations}
            hover={lsif.hover}
        />
    )
}

interface LSIFLocationResult {
    __typename?: 'LocationConnection'
    nodes: ({ __typename?: 'Location' } & LocationFields)[]
    pageInfo: { __typename?: 'PageInfo'; endCursor: Maybe<string> }
}

export const SideReferencesLists: React.FunctionComponent<
    {
        clickedToken: CoolClickedToken
        setActiveLocation: (location: Location | undefined) => void
        activeLocation: Location | undefined
        filter: string | undefined
        references: LSIFLocationResult
        definitions: Omit<LSIFLocationResult, 'pageInfo'>
        implementations: LSIFLocationResult
        hover: Maybe<{
            __typename?: 'Hover'
            markdown: { __typename?: 'Markdown'; html: string; text: string }
        }>
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const { references, definitions, implementations, hover } = props
    const references_: Location[] = useMemo(() => references.nodes.map(buildLocation), [references])
    const defs: Location[] = useMemo(() => definitions.nodes.map(buildLocation), [definitions])
    const impls: Location[] = useMemo(() => implementations.nodes.map(buildLocation), [implementations])

    return (
        <>
            {hover && (
                <Markdown
                    className={classNames('mb-0 card-body text-small', styles.hoverMarkdown)}
                    dangerousInnerHTML={renderMarkdown(hover.markdown.text)}
                />
            )}
            <CardHeader>
                <h4 className="py-1 px-1 mb-0">Definitions</h4>
            </CardHeader>
            {defs.length > 0 ? (
                <LocationsList
                    locations={defs}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
                />
            ) : (
                <p className="text-muted my-1 pl-2">
                    {props.filter ? (
                        <i>
                            No definitions matching <strong>{props.filter}</strong> found
                        </i>
                    ) : (
                        <i>No definitions found</i>
                    )}
                </p>
            )}
            <CardHeader>
                <h4 className="py-1 px-1 mb-0">References</h4>
            </CardHeader>
            {references_.length > 0 ? (
                <LocationsList
                    locations={references_}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
                />
            ) : (
                <p className="text-muted pl-2">
                    {props.filter ? (
                        <i>
                            No references matching <strong>{props.filter}</strong> found
                        </i>
                    ) : (
                        <i>No references found</i>
                    )}
                </p>
            )}
            {impls.length > 0 && (
                <>
                    <CardHeader>
                        <h4 className="py-1 px-1 mb-0">Implementations</h4>
                    </CardHeader>
                    <LocationsList
                        locations={impls}
                        activeLocation={props.activeLocation}
                        setActiveLocation={props.setActiveLocation}
                        filter={props.filter}
                    />
                </>
            )}
        </>
    )
}

export const SideBlob: React.FunctionComponent<
    {
        activeLocation: Location
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const { data, error, loading } = useQuery<
        CoolCodeIntelHighlightedBlobResult,
        CoolCodeIntelHighlightedBlobVariables
    >(FETCH_HIGHLIGHTED_BLOB, {
        variables: {
            repository: props.activeLocation.resource.repository.name,
            commit: props.activeLocation.resource.commit.oid,
            path: props.activeLocation.resource.path,
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'network-only',
    })

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <>
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>
                        Loading <code>{props.activeLocation.resource.path}</code>...
                    </i>
                </p>
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !data) {
        throw new Error(error.message)
    }

    // If there weren't any errors and we just didn't receive any data
    if (!data?.repository?.commit?.blob?.highlight) {
        return <>Nothing found</>
    }

    const { html, aborted } = data?.repository?.commit?.blob?.highlight
    if (aborted) {
        return (
            <p className="text-warning text-center">
                <i>
                    Highlighting <code>{props.activeLocation.resource.path}</code> failed
                </i>
            </p>
        )
    }

    return (
        <Blob
            {...props}
            onTokenClick={(token: CoolClickedToken) => {
                console.log('sideblob token', token)
                if (props.onTokenClick) {
                    props.onTokenClick(token)
                }
            }}
            disableStatusBar={true}
            wrapCode={true}
            className={styles.referencesSideBlob}
            blobInfo={{
                content: props.activeLocation.resource.content,
                html,
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
    filter: string | undefined
}> = ({ locations, activeLocation, setActiveLocation, filter }) => {
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
                    filter={filter}
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
    filter: string | undefined
}> = ({ repoReferenceGroup, setActiveLocation, getLineContent, activeLocation, filter }) => {
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
                        filter={filter}
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
    filter: string | undefined
}> = ({ group, setActiveLocation: setActiveLocation, getLineContent, activeLocation, filter }) => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])

    let highlighted = [group.path]
    if (filter !== undefined) {
        highlighted = group.path.split(filter)
    }

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
                    {highlighted.length === 2 ? (
                        <span>
                            {highlighted[0]}
                            <mark>{filter}</mark>
                            {highlighted[1]}
                        </span>
                    ) : (
                        group.path
                    )}{' '}
                    ({group.locations.length} references)
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

const TABS: CoolCodeIntelTab[] = [{ id: 'references', label: 'References', component: ReferencesPanel }]

export const CoolCodeIntelPanel = React.memo<CoolCodeIntelProps & { handlePanelClose: (closed: boolean) => void }>(
    props => {
        const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
        const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

        return (
            <Tabs size="medium" className={styles.panel} index={tabIndex} onChange={handleTabsChange}>
                <div
                    className={classNames(
                        'tablist-wrapper d-flex justify-content-between sticky-top',
                        styles.panelHeader
                    )}
                >
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
                    <div className="align-items-center d-flex">
                        <Button
                            onClick={() => props.handlePanelClose(true)}
                            className={classNames('btn-icon ml-2', styles.dismissButton)}
                            title="Close panel"
                            data-tooltip="Close panel"
                            data-placement="left"
                        >
                            <CloseIcon className="icon-inline" />
                        </Button>
                    </div>
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
    }
)

export const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    const [closed, close] = useState(false)
    const handlePanelClose = useCallback(() => close(true), [])
    useEffect(() => {
        if (props.clickedToken) {
            close(false)
        }
    }, [props.clickedToken])

    if (closed) {
        return null
    }

    if (!props.clickedToken) {
        return null
    }

    return (
        <Resizable
            className={styles.resizablePanel}
            handlePosition="top"
            defaultSize={350}
            storageKey="panel-size"
            element={<CoolCodeIntelPanel {...props} handlePanelClose={handlePanelClose} />}
        />
    )
}
