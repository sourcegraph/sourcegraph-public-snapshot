import React, { useMemo } from 'react'

import * as H from 'history'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, Typography, useObservable } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { DiffStat } from '../../../components/diff/DiffStat'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { RepositoryFields, RepoBatchChangeStats } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { BatchChangeStatsTotalAction } from '../detail/BatchChangeStatsCard'
import {
    ChangesetStatusUnpublished,
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from '../detail/changesets/ChangesetStatusCell'
import { NewBatchChangeButton } from '../list/NewBatchChangeButton'

import {
    queryRepoBatchChanges as _queryRepoBatchChanges,
    queryRepoBatchChangeStats as _queryRepoBatchChangeStats,
} from './backend'
import { RepoBatchChanges } from './RepoBatchChanges'

interface BatchChangeRepoPageProps extends ThemeProps {
    history: H.History
    location: H.Location
    repo: RepositoryFields
    /** For testing only. */
    queryRepoBatchChangeStats?: typeof _queryRepoBatchChangeStats
    /** For testing only. */
    queryRepoBatchChanges?: typeof _queryRepoBatchChanges
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const BatchChangeRepoPage: React.FunctionComponent<React.PropsWithChildren<BatchChangeRepoPageProps>> = ({
    repo,
    queryRepoBatchChangeStats = _queryRepoBatchChangeStats,
    ...context
}) => {
    const repoDisplayName = displayRepoName(repo.name)

    const stats: RepoBatchChangeStats | undefined = useObservable(
        useMemo(() => queryRepoBatchChangeStats({ name: repo.name }), [queryRepoBatchChangeStats, repo.name])
    )
    const hasChangesets = stats?.changesetsStats.total

    return (
        <Page>
            <PageTitle title="Batch Changes" />
            <PageHeader
                path={[{ icon: BatchChangesIcon, text: 'Batch Changes' }]}
                headingElement="h1"
                actions={hasChangesets ? undefined : <NewBatchChangeButton to="/batch-changes/create" />}
                description={
                    hasChangesets
                        ? undefined
                        : 'Run custom code over this repository and others, and manage the resulting changesets.'
                }
            />
            {hasChangesets && stats?.batchChangesDiffStat && stats?.changesetsStats ? (
                <div className="d-flex align-items-center mt-4 mb-3">
                    <Typography.H2 className="mb-0 pb-1">{repoDisplayName}</Typography.H2>
                    <DiffStat className="d-flex flex-1 ml-2" expandedCounts={true} {...stats.batchChangesDiffStat} />
                    <StatsBar stats={stats.changesetsStats} />
                </div>
            ) : null}
            {hasChangesets ? (
                <p>
                    Batch changes has created {stats?.changesetsStats.total} changesets on {repoDisplayName}
                </p>
            ) : (
                <div className="mb-3" />
            )}
            <RepoBatchChanges viewerCanAdminister={true} repo={repo} {...context} />
        </Page>
    )
}

const ACTION_CLASSNAMES = 'd-flex flex-column text-muted justify-content-center align-items-center mx-2'

interface StatsBarProps {
    stats: RepoBatchChangeStats['changesetsStats']
}

const StatsBar: React.FunctionComponent<React.PropsWithChildren<StatsBarProps>> = ({
    stats: { total, draft, open, unpublished, closed, merged },
}) => (
    <div className="d-flex flex-wrap align-items-center">
        <BatchChangeStatsTotalAction count={total} />
        <ChangesetStatusOpen className={ACTION_CLASSNAMES} label={`${(draft + open).toString()} Open`} />
        <ChangesetStatusUnpublished className={ACTION_CLASSNAMES} label={`${unpublished} Unpublished`} />
        <ChangesetStatusClosed className={ACTION_CLASSNAMES} label={`${closed} Closed`} />
        <ChangesetStatusMerged className={ACTION_CLASSNAMES} label={`${merged} Merged`} />
    </div>
)
