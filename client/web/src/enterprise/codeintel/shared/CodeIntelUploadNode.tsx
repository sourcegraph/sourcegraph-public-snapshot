import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelState } from './CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from './CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from './CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from './CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from './CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from './CodeIntelUploadOrIndexRoot'

export interface CodeIntelUploadNodeProps {
    node: LsifUploadFields
    now?: () => Date
    summaryView?: boolean
}

export const CodeIntelUploadNode: FunctionComponent<CodeIntelUploadNodeProps> = ({
    node,
    now,
    summaryView = false,
}) => (
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
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
                </small>
            </div>
        </div>

        {!summaryView && (
            <>
                <span className="d-none d-md-inline codeintel-upload-node__state">
                    <CodeIntelState node={node} />
                </span>
                <span>
                    <Link to={`./uploads/${node.id}`}>
                        <ChevronRightIcon />
                    </Link>
                </span>
            </>
        )}
    </>
)
