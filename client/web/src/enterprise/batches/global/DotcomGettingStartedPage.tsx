import React, { useEffect } from 'react'

import classNames from 'classnames'

import { DownloadSourcegraphIcon } from '@sourcegraph/branded/src/components/DownloadSourcegraphIcon'
import { PageHeader, CardBody, Card, Text, H3 } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CtaBanner } from '../../../components/CtaBanner'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { GettingStarted } from '../list/GettingStarted'

import styles from './DotcomGettingStartedPage.module.scss'

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
    <div className="d-flex justify-content-between">
        <Card className={classNames(styles.requirementsCard, 'mr-3')}>
            <CardBody>
                <Text>
                    <strong>Batch Changes requirements</strong>
                </Text>
                <ul className={classNames(styles.narrowList, 'mb-0')}>
                    <li>On prem installation</li>
                    <li>Unlicensed users can create 5 changesets per batch change</li>
                    <li>Enterprise plan (or trial) allows unlimited use </li>
                </ul>
            </CardBody>
        </Card>
        <CtaBanner
            bodyText="Batch Changes requires a Sourcegraph Cloud or self-hosted instance."
            title={<H3>Start using Batch Changes</H3>}
            linkText="Get started"
            href="/help/admin/install?utm_medium=inproduct&utm_source=inproduct-batch-changes&utm_campaign=inproduct-batch-changes&term="
            icon={<DownloadSourcegraphIcon />}
            className="ml-3"
            onClick={() => eventLogger.log('BatchChangesInstallFromCloudClicked')}
        />
    </div>
)
