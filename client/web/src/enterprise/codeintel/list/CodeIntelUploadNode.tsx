import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadRoot } from '../shared/CodeIntelUploadRoot'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeInteUploadOrIndexerRepository'
import { CodeIntelState } from './CodeIntelState'

export interface CodeIntelUploadNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

export const CodeIntelUploadNode: FunctionComponent<CodeIntelUploadNodeProps> = ({ node, now }) => (
    <>
        <span className="codeintel-upload-node__separator" />

        <div className="d-flex flex-column codeintel-upload-node__information">
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
                </small>
            </div>
        </div>
        <span className="d-none d-md-inline codeintel-upload-node__state">
            <CodeIntelState node={node} />
        </span>
        <span>
            <Link to={`./uploads/${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)
