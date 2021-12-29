import React, { FunctionComponent } from 'react'

import { RouterLink } from '@sourcegraph/wildcard'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexCommitProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot' | 'inputCommit'>
    abbreviated?: boolean
}

export const CodeIntelUploadOrIndexCommit: FunctionComponent<CodeIntelUploadOrIndexCommitProps> = ({
    node,
    abbreviated = true,
}) => (
    <code>
        {node.projectRoot ? (
            <RouterLink to={node.projectRoot.commit.url}>
                <code>{abbreviated ? node.projectRoot.commit.abbreviatedOID : node.projectRoot.commit.oid}</code>
            </RouterLink>
        ) : (
            <span>{abbreviated ? node.inputCommit.slice(0, 7) : node.inputCommit}</span>
        )}
    </code>
)
