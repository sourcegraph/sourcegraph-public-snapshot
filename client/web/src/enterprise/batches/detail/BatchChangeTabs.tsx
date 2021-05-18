import classNames from 'classnames'
import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
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
    queryBulkOperations as _queryBulkOperations,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BatchSpecTab } from './BatchSpecTab'
import { BulkOperationsTab } from './BulkOperationsTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'

export enum BatchChangeTab {
    CHANGESETS = 'changesets',
    CHART = 'chart',
    SPEC = 'spec',
    ARCHIVED = 'archived',
    BULK_OPERATIONS = 'bulkoperations',
}

export interface BatchChangeTabsProps
    extends ExtensionsControllerProps,
        ThemeProps,
        PlatformContextProps,
        TelemetryProps {
    batchChange: BatchChangeFields
    changesetsCount: number
    archivedCount: number
    bulkOperationsCount: number
    history: H.History
    location: H.Location
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
    /** For testing only. */
    queryBulkOperations?: typeof _queryBulkOperations
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
    bulkOperationsCount,
    queryChangesets,
    queryChangesetCountsOverTime,
    queryExternalChangesetWithFileDiffs,
    queryBulkOperations,
}) => {
    const [selectedTab, setSelectedTab] = useState<BatchChangeTab>(selectedTabFromLocation(location.search))
    useEffect(() => {
        const newTab = selectedTabFromLocation(location.search)
        if (newTab !== selectedTab) {
            setSelectedTab(newTab)
        }
    }, [location.search, selectedTab])

    const onSelectChangesets = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab(BatchChangeTab.CHANGESETS)
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
            setSelectedTab(BatchChangeTab.CHART)
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', BatchChangeTab.CHART)
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
            setSelectedTab(BatchChangeTab.SPEC)
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', BatchChangeTab.SPEC)
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
            setSelectedTab(BatchChangeTab.ARCHIVED)
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', BatchChangeTab.ARCHIVED)
            removeConnectionParameters(urlParameters)
            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )
    const onSelectBulkOperations = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab(BatchChangeTab.BULK_OPERATIONS)
            const urlParameters = new URLSearchParams(location.search)
            urlParameters.set('tab', BatchChangeTab.BULK_OPERATIONS)
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
                            className={classNames('nav-link', selectedTab === BatchChangeTab.CHANGESETS && 'active')}
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
                            className={classNames('nav-link', selectedTab === BatchChangeTab.CHART && 'active')}
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
                            className={classNames('nav-link', selectedTab === BatchChangeTab.SPEC && 'active')}
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
                            className={classNames('nav-link', selectedTab === BatchChangeTab.ARCHIVED && 'active')}
                        >
                            <ArchiveIcon className="icon-inline text-muted mr-1" /> Archived{' '}
                            <span className="badge badge-pill badge-secondary ml-1">{archivedCount}</span>
                        </a>
                    </li>
                    <li className="nav-item">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            role="button"
                            onClick={onSelectBulkOperations}
                            className={classNames(
                                'nav-link',
                                selectedTab === BatchChangeTab.BULK_OPERATIONS && 'active'
                            )}
                        >
                            <MonitorStarIcon className="icon-inline text-muted mr-1" /> Bulk operations{' '}
                            <span className="badge badge-pill badge-secondary ml-1">{bulkOperationsCount}</span>
                        </a>
                    </li>
                </ul>
            </div>
            {selectedTab === BatchChangeTab.CHART && (
                <BatchChangeBurndownChart
                    batchChangeID={batchChange.id}
                    queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                    history={history}
                />
            )}
            {selectedTab === BatchChangeTab.CHANGESETS && (
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
            {selectedTab === BatchChangeTab.SPEC && (
                <BatchSpecTab batchChange={batchChange} originalInput={batchChange.currentSpec.originalInput} />
            )}
            {selectedTab === BatchChangeTab.ARCHIVED && (
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
                />
            )}
            {selectedTab === BatchChangeTab.BULK_OPERATIONS && (
                <BulkOperationsTab
                    batchChangeID={batchChange.id}
                    history={history}
                    location={location}
                    queryBulkOperations={queryBulkOperations}
                />
            )}
        </>
    )
}

function selectedTabFromLocation(locationSearch: string): BatchChangeTab {
    const urlParameters = new URLSearchParams(locationSearch)
    const tabParameter = urlParameters.get('tab')
    if (tabParameter && isValidTabParameter(tabParameter)) {
        return tabParameter
    }
    return BatchChangeTab.CHANGESETS
}

function isValidTabParameter(value: string): value is BatchChangeTab {
    return Object.values<string>(BatchChangeTab).includes(value)
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
