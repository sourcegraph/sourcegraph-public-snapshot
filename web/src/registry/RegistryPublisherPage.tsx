import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { isEqual } from 'lodash'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { extensionIDPrefix } from './extension'
import { RegistryAreaPageProps } from './RegistryArea'
import { ExtensionsListViewMode, RegistryExtensionsList } from './RegistryExtensionsPage'

interface RouteParams {
    publisherType: 'users' | 'organizations'
    publisherName: string
}

function queryPublisher(args: RouteParams): Observable<GQL.RegistryPublisher | null> {
    return (args.publisherType === 'users'
        ? queryGraphQL(
              gql`
                  query RegistryPublisher($username: String!) {
                      user(username: $username) {
                          __typename
                          id
                          username
                          displayName
                          url
                          configuredExtensions {
                              url
                          }
                          registryExtensions {
                              totalCount
                          }
                      }
                  }
              `,
              { username: args.publisherName }
          )
        : queryGraphQL(
              gql`
                  query RegistryPublisher($name: String!) {
                      organization(name: $name) {
                          __typename
                          id
                          name
                          displayName
                          url
                          configuredExtensions {
                              url
                          }
                          registryExtensions {
                              totalCount
                          }
                      }
                  }
              `,
              { name: args.publisherName }
          )
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (!data.user && !data.organization) || errors) {
                throw createAggregateError(errors)
            }
            return data.user || data.organization
        })
    )
}

interface Props extends RegistryAreaPageProps, RouteComponentProps<RouteParams> {}

interface State {
    /** The publisher, undefined while loading, or an error.  */
    publisherOrError?: GQL.RegistryPublisher | ErrorLike
}

/**
 * A page for a publisher of registry extensions.
 */
export class RegistryPublisherPage extends React.Component<Props> {
    public state: State = {}

    private routeMatchChanges = new Subject<RouteParams>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Changes to the route params.
        const routeParamChanges = this.routeMatchChanges.pipe(distinctUntilChanged(isEqual))

        // Fetch extension.
        this.subscriptions.add(
            combineLatest(routeParamChanges, merge(this.refreshRequests.pipe(mapTo(true)), of(false)))
                .pipe(
                    switchMap(([routeParams, forceRefresh]) => {
                        type PartialStateUpdate = Pick<State, 'publisherOrError'>
                        return queryPublisher(routeParams).pipe(
                            catchError(error => [error]),
                            map(c => ({ publisherOrError: c } as PartialStateUpdate)),

                            // Don't clear old data while we reload, to avoid unmounting all components during
                            // loading.
                            startWith<PartialStateUpdate>(forceRefresh ? {} : { publisherOrError: undefined })
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.routeMatchChanges.next(this.props.match.params)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.match.params !== this.props.match.params) {
            this.routeMatchChanges.next(props.match.params)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.publisherOrError) {
            return null // loading
        }
        if (isErrorLike(this.state.publisherOrError)) {
            return (
                <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.publisherOrError.message)} />
            )
        }

        return (
            <div className="registry-publisher-page">
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <h2 className="mr-sm-2 mb-0">
                        Extensions published by{' '}
                        <Link to={this.state.publisherOrError.url}>
                            {extensionIDPrefix(this.state.publisherOrError)}
                        </Link>
                    </h2>
                    <div>
                        {this.state.publisherOrError &&
                            this.state.publisherOrError.configuredExtensions.url && (
                                <Link
                                    className="btn btn-outline-link"
                                    to={this.state.publisherOrError.configuredExtensions.url}
                                >
                                    Extensions used by {extensionIDPrefix(this.state.publisherOrError)}
                                </Link>
                            )}
                    </div>
                </div>
                <RegistryExtensionsList
                    {...this.props}
                    mode={ExtensionsListViewMode.Cards}
                    publisher={this.state.publisherOrError}
                    showUserActions={true}
                    showEditAction={true}
                    showExtensionID="name"
                    showTimestamp={true}
                />
            </div>
        )
    }
}
