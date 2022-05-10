import React, { useEffect } from 'react'

import classNames from 'classnames'

import { PageHeader, CardBody, Card } from '@sourcegraph/wildcard'

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
                <p>
                    <strong>Batch Changes requirements</strong>
                </p>
                <ul className={classNames(styles.narrowList, 'mb-0')}>
                    <li>On prem installation</li>
                    <li>Unlicensed users can create 5 changesets per batch change</li>
                    <li>Enterprise plan (or trial) allows unlimited use </li>
                </ul>
            </CardBody>
        </Card>
        <CtaBanner
            bodyText="Batch Changes requires a local Sourcegraph installation. You can check it out for free by installing with a single line of code."
            title="Install locally to get started"
            linkText="Install local instance"
            href="/help/admin/install?utm_medium=inproduct&utm_source=inproduct-batch-changes&utm_campaign=inproduct-batch-changes&term="
            icon={<DownloadSourcegraphIcon />}
            className="ml-3"
            onClick={() => eventLogger.log('BatchChangesInstallFromCloudClicked')}
        />
    </div>
)

const DownloadSourcegraphIcon: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <svg width="51" height="82" viewBox="0 0 51 82" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            d="M24.2188 49.7734C24.2188 49.2212 24.6665 48.7734 25.2188 48.7734H26.5313C27.0835 48.7734 27.5312 49.2212 27.5312 49.7734V66.2898C27.5312 67.1814 28.6099 67.6273 29.2394 66.9958L35.9324 60.2823C36.3232 59.8904 36.958 59.8904 37.3488 60.2823L38.2555 61.1918C38.6446 61.5821 38.6446 62.2136 38.2555 62.6039L26.5832 74.3119C26.1924 74.7038 25.5576 74.7038 25.1668 74.3119L13.4945 62.6039C13.1054 62.2136 13.1054 61.5821 13.4945 61.1918L14.4012 60.2823C14.792 59.8904 15.4268 59.8904 15.8176 60.2823L22.5106 66.9958C23.1401 67.6273 24.2188 67.1814 24.2188 66.2898V49.7734Z"
            fill="url(#paint0_linear)"
            fillOpacity="0.23"
        />
        <rect x="9" y="76.7339" width="32.25" height="2.25688" rx="1.12844" fill="#F2F4F8" />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M16.6279 10.503L25.4817 43.0194C26.0779 45.2041 28.3225 46.4909 30.4971 45.8946C32.6717 45.2949 33.9511 43.0382 33.3567 40.8536L24.5011 8.33542C23.905 6.15075 21.6603 4.86394 19.4857 5.46193C17.3129 6.05993 16.0317 8.31657 16.6279 10.503"
            fill="#ff5543"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M33.0244 10.3795L10.8514 35.5314C9.35669 37.2277 9.51214 39.8219 11.1999 41.3246C12.8876 42.8274 15.4671 42.6697 16.9618 40.9734L39.1348 15.8214C40.6295 14.1251 40.474 11.5326 38.7863 10.0299C37.0985 8.5272 34.5208 8.68312 33.0244 10.3795"
            fill="#B200F8"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M7.90114 24.319L39.5224 34.826C41.6628 35.5388 43.9707 34.3702 44.6796 32.2198C45.3885 30.0677 44.2252 27.7476 42.0865 27.0383L10.4635 16.5262C8.32307 15.8168 6.01524 16.9819 5.30803 19.134C4.59911 21.2862 5.76071 23.6062 7.90114 24.319"
            fill="#00cbec"
        />
        <defs>
            <linearGradient
                id="paint0_linear"
                x1="25.875"
                y1="76.3575"
                x2="26.2523"
                y2="36.1098"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#889ABF" />
                <stop offset="1" stopColor="#889ABF" stopOpacity="0" />
            </linearGradient>
        </defs>
    </svg>
)
