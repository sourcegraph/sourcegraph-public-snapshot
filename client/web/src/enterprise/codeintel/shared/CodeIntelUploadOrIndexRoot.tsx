import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifIndexFields, LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadOrIndexRootProps {
    node: LsifUploadFields | LsifIndexFields
}

export const CodeIntelUploadOrIndexRoot: FunctionComponent<CodeIntelUploadOrIndexRootProps> = ({ node }) =>
    node.projectRoot ? (
        <Link to={node.projectRoot.url}>
            <strong>{node.projectRoot.path || '/'}</strong>
        </Link>
    ) : (
        <span>{node.inputRoot || '/'}</span>
    )
