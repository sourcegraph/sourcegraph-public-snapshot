import React, { useMemo } from 'react'
import { Card, CardHeader, CardTitle, ListGroup, ListGroupItem, ListGroupItemHeading } from 'reactstrap'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
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
import { PersonLink, personLinkFieldsFragment } from '../../person/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'

const userCodeGraphOverDataGQLFragment = gql`
    fragment UserCodeGraphOverviewData on User {
        avatarURL
        displayName
        username

        codeGraph {
            symbols

            dependencies {
                id
                name
                url
            }
            callers {
                ...PersonLinkFields
            }

            dependents {
                id
                name
                url
            }
            callers {
                ...PersonLinkFields
            }
        }
    }
    ${personLinkFieldsFragment}
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
                    <h2 className="text-center mb-4">Dependencies</h2>
                    <h4 className="border-bottom pb-1">Packages you use</h4>
                    <ListGroup className="mb-4">
                        {codeGraphPersonNode.codeGraph.dependencies.map(dep => (
                            <ListGroupItem key={dep.id}>
                                <RepoLink repoName={dep.name} to={dep.url} />
                            </ListGroupItem>
                        ))}
                    </ListGroup>

                    <h4 className="border-bottom pb-1 mt-4">Authors whose code you use</h4>
                    {codeGraphPersonNode.codeGraph.authors.map(author => (
                        <ListGroupItem key={author.email}>
                            <PersonLink person={author} />
                        </ListGroupItem>
                    ))}
                </Container>
                <div className="col-4 d-flex flex-column justify-content-center px-4">
                    <hr
                        className="border-white"
                        // eslint-disable-next-line react/forbid-dom-props
                        style={{
                            position: 'relative',
                            top: '52px',
                            zIndex: -100,
                            marginLeft: '-30px',
                            marginRight: '-30px',
                            width: '150%',
                            opacity: '0.5',
                        }}
                    />
                    <h3 className="text-center h2">
                        <UserAvatar user={codeGraphPersonNode} className="icon-inline h2" size={100} />
                    </h3>
                    <Card className="mx-2 bg-transparent">
                        <CardHeader className="text-center bg-transparent h5 font-weight-bold mb-0">
                            Contributions
                        </CardHeader>
                        <ListGroup flush={true}>
                            {codeGraphPersonNode.codeGraph.symbols.slice(0, 10).map(symbol => (
                                <ListGroupItem key={symbol} className="text-truncate text-monospace small">
                                    {symbol}
                                </ListGroupItem>
                            ))}
                        </ListGroup>
                    </Card>
                </div>
                <Container className="col-4">
                    <h2 className="text-center mb-4">Dependents</h2>
                    <h4 className="">Packages using your code</h4>
                    <ListGroup className="mb-4">
                        {codeGraphPersonNode.codeGraph.dependents.map(dep => (
                            <ListGroupItem key={dep.id}>
                                <RepoLink repoName={dep.name} to={dep.url} />
                            </ListGroupItem>
                        ))}
                    </ListGroup>

                    <h4 className="">People using your code</h4>
                    <ListGroup>
                        {codeGraphPersonNode.codeGraph.callers.map(caller => (
                            <ListGroupItem key={caller.email}>
                                <PersonLink person={caller} />
                            </ListGroupItem>
                        ))}
                    </ListGroup>
                </Container>
            </div>
        </div>
    )
}
