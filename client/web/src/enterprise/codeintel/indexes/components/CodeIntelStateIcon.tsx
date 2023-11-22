import type { FunctionComponent } from 'react'

import { mdiAlertCircle, mdiCheckCircle, mdiDatabase, mdiFileUpload, mdiSourceRepository, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { PreciseIndexState } from '../../../../graphql-operations'

export interface CodeIntelStateIconProps {
    state: PreciseIndexState
    autoIndexed: boolean
    className?: string
}

export const CodeIntelStateIcon: FunctionComponent<CodeIntelStateIconProps> = ({ state, autoIndexed, className }) =>
    state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
        <div className="text-center">
            <Icon className={className} svgPath={mdiTimerSand} inline={false} aria-label="Queued" />
            <Icon className={className} svgPath={mdiDatabase} inline={false} aria-label="Queued for processing" />
        </div>
    ) : state === PreciseIndexState.PROCESSING ? (
        <LoadingSpinner inline={false} className={className} />
    ) : state === PreciseIndexState.PROCESSING_ERRORED ? (
        <Icon
            className={classNames('text-danger', className)}
            svgPath={mdiAlertCircle}
            inline={false}
            aria-label="Errored"
        />
    ) : state === PreciseIndexState.COMPLETED ? (
        <Icon
            className={classNames('text-success', className)}
            svgPath={mdiCheckCircle}
            inline={false}
            aria-label="Completed"
        />
    ) : state === PreciseIndexState.UPLOADING_INDEX ? (
        <Icon className={className} svgPath={mdiFileUpload} inline={false} aria-label="Uploading" />
    ) : state === PreciseIndexState.DELETING ? (
        <Icon
            className={classNames('text-muted', className)}
            svgPath={mdiCheckCircle}
            inline={false}
            aria-label="Deleting"
        />
    ) : state === PreciseIndexState.QUEUED_FOR_INDEXING ? (
        <div className="text-center">
            <Icon className={className} svgPath={mdiTimerSand} inline={false} aria-label="Queued" />
            <Icon className={className} svgPath={mdiSourceRepository} inline={false} aria-label="Queued for indexing" />
        </div>
    ) : state === PreciseIndexState.INDEXING ? (
        <LoadingSpinner inline={false} className={className} />
    ) : state === PreciseIndexState.INDEXING_ERRORED ? (
        <Icon
            className={classNames('text-danger', className)}
            svgPath={mdiAlertCircle}
            inline={false}
            aria-label="Errored"
        />
    ) : autoIndexed ? (
        <Icon
            className={classNames('text-success', className)}
            svgPath={mdiCheckCircle}
            inline={false}
            aria-label="Completed"
        />
    ) : (
        <></>
    )
