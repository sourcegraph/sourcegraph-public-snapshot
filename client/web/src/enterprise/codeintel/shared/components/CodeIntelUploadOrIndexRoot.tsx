import { FunctionComponent } from 'react'

import { Link } from '@sourcegraph/wildcard'

import { LsifIndexFields, LsifUploadFields } from '../../../../graphql-operations'

export interface CodeIntelUploadOrIndexRootProps {
    node: Pick<LsifUploadFields | LsifIndexFields, 'projectRoot' | 'inputRoot'>
}

export const CodeIntelUploadOrIndexRoot: FunctionComponent<
    React.PropsWithChildren<CodeIntelUploadOrIndexRootProps>
> = ({ node }) =>
    node.projectRoot ? (
        <Link to={node.projectRoot.url}>
            <strong>{node.projectRoot.path || '/'}</strong>
        </Link>
    ) : (
        <span>{node.inputRoot || '/'}</span>
    )
