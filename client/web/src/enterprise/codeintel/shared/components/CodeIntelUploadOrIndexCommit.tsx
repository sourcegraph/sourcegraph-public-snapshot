import { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/wildcard'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexCommitProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot' | 'inputCommit'>
    abbreviated?: boolean
}

export const CodeIntelUploadOrIndexCommit: FunctionComponent<
    React.PropsWithChildren<CodeIntelUploadOrIndexCommitProps>
> = ({ node, abbreviated = true }) => (
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
