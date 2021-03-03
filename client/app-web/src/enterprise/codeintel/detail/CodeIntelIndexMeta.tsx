import React, { FunctionComponent } from 'react'
import { LsifIndexFields } from '../../../graphql-operations'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

export interface CodeIntelIndexMetaProps {
    node: LsifIndexFields
    now?: () => Date
}

export const CodeIntelIndexMeta: FunctionComponent<CodeIntelIndexMetaProps> = ({ node, now }) => (
    <>
        <div className="card border-0">
            <div className="card-body">
                <h3 className="card-title">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>

                <p className="card-subtitle mb-2 text-muted">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, uploadedAt: null }} now={now} />
                </p>

                <p className="card-text">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </p>
            </div>
        </div>
    </>
)
