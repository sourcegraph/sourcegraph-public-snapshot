import React, { useCallback } from 'react'

import { useHistory } from 'react-router'
import { Subject } from 'rxjs'
import { delay, repeatWhen } from 'rxjs/operators'

import { H3 } from '@sourcegraph/wildcard'

import {
    ExternalServiceSyncJobConnectionFields,
    ExternalServiceSyncJobListFields,
    Scalars,
} from '../../graphql-operations'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../FilteredConnection'

import { queryExternalServiceSyncJobs as _queryExternalServiceSyncJobs } from './backend'
import { ExternalServiceSyncJobNode, ExternalServiceSyncJobNodeProps } from './ExternalServiceSyncJobNode'

interface ExternalServiceSyncJobsListProps {
    externalServiceID: Scalars['ID']
    updates: Subject<void>

    /** For testing only. */
    queryExternalServiceSyncJobs?: typeof _queryExternalServiceSyncJobs
}

export const ExternalServiceSyncJobsList: React.FunctionComponent<ExternalServiceSyncJobsListProps> = ({
    externalServiceID,
    updates,
    queryExternalServiceSyncJobs = _queryExternalServiceSyncJobs,
}) => {
    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryExternalServiceSyncJobs({
                first: args.first ?? null,
                externalService: externalServiceID,
            }).pipe(repeatWhen(obs => obs.pipe(delay(1500)))),
        [externalServiceID, queryExternalServiceSyncJobs]
    )

    const history = useHistory()

    return (
        <>
            <H3 className="mt-3">Recent sync jobs</H3>
            <FilteredConnection<
                ExternalServiceSyncJobListFields,
                Omit<ExternalServiceSyncJobNodeProps, 'node'>,
                {},
                ExternalServiceSyncJobConnectionFields
            >
                className="mb-0"
                listClassName="list-group list-group-flush mb-0"
                noun="sync job"
                pluralNoun="sync jobs"
                queryConnection={queryConnection}
                nodeComponent={ExternalServiceSyncJobNode}
                nodeComponentProps={{ onUpdate: updates }}
                hideSearch={true}
                noSummaryIfAllNodesVisible={true}
                history={history}
                updates={updates}
                location={history.location}
            />
        </>
    )
}
