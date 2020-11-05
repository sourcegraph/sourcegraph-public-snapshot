import React, { FunctionComponent } from 'react'
import { PageTitle } from '../../../components/PageTitle'

export interface CodeIntelUploadsPageTitleProps {}

export const CodeIntelUploadsPageTitle: FunctionComponent<CodeIntelUploadsPageTitleProps> = () => (
    <>
        <PageTitle title="Precise code intelligence uploads" />
        <h2>Precise code intelligence uploads</h2>
        <p>
            Enable precise code intelligence by{' '}
            <a
                href="https://docs.sourcegraph.com/code_intelligence/precise_code_intelligence"
                target="_blank"
                rel="noreferrer noopener"
            >
                uploading LSIF data
            </a>
            .
        </p>

        <p>
            Current uploads provide code intelligence for the latest commit on the default branch and are used in
            cross-repository <em>Find References</em> requests. Non-current uploads may still provide code intelligence
            for historic and branch commits.
        </p>
    </>
)
