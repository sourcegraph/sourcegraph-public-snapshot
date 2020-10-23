import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadRootProps {
    node: LsifUploadFields
}

export const CodeIntelUploadRoot: FunctionComponent<CodeIntelUploadRootProps> = ({ node }) =>
    node.projectRoot ? (
        <Link to={node.projectRoot.url}>
            <strong>{node.projectRoot.path || '/'}</strong>
        </Link>
    ) : (
        <span>{node.inputRoot || '/'}</span>
    )
