import { type FC, useEffect } from 'react'

import { mdiServer } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H3, PageHeader, Text, ErrorAlert, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { humanizeSize } from '../util/size'

import { useGitserversConnection } from './backend'

import styles from './SiteAdminGitserversPage.module.scss'

export interface GitserversPageProps extends TelemetryProps {}

export const SiteAdminGitserversPage: FC<GitserversPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminGitserversPage')
    }, [telemetryService])

    const { data, loading, error } = useGitserversConnection()

    return (
        <div>
            <PageTitle title="Gitservers" />
            <PageHeader
                path={[{ icon: mdiServer }, { text: 'Gitservers' }]}
                className="mb-3"
                headingElement="h2"
                description="Overview of all gitserver instances, on which this Sourcegraph instance stores source code."
            />

            {error && <ErrorAlert error={error} />}
            {loading && <LoadingSpinner />}
            {!loading && data && (
                <ul className={styles.nodeWrapper}>
                    {data.gitservers.nodes.map((gitserverInfo, index) => {
                        const serverName = `Server ${index + 1}`
                        return (
                            <li key={gitserverInfo.id} className={classNames('border', styles.node)}>
                                <H3>{serverName}</H3>
                                <div className="border-top mt-2">
                                    <Text className="d-flex justify-content-between mb-0 mt-3">
                                        <span className="font-weight-medium">Address</span>
                                        <span className="text-muted">{gitserverInfo.address}</span>
                                    </Text>
                                    <Text className="d-flex justify-content-between mb-0 mt-3">
                                        <span className="font-weight-medium">Free disk space</span>
                                        <span className="text-muted">
                                            {humanizeSize(Number(gitserverInfo.freeDiskSpaceBytes))}
                                        </span>
                                    </Text>
                                    <Text className="d-flex justify-content-between mb-0 mt-3">
                                        <span className="font-weight-medium">Total disk space</span>
                                        <span className="text-muted">
                                            {humanizeSize(Number(gitserverInfo.totalDiskSpaceBytes))}
                                        </span>
                                    </Text>
                                </div>
                            </li>
                        )
                    })}
                </ul>
            )}
        </div>
    )
}
