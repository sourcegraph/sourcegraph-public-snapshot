import { FunctionComponent } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'

import { PreciseIndexFields } from '../../../../graphql-operations'

export interface CodeIntelLastUpdatedProps {
    index: PreciseIndexFields
    now?: () => Date
}

// TODO - add additional timestamps
export const PreciseIndexLastUpdated: FunctionComponent<CodeIntelLastUpdatedProps> = ({ index, now }) =>
    index.uploadedAt ? (
        <span>
            Uploaded <Timestamp date={index.uploadedAt} now={now} noAbout={true} />
        </span>
    ) : index.queuedAt ? (
        <span>
            Queued <Timestamp date={index.queuedAt} now={now} noAbout={true} />
        </span>
    ) : (
        <></>
    )
