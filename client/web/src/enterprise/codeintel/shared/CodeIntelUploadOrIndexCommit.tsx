import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifIndexFields, LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadOrIndexCommitProps {
    node: LsifUploadFields | LsifIndexFields
    abbreviated?: boolean
}

export const CodeIntelUploadOrIndexCommit: FunctionComponent<CodeIntelUploadOrIndexCommitProps> = ({
    node,
    abbreviated = true,
}) => (
    <code>
        {node.projectRoot ? (
            <Link to={node.projectRoot.commit.url}>
                <code>{abbreviated ? node.projectRoot.commit.abbreviatedOID : node.projectRoot.commit.oid}</code>
            </Link>
        ) : (
            <span>{abbreviated ? node.inputCommit.slice(0, 7) : node.inputCommit}</span>
        )}
    </code>
)
