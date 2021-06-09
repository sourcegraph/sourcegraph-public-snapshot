import React, { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { LsifIndexFields, LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadOrIndexRepositoryProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot'>
}

export const CodeIntelUploadOrIndexRepository: FunctionComponent<CodeIntelUploadOrIndexRepositoryProps> = ({ node }) =>
    node.projectRoot ? (
        <Link to={node.projectRoot.repository.url}>{node.projectRoot.repository.name}</Link>
    ) : (
        <span>unknown</span>
    )
