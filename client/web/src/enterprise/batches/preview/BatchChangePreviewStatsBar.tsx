import React, { useContext, useMemo } from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { Badge, H2, H3, H4, useObservable } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../components/diff/DiffStat'
import type { ApplyPreviewStatsFields, DiffStatFields, Scalars } from '../../../graphql-operations'

import { queryApplyPreviewStats as _queryApplyPreviewStats } from './backend'
import { BatchChangePreviewContext } from './BatchChangePreviewContext'
import { ChangesetAddedIcon, ChangesetModifiedIcon, ChangesetRemovedIcon } from './icons'
import {
    PreviewArchiveStat,
    PreviewCloseStat,
    PreviewImportStat,
    PreviewPublishStat,
    PreviewReattachStat,
    PreviewReopenStat,
    PreviewUndraftStat,
    PreviewUpdateStat,
} from './list/PreviewActions'

import styles from './BatchChangePreviewStatsBar.module.scss'

const actionClassNames = classNames(
    styles.batchChangePreviewStatsBarStat,
    styles.batchChangePreviewStatsBarState,
    'd-flex flex-column justify-content-center align-items-center mx-2'
)

export interface BatchChangePreviewStatsBarProps {
    batchSpec: Scalars['ID']
    diffStat: DiffStatFields
    /** For testing purposes only. */
    queryApplyPreviewStats?: typeof _queryApplyPreviewStats
}

export const BatchChangePreviewStatsBar: React.FunctionComponent<
    React.PropsWithChildren<BatchChangePreviewStatsBarProps>
> = ({ batchSpec, diffStat, queryApplyPreviewStats = _queryApplyPreviewStats }) => {
    // `BatchChangePreviewContext` is responsible for managing the overrideable
    // publication states for preview changesets on the clientside.
    const { publicationStates } = useContext(BatchChangePreviewContext)

    /** We use this to recalculate the stats when the publication states are modified. */
    const stats = useObservable<ApplyPreviewStatsFields['stats']>(
        useMemo(
            () => queryApplyPreviewStats({ batchSpec, publicationStates }),
            [publicationStates, batchSpec, queryApplyPreviewStats]
        )
    )

    if (!stats) {
        return null
    }

    return (
        <div className="d-flex flex-wrap mb-3 align-items-center">
            <H2 className="m-0 align-self-center">
                <VisuallyHidden>
                    This is a preview of the changesets generated from executing the batch spec.
                </VisuallyHidden>
                <Badge variant="info" className="text-uppercase mb-0" aria-hidden={true}>
                    Preview
                </Badge>
            </H2>
            <div className={classNames(styles.batchChangePreviewStatsBarDivider, 'd-none d-sm-block mx-3')} />
            <DiffStatStack className={styles.batchChangePreviewStatsBarDiff} {...diffStat} />
            <div className={classNames(styles.batchChangePreviewStatsBarHorizontalDivider, 'd-block d-sm-none my-3')} />
            <div className={classNames(styles.batchChangePreviewStatsBarDivider, 'mx-3 d-none d-sm-block d-md-none')} />
            <div
                className={classNames(
                    styles.batchChangePreviewStatsBarMetrics,
                    'flex-grow-1 d-flex justify-content-end'
                )}
            >
                <PreviewStatsAdded count={stats.added} />
                <PreviewStatsRemoved count={stats.removed} />
                <PreviewStatsModified count={stats.modified} />
            </div>
            <div className={classNames(styles.batchChangePreviewStatsBarHorizontalDivider, 'd-block d-md-none my-3')} />
            <div className={classNames(styles.batchChangePreviewStatsBarDivider, 'd-none d-md-block ml-3 mr-2')} />
            <div className={classNames(styles.batchChangePreviewStatsBarStates, 'd-flex justify-content-end')}>
                <PreviewReopenStat className={actionClassNames} count={stats.reopen} />
                <PreviewCloseStat className={actionClassNames} count={stats.close} />
                <PreviewUpdateStat className={actionClassNames} count={stats.update} />
                <PreviewUndraftStat className={actionClassNames} count={stats.undraft} />
                <PreviewPublishStat className={actionClassNames} count={stats.publish} />
                <PreviewImportStat className={actionClassNames} count={stats.import} />
                <PreviewArchiveStat className={actionClassNames} count={stats.archive} />
                <PreviewReattachStat className={actionClassNames} count={stats.reattach} />
            </div>
        </div>
    )
}

export const PreviewStatsAdded: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({ count }) => (
    <div className={classNames(styles.batchChangePreviewStatsBarStat, 'd-flex flex-column mr-2 text-nowrap')}>
        <div className="d-flex flex-column align-items-center justify-content-center">
            <span className={styles.previewStatsAddedLine}>&nbsp;</span>
            <span
                className={classNames(styles.previewStatsAddedIcon, 'd-flex justify-content-center align-items-center')}
            >
                <ChangesetAddedIcon />
            </span>
            <span className={styles.previewStatsAddedLine}>&nbsp;</span>
        </div>
        <H4
            as={H3}
            className="font-weight-normal mt-1 mb-0"
            aria-label={`${count} ${pluralize('changeset', count)} added`}
        >
            {`${count} added`}
        </H4>
    </div>
)
export const PreviewStatsModified: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({
    count,
}) => (
    <div className={classNames(styles.batchChangePreviewStatsBarStat, 'd-flex flex-column text-nowrap ml-2')}>
        <div className="d-flex flex-column align-items-center">
            <span className={styles.previewStatsModifiedLine}>&nbsp;</span>
            <span
                className={classNames(
                    styles.previewStatsModifiedIcon,
                    'd-flex justify-content-center align-items-center'
                )}
            >
                <ChangesetModifiedIcon />
            </span>
            <span className={styles.previewStatsModifiedLine}>&nbsp;</span>
        </div>
        <H4
            as={H3}
            className="font-weight-normal mt-1 mb-0"
            aria-label={`${count} ${pluralize('changeset', count)} modified`}
        >
            {`${count} modified`}
        </H4>
    </div>
)
export const PreviewStatsRemoved: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({ count }) => (
    <div className={classNames(styles.batchChangePreviewStatsBarStat, 'd-flex flex-column mx-2 text-nowrap')}>
        <div className="d-flex flex-column align-items-center">
            <span className={styles.previewStatsRemovedLine}>&nbsp;</span>
            <span
                className={classNames(
                    styles.previewStatsRemovedIcon,
                    'd-flex justify-content-center align-items-center'
                )}
            >
                <ChangesetRemovedIcon />
            </span>
            <span className={styles.previewStatsRemovedLine}>&nbsp;</span>
        </div>
        <H4
            as={H3}
            className="font-weight-normal mt-1 mb-0"
            aria-label={`${count} ${pluralize('changeset', count)} removed`}
        >
            {`${count} removed`}
        </H4>
    </div>
)
