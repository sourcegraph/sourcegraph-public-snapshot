import React, { FunctionComponent } from 'react'

export interface CodeIntelUploadOrIndexIndexerProps {
    node: { indexer?: string; inputIndexer?: string }
}

export const CodeIntelUploadOrIndexIndexer: FunctionComponent<CodeIntelUploadOrIndexIndexerProps> = ({ node }) => (
    <span>{node.indexer || node.inputIndexer}</span>
)
