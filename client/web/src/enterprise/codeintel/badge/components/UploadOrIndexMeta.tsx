import React from 'react'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'
import { CodeIntelStateIcon } from '../../shared/components/CodeIntelStateIcon'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'

export interface UploadOrIndexMetaProps {
    data: LsifUploadFields | LsifIndexFields
    now?: () => Date
}

export const UploadOrIndexMeta: React.FunctionComponent<React.PropsWithChildren<UploadOrIndexMetaProps>> = ({
    data: node,
    now,
}) => (
    <tr>
        <td>
            <CodeIntelUploadOrIndexRoot node={node} />
        </td>
        <td>
            <CodeIntelUploadOrIndexCommit node={node} />
        </td>
        <td>
            <CodeIntelUploadOrIndexIndexer node={node} />
        </td>
        <td>
            <CodeIntelStateIcon state={node.state} />
        </td>
        <td>
            <CodeIntelUploadOrIndexLastActivity node={{ uploadedAt: null, queuedAt: null, ...node }} now={now} />
        </td>
    </tr>
)
