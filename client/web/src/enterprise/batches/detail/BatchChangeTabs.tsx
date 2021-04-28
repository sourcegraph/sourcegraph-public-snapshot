import classNames from 'classnames'
import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React, { useState, useCallback, useEffect } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BatchChangeFields } from '../../../graphql-operations'

import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BatchSpecTab } from './BatchSpecTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'

type SelectedTab = 'changesets' | 'chart' | 'spec' | 'archived'

export interface BatchChangeTabsProps
    extends ExtensionsControllerProps,
        ThemeProps,
        PlatformContextProps,
        TelemetryProps {
    batchChange: BatchChangeFields
    changesetsCount: number
    archivedCount: number
    history: H.History
    location: H.Location
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
}

export const BatchChangeTabs: React.FunctionComponent<BatchChangeTabsProps> = ({
    extensionsController,
    history,
    isLightTheme,
    location,
    platformContext,
    telemetryService,
    batchChange,
    changesetsCount,
    archivedCount,
    queryChangesets,
    queryChangesetCountsOverTime,
    queryExternalChangesetWithFileDiffs,
}) => {
    const [selectedTab, setSelectedTab] = useState<SelectedTab>(selectedTabFromLocation(location.search))
    useEffect(() => {
        const newTab = selectedTabFromLocation(location.search)
        if (newTab !== selectedTab) {
            setSelectedTab(newTab)
        }
    }, [location.search, selectedTab])

    const onSelectChangesets = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('changesets')
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.delete('tab')
            removeConnectionParameters(urlParameters)
            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )
    const onSelectChart = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('chart')
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', 'chart')
            removeConnectionParameters(urlParameters)
            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )
    const onSelectSpec = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('spec')
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', 'spec')
            removeConnectionParameters(urlParameters)
            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )
    const onSelectArchived = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('archived')
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', 'archived')
            removeConnectionParameters(urlParameters)
            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )

    return (
        <>
            <div className="overflow-auto mb-2">
                <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
                    <li className="nav-item">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            role="button"
                            onClick={onSelectChangesets}
                            className={classNames('nav-link', selectedTab === 'changesets' && 'active')}
                        >
                            <SourceBranchIcon className="icon-inline text-muted mr-1" />
                            Changesets <span className="badge badge-pill badge-secondary ml-1">{changesetsCount}</span>
                        </a>
                    </li>
                    <li className="nav-item test-batches-chart-tab">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            role="button"
                            onClick={onSelectChart}
                            className={classNames('nav-link', selectedTab === 'chart' && 'active')}
                        >
                            <ChartLineVariantIcon className="icon-inline text-muted mr-1" /> Burndown chart
                        </a>
                    </li>
                    <li className="nav-item test-batches-spec-tab">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            role="button"
                            onClick={onSelectSpec}
                            className={classNames('nav-link', selectedTab === 'spec' && 'active')}
                        >
                            <FileDocumentIcon className="icon-inline text-muted mr-1" /> Spec
                        </a>
                    </li>
                    <li className="nav-item">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            role="button"
                            onClick={onSelectArchived}
                            className={classNames('nav-link', selectedTab === 'archived' && 'active')}
                        >
                            <ArchiveIcon className="icon-inline text-muted mr-1" /> Archived{' '}
                            <span className="badge badge-pill badge-secondary ml-1">{archivedCount}</span>
                        </a>
                    </li>
                </ul>
            </div>
            {selectedTab === 'chart' && (
                <BatchChangeBurndownChart
                    batchChangeID={batchChange.id}
                    queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                    history={history}
                />
            )}
            {selectedTab === 'changesets' && (
                <BatchChangeChangesets
                    batchChangeID={batchChange.id}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    onlyArchived={false}
                />
            )}
            {selectedTab === 'spec' && (
                <BatchSpecTab batchChange={batchChange} originalInput={batchChange.currentSpec.originalInput} />
            )}
            {selectedTab === 'archived' && (
                <BatchChangeChangesets
                    batchChangeID={batchChange.id}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    onlyArchived={true}
                    enableSelect={true}
                />
            )}
        </>
    )
}

function selectedTabFromLocation(locationSearch: string): SelectedTab {
    const urlParameters = new URLSearchParams(locationSearch)
    if (urlParameters.get('tab') === 'chart') {
        return 'chart'
    }
    if (urlParameters.get('tab') === 'spec') {
        return 'spec'
    }
    if (urlParameters.get('tab') === 'archived') {
        return 'archived'
    }
    return 'changesets'
}

function removeConnectionParameters(parameters: URLSearchParams): void {
    if (parameters.has('visible')) {
        parameters.delete('visible')
    }
    if (parameters.has('first')) {
        parameters.delete('first')
    }
    if (parameters.has('after')) {
        parameters.delete('after')
    }
}
