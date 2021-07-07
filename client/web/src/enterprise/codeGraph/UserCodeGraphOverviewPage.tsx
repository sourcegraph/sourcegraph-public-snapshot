import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { requestGraphQL } from '../../backend/graphql'
import {
    Scalars,
    UserCodeGraphVariables,
    UserCodeGraphResult,
    UserCodeGraphOverviewData,
} from '../../graphql-operations'

const userCodeGraphOverDataGQLFragment = gql`
    fragment UserCodeGraphOverviewData on User {
        codeGraph {
            dependencies
            dependents
        }
    }
`

const queryUserCodeGraph = (vars: UserCodeGraphVariables): Observable<UserCodeGraphOverviewData | null> =>
    requestGraphQL<UserCodeGraphResult, UserCodeGraphVariables>(
        gql`
            query UserCodeGraph($userID: ID!) {
                node(id: $userID) {
                    ... on User {
                        ...UserCodeGraphOverviewData
                    }
                }
            }
            ${userCodeGraphOverDataGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node || null)
    )

interface Props extends ThemeProps, ExtensionsControllerProps, TelemetryProps, PlatformContextProps {
    namespaceID: Scalars['ID']
}

export const UserCodeGraphOverviewPage: React.FunctionComponent<Props> = ({ namespaceID, ...props }) => {
    const codeGraphPersonNode = useObservable(useMemo(() => queryUserCodeGraph({ userID: namespaceID }), [namespaceID]))

    return codeGraphPersonNode === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : codeGraphPersonNode === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <div className="pb-3">{JSON.stringify(codeGraphPersonNode, null, 2)}</div>
    )
}
