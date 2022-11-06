import React, { useEffect } from 'react'

import { DownloadSourcegraphIcon } from '@sourcegraph/branded/src/components/DownloadSourcegraphIcon'
import { PageHeader, H3 } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CtaBanner } from '../../../components/CtaBanner'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { GettingStarted } from '../list/GettingStarted'

export interface DotcomGettingStartedPageProps {
    // Nothing for now.
}

export const DotcomGettingStartedPage: React.FunctionComponent<
    React.PropsWithChildren<DotcomGettingStartedPageProps>
> = () => {
    useEffect(() => eventLogger.logViewEvent('BatchChangesCloudLandingPage'), [])

    return (
        <Page>
            <PageTitle title="Batch Changes" />
            <PageHeader
                path={[{ icon: BatchChangesIcon, text: 'Batch Changes' }]}
                description="Run custom code over hundreds of repositories and manage the resulting changesets."
                className="mb-3"
            />
            <GettingStarted footer={<DotcomGettingStartedPageFooter />} />
        </Page>
    )
}

const DotcomGettingStartedPageFooter: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="d-flex justify-content-start">
        <CtaBanner
            bodyText="Batch Changes requires a Sourcegraph Cloud or self-hosted instance."
            title={<H3>Start using Batch Changes</H3>}
            linkText="Get started"
            href="/help/admin/install?utm_medium=inproduct&utm_source=inproduct-batch-changes&utm_campaign=inproduct-batch-changes&term="
            icon={<DownloadSourcegraphIcon />}
            onClick={() => eventLogger.log('BatchChangesInstallFromCloudClicked')}
        />
    </div>
)
