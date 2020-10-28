import React, { FunctionComponent } from 'react'
import { PageTitle } from '../../../components/PageTitle'

export interface CodeIntelIndexPageTitleProps {}

export const CodeIntelIndexPageTitle: FunctionComponent<CodeIntelIndexPageTitleProps> = () => (
    <>
        <PageTitle title="Precise code intelligence auto-index records" />
        <h2>Precise code intelligence auto-index records</h2>
        <p>
            Popular Go repositories are indexed automatically via{' '}
            <a href="https://github.com/sourcegraph/lsif-go" target="_blank" rel="noreferrer noopener">
                lsif-go
            </a>{' '}
            on{' '}
            <a href="https://sourcegraph.com" target="_blank" rel="noreferrer noopener">
                Sourcegraph.com
            </a>
            . Enable precise code intelligence for non-Go code by{' '}
            <a
                href="https://docs.sourcegraph.com/code_intelligence/precise_code_intelligence"
                target="_blank"
                rel="noreferrer noopener"
            >
                uploading LSIF data
            </a>
            .
        </p>
    </>
)
