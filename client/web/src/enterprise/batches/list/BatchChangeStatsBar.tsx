import React from 'react'

import { useQuery } from '@sourcegraph/http-client'

import { ChangesetStatisticsResult, ChangesetStatisticsVariables } from '../../../graphql-operations'

import { CHANGESET_STATISTICS } from './backend'

export const BatchChangeStatsBar: React.FunctionComponent = () => {
    const { data } = useQuery<ChangesetStatisticsResult, ChangesetStatisticsVariables>(CHANGESET_STATISTICS, {})

    console.log(data)

    return <div>hello</div>
}
