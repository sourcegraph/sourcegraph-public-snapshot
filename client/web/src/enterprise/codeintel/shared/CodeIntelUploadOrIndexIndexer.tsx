import React, { FunctionComponent } from 'react'

import { LsifIndexFields } from '../../../graphql-operations'

export interface CodeIntelUploadOrIndexIndexerProps {
    node: Partial<Pick<LsifIndexFields, 'inputIndexer'>>
}

export const CodeIntelUploadOrIndexIndexer: FunctionComponent<CodeIntelUploadOrIndexIndexerProps> = ({ node }) => (
    <span>{node.inputIndexer}</span>
)
