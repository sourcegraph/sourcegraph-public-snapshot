import * as React from 'react'
import {useEffect} from 'react'
import classNames from 'classnames'
import {logger} from '@sourcegraph/common'
import {useQuery} from '@sourcegraph/http-client'
import {TelemetryProps} from '@sourcegraph/shared/src/telemetry/telemetryService'
import {ErrorAlert, LoadingSpinner} from '@sourcegraph/wildcard'
import {FetchOwnershipResult, FetchOwnershipVariables,} from '../../../graphql-operations'
import {FETCH_OWNERS} from './grapqlQueries'

import styles from './FileOwnershipPanel.module.scss'
import {OwnerList} from "./OwnerList";

export const FileOwnershipPanel: React.FunctionComponent<
    {
        repoID: string
        revision?: string
        filePath: string
    } & TelemetryProps
> = ({ repoID, revision, filePath, telemetryService }) => {
    useEffect(() => {
        telemetryService.log('OwnershipPanelOpened')
    }, [telemetryService])

    const { data, loading, error } = useQuery<FetchOwnershipResult, FetchOwnershipVariables>(FETCH_OWNERS, {
        variables: {
            repo: repoID,
            revision: revision ?? '',
            currentPath: filePath,
        },
    })

    if (loading) {
        return (
            <div className={classNames(styles.loaderWrapper, 'text-muted')}>
                <LoadingSpinner inline={true} className="mr-1" /> Loading...
            </div>
        )
    }

    if (error) {
        logger.log(error)
        return (
            <div className={styles.contents}>
                <ErrorAlert error={error} prefix="Error getting ownership data" className="mt-2" />
            </div>
        )
    }

    if (data?.node?.__typename === 'Repository') {
        return <OwnerList data={data?.node?.commit?.blob?.ownership} />
    }
    return <OwnerList />
}

