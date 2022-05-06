import { FunctionComponent } from 'react'

import { Timestamp } from '../../../../components/time/Timestamp'
import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexLastActivityProps {
    node: Pick<
        (LsifUploadFields & { queuedAt: null }) | (LsifIndexFields & { uploadedAt: null }),
        'queuedAt' | 'uploadedAt' | 'startedAt' | 'finishedAt'
    >
    now?: () => Date
}

export const CodeIntelUploadOrIndexLastActivity: FunctionComponent<
    React.PropsWithChildren<CodeIntelUploadOrIndexLastActivityProps>
> = ({ node, now }) =>
    node.finishedAt ? (
        <span>
            Completed <Timestamp date={node.finishedAt} now={now} noAbout={true} />
        </span>
    ) : node.startedAt ? (
        <span>
            Started <Timestamp date={node.startedAt} now={now} noAbout={true} />
        </span>
    ) : node.uploadedAt ? (
        <span>
            Uploaded <Timestamp date={node.uploadedAt} now={now} noAbout={true} />
        </span>
    ) : node.queuedAt ? (
        <span>
            Queued <Timestamp date={node.queuedAt} now={now} noAbout={true} />
        </span>
    ) : (
        <></>
    )
