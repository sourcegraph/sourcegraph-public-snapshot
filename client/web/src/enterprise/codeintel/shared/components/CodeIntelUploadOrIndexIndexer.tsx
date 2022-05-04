import { FunctionComponent } from 'react'

import { LsifIndexFields } from '../../../../graphql-operations'

import { CodeIntelIndexer } from './CodeIntelIndexer'

export interface CodeIntelUploadOrIndexIndexerProps {
    node: Partial<Pick<LsifIndexFields, 'indexer'>>
}

export const CodeIntelUploadOrIndexIndexer: FunctionComponent<
    React.PropsWithChildren<CodeIntelUploadOrIndexIndexerProps>
> = ({ node }) => <span>{node.indexer && <CodeIntelIndexer indexer={node.indexer} />}</span>
