import React from 'react'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

import { UploadOrIndexMeta } from './UploadOrIndexMeta'

export interface UploadOrIndexMetaTableProps {
    prefix: string
    nodes: (LsifUploadFields | LsifIndexFields)[]
}

export const UploadOrIndexMetaTable: React.FunctionComponent<React.PropsWithChildren<UploadOrIndexMetaTableProps>> = ({
    nodes,
    prefix,
}) => (
    <table className="table">
        <thead>
            <tr>
                <th>Root</th>
                <th>Commit</th>
                <th>Indexer</th>
                <th>State</th>
                <th>LastActivity</th>
            </tr>
        </thead>
        <tbody>
            {nodes.map(node => (
                <UploadOrIndexMeta key={`${prefix}-${node.id}`} data={node} />
            ))}
        </tbody>
    </table>
)
