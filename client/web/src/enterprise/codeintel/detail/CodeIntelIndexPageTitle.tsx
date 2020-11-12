import React, { FunctionComponent } from 'react'
import { LsifIndexFields } from '../../../graphql-operations'

export interface CodeIntelIndexPageTitleProps {
    index: LsifIndexFields
}

export const CodeIntelIndexPageTitle: FunctionComponent<CodeIntelIndexPageTitleProps> = ({ index }) => (
    <div className="mb-1">
        <h2 className="mb-0">
            <span className="text-muted">Auto-index record for commit</span>
            <span className="ml-2">
                {index.projectRoot ? index.projectRoot.commit.abbreviatedOID : index.inputCommit.slice(0, 7)}
            </span>
        </h2>
    </div>
)
