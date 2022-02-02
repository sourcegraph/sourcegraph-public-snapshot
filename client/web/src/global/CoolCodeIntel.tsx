import classNames from 'classnames'
import { createMemoryHistory } from 'history'
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
} from '@sourcegraph/wildcard'

import {
    CoolCodeIntelHighlightedBlobResult,
    CoolCodeIntelHighlightedBlobVariables,
    CoolCodeIntelReferencesResult,
    CoolCodeIntelReferencesVariables,
    LocationFields,
} from '../graphql-operations'
import { Blob, BlobProps } from '../repo/blob/Blob'

import styles from './CoolCodeIntel.module.scss'
import { FETCH_HIGHLIGHTED_BLOB, FETCH_REFERENCES_QUERY } from './CoolCodeIntelQueries'

export interface GlobalCoolCodeIntelProps {
    coolCodeIntelEnabled: boolean
    onTokenClick?: (clickedToken: CoolHoveredToken) => void
}

export type CoolHoveredToken = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'> {
    hoveredToken?: CoolHoveredToken
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
        hoveredToken: CoolHoveredToken
    } & Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'>
> = props => {
    const [activeLocation, setActiveLocation] = useState<Location | undefined>(undefined)
    const [filter, setFilter] = useState<string | undefined>(undefined)
    const debouncedFilter = useDebounce(filter, 250)

    const [hoverMarkdown, setHoverMarkdown] = useState<string | undefined>(undefined)

    useEffect(() => {
        setActiveLocation(undefined)
        setFilter(undefined)
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
                {hoverMarkdown && (
                    <>
                        <Markdown
                            className={classNames('mb-0 card-body text-small', styles.hoverMarkdown)}
                            dangerousInnerHTML={renderMarkdown(hoverMarkdown)}
                        />
                        <hr />
                    </>
                )}
                <input
                    className="form-control px-2 mt-4"
                    type="text"
                    placeholder="Filter by filename..."
                    value={filter === undefined ? '' : filter}
                    onChange={event => {
                        setFilter(event.target.value)
                    }}
                />
                <SideReferences
                    {...props}
                    activeLocation={activeLocation}
                    setActiveLocation={onReferenceClick}
                    setHoverMarkdown={setHoverMarkdown}
                    filter={debouncedFilter}
                />
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
        filter: string | undefined
        setHoverMarkdown: (hoverMarkdown: string | undefined) => void
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

    const references: Location[] = []
    for (const node of data.repository.commit?.blob?.lsif?.references.nodes) {
        references.push(buildLocation(node))
    }

    const definitions: Location[] = []
    for (const node of data.repository.commit?.blob?.lsif?.definitions.nodes) {
        definitions.push(buildLocation(node))
    }

    const hover = data.repository.commit?.blob?.lsif?.hover
    if (hover) {
        props.setHoverMarkdown(hover.markdown.text)
    }

    return (
        <>
            <CardHeader>
                <h4 className="py-1 mb-0">Definitions</h4>
            </CardHeader>
            {definitions.length > 0 ? (
                <LocationsList
                    locations={definitions}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
                />
            ) : (
                <p className="text-muted pl-2">
                    <i>No definitions found</i>
                </p>
            )}
            <CardHeader>
                <h4 className="py-1 mb-0">References</h4>
            </CardHeader>
            {references.length > 0 ? (
                <LocationsList
                    locations={references}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
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
            onTokenClick={(token: CoolHoveredToken) => {
                console.log('sideblob token', token)
                if (props.onTokenClick) {
                    props.onTokenClick(token)
                }
            }}
            disableStatusBar={true}
            wrapCode={true}
            className={styles.sideBlob}
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
