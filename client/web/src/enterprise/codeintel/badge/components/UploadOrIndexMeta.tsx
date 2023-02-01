import React from 'react'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface UploadOrIndexMetaProps {
    data: LsifUploadFields | LsifIndexFields
    now?: () => Date
}

export const UploadOrIndexMeta: React.FunctionComponent<React.PropsWithChildren<UploadOrIndexMetaProps>> = ({
    data: node,
    now,
}) => (
    <tr>
        <td>{node.inputRoot}</td>
        <td>{node.inputCommit}</td>
        <td>{node.inputIndexer}</td>
        <td>{node.state}</td>
        <td>{node.finishedAt || node.startedAt}</td>
    </tr>
)
