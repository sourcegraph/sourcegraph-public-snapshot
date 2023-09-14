import type { FunctionComponent } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'

import type { PreciseIndexFields } from '../../../../graphql-operations'

export interface CodeIntelLastUpdatedProps {
    index: PreciseIndexFields
    now?: () => Date
}

export const PreciseIndexLastUpdated: FunctionComponent<CodeIntelLastUpdatedProps> = ({ index, now }) =>
    index.processingFinishedAt ? (
        <span>
            Completed <Timestamp date={index.processingFinishedAt} now={now} noAbout={true} />
        </span>
    ) : index.processingStartedAt ? (
        <span>
            Processing started <Timestamp date={index.processingStartedAt} now={now} noAbout={true} />
        </span>
    ) : index.uploadedAt ? (
        <span>
            Uploaded <Timestamp date={index.uploadedAt} now={now} noAbout={true} />
        </span>
    ) : index.indexingFinishedAt ? (
        <span>
            Indexed <Timestamp date={index.indexingFinishedAt} now={now} noAbout={true} />
        </span>
    ) : index.indexingStartedAt ? (
        <span>
            Indexing started <Timestamp date={index.indexingStartedAt} now={now} noAbout={true} />
        </span>
    ) : index.queuedAt ? (
        <span>
            Queued <Timestamp date={index.queuedAt} now={now} noAbout={true} />
        </span>
    ) : (
        <></>
    )
