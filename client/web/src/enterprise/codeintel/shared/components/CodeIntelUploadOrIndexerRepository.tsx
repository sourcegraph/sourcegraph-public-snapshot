import React, { FunctionComponent } from 'react'

import { RouterLink } from '@sourcegraph/wildcard'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexRepositoryProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot'>
}

export const CodeIntelUploadOrIndexRepository: FunctionComponent<CodeIntelUploadOrIndexRepositoryProps> = ({ node }) =>
    node.projectRoot ? (
        <RouterLink to={node.projectRoot.repository.url}>{node.projectRoot.repository.name}</RouterLink>
    ) : (
        <span>unknown</span>
    )
