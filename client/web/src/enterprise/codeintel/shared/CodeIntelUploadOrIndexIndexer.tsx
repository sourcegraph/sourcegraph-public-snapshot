import React, { FunctionComponent } from 'react'

export interface CodeIntelUploadOrIndexIndexerProps {
    node: { inputIndexer?: string }
}

export const CodeIntelUploadOrIndexIndexer: FunctionComponent<CodeIntelUploadOrIndexIndexerProps> = ({ node }) => (
    <span>{node.inputIndexer}</span>
)
