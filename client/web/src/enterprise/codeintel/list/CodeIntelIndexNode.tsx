import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifIndexFields } from '../../../graphql-operations'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeInteUploadOrIndexerRepository'
import { CodeIntelState } from './CodeIntelState'

export interface CodeIntelIndexNodeProps {
    node: LsifIndexFields
    now?: () => Date
}

export const CodeIntelIndexNode: FunctionComponent<CodeIntelIndexNodeProps> = ({ node, now }) => (
    <>
        <span className="codeintel-index-node__separator" />

        <div className="d-flex flex-column codeintel-index-node__information">
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Indexed at commit <CodeIntelUploadOrIndexCommit node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, uploadedAt: null }} now={now} />
                </small>
            </div>
        </div>
        <span className="d-none d-md-inline codeintel-index-node__state">
            <CodeIntelState node={node} />
        </span>
        <span>
            <Link to={`./indexes/${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)
