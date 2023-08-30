import { type FC, useEffect } from 'react'

import { mdiServer } from '@mdi/js'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { humanizeSize } from '../util/size'

import styles from './SiteAdminGitserversPage.module.scss'

const gitserversMockData = [
    {
        id: '1',
        address: '127.0.0.1:3501',
        freeDiskSpaceBytes: 102930,
        totalDiskSpaceBytes: 4059302032,
    },
    {
        id: '2',
        address: '127.0.0.1:3502',
        freeDiskSpaceBytes: 202930,
        totalDiskSpaceBytes: 3059302032,
    },
]

export interface GitserversPageProps extends TelemetryProps {}

export const SiteAdminGitserversPage: FC<GitserversPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminGitserversPage')
    }, [telemetryService])

    return (
        <Container>
            <PageTitle title="Gitservers" />
            <PageHeader
                path={[{ icon: mdiServer }, { text: 'Gitservers' }]}
                className="mb-3"
                headingElement="h2"
                description="Manage your Gitservers"
            />

            <ul className={styles.nodeWrapper}>
                {gitserversMockData.map(gitserverInfo => (
                    <li key={gitserverInfo.id} className="border mt-3 p-3">
                        <Text>
                            <span className="font-weight-bold">Address:</span> {gitserverInfo.address}
                        </Text>
                        <Text>
                            <span className="font-weight-bold">Free disk space:</span>{' '}
                            {humanizeSize(gitserverInfo.freeDiskSpaceBytes)}
                        </Text>
                        <Text className="mb-0">
                            <span className="font-weight-bold">Total disk space:</span>{' '}
                            {humanizeSize(gitserverInfo.totalDiskSpaceBytes)}
                        </Text>
                    </li>
                ))}
            </ul>
        </Container>
    )
}
