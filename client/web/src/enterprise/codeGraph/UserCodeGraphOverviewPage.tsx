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
import { Container } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import {
    Scalars,
    UserCodeGraphVariables,
    UserCodeGraphResult,
    UserCodeGraphOverviewData,
} from '../../graphql-operations'
import { UserAvatar } from '../../user/UserAvatar'

const userCodeGraphOverDataGQLFragment = gql`
    fragment UserCodeGraphOverviewData on User {
        avatarURL
        displayName
        username

        codeGraph {
            symbols
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
        <div className="pb-3 mt-5 mx-3">
            <div className="row">
                <Container className="col-4">
                    <h2 className="text-center">Dependencies</h2>
                    <h4 className="border-bottom pb-1">Packages</h4>
                    {codeGraphPersonNode.codeGraph.dependencies.join(' ')}
                    <h4 className="border-bottom pb-1 mt-4">Authors</h4>
                </Container>
                <div className="col-4 d-flex flex-column justify-content-center px-4">
                    <h3 className="text-center h2">
                        <UserAvatar user={codeGraphPersonNode} className="icon-inline" />
                    </h3>
                    <h4 className="text-center">Contributions</h4>
                    {codeGraphPersonNode.codeGraph.symbols.join(' ')}
                </div>
                <Container className="col-4">
                    <h2 className="text-center">Dependents</h2>
                    <h4 className="border-bottom pb-1">Packages</h4>
                    {codeGraphPersonNode.codeGraph.dependencies.join(' ')}
                    <h4 className="border-bottom pb-1 mt-4">Callers</h4>
                </Container>
            </div>
        </div>
    )
}
