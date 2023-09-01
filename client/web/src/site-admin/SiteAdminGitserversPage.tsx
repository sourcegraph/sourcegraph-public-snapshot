import { type FC, useEffect } from 'react'

import { mdiServer, mdiPuzzle } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, Icon, PageHeader, Text } from '@sourcegraph/wildcard'

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
    {
        id: '3',
        address: '127.0.0.1:3503',
        freeDiskSpaceBytes: 20293000,
        totalDiskSpaceBytes: 1059302032,
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
                {gitserversMockData.map((gitserverInfo, index) => {
                    const serverName = `Server ${index + 1}`
                    return (
                        <li key={gitserverInfo.id} className={classNames('border', styles.node)}>
                            <div className="d-flex justify-content-center flex-column align-items-center">
                                <div className={styles.nodeIcon}>
                                    <Icon aria-hidden={true} svgPath={mdiPuzzle} size="md" />
                                </div>
                                <Text className="mb-0">{serverName}</Text>
                            </div>
                            <div className="border-top mt-3">
                                <Text className="d-flex justify-content-between mt-3">
                                    <span className="font-weight-bold">Address</span>
                                    <span>{gitserverInfo.address}</span>
                                </Text>
                                <Text className="d-flex justify-content-between">
                                    <span className="font-weight-bold">Free disk space</span>
                                    <span>{humanizeSize(gitserverInfo.freeDiskSpaceBytes)}</span>
                                </Text>
                                <Text className="d-flex justify-content-between mb-0">
                                    <span className="font-weight-bold">Total disk space</span>
                                    <span>{humanizeSize(gitserverInfo.totalDiskSpaceBytes)}</span>
                                </Text>
                            </div>
                        </li>
                    )
                })}
            </ul>
        </Container>
    )
}
