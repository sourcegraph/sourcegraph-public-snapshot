import React, { FunctionComponent } from 'react'

import { RouterLink } from '@sourcegraph/wildcard'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexRootProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot' | 'inputRoot'>
}

export const CodeIntelUploadOrIndexRoot: FunctionComponent<CodeIntelUploadOrIndexRootProps> = ({ node }) =>
    node.projectRoot ? (
        <RouterLink to={node.projectRoot.url}>
            <strong>{node.projectRoot.path || '/'}</strong>
        </RouterLink>
    ) : (
        <span>{node.inputRoot || '/'}</span>
    )
