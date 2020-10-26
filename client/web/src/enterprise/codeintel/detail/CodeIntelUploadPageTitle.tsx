import React, { FunctionComponent } from 'react'
import { LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadPageTitleProps {
    upload: LsifUploadFields
}

export const CodeIntelUploadPageTitle: FunctionComponent<CodeIntelUploadPageTitleProps> = ({ upload }) => (
    <div className="mb-1">
        <h2 className="mb-0">
            <span className="text-muted">Upload for commit</span>
            <span className="ml-2">
                {upload.projectRoot ? upload.projectRoot.commit.abbreviatedOID : upload.inputCommit.slice(0, 7)}
            </span>
            <span className="ml-2 text-muted">indexed by</span>
            <span className="ml-2">{upload.inputIndexer}</span>
            <span className="ml-2 text-muted">rooted at</span>
            <span className="ml-2">{(upload.projectRoot ? upload.projectRoot.path : upload.inputRoot) || '/'}</span>
        </h2>
    </div>
)
