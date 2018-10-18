import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { combineLatest, concat, from, Observable, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    filter,
    map,
    switchMap,
    take,
    takeUntil,
    withLatestFrom,
} from 'rxjs/operators'
import { ExtensionsProps } from '../../context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../errors'
import { gql, graphQLContent } from '../../graphql'
import * as GQL from '../../schema/graphqlschema'
import { ConfigurationCascadeProps, ConfigurationSubject, Settings } from '../../settings'
import { ConfiguredExtension } from '../extension'
import { ExtensionCard } from './ExtensionCard'

export const registryExtensionFragment = gql`
    fragment RegistryExtensionFields on RegistryExtension {
        id
        publisher {
            __typename
            ... on User {
                id
                username
                displayName
                url
            }
            ... on Org {
                id
                name
                displayName
                url
            }
        }
        extensionID
        extensionIDWithoutRegistry
        name
        manifest {
            raw
            title
            description
        }
        createdAt
        updatedAt
        url
        remoteURL
        registryName
        isLocal
    }
`

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C>,
        RouteComponentProps<{}> {
    subject: Pick<ConfigurationSubject, 'id' | 'viewerCanAdminister'>
    emptyElement?: React.ReactFragment
}

const LOADING: 'loading' = 'loading'

interface ExtensionsResult {
    /** The configured extensions. */
    Extensions: ConfiguredExtension[]

    /** An error message that should be displayed to the user (in addition to the configured extensions). */
    error: string | null
}

interface State {
    /** The current value of the query field. */
    query: string

    /** The data to display. */
    data: {
        /** The query that was used to retrieve the results. */
        query: string

        /** The results, loading, or an error. */
        resultOrError: typeof LOADING | ExtensionsResult | ErrorLike
    }
}

/**
 * Displays a list of all extensions used by a configuration subject.
 */
export class ExtensionsList<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    private static URL_QUERY_PARAM = 'query'

    private updates = new Subject<void>()

    private componentUpdates = new Subject<Props<S, C>>()
    private queryChanges = new Subject<string>()
    private subscriptions = new Subscription()

    constructor(props: Props<S, C>) {
        super(props)
        this.state = {
            query: this.getQueryFromProps(props),
            data: { query: '', resultOrError: LOADING },
        }
    }

    private getQueryFromProps(props: Pick<Props<S, C>, 'location'>): string {
        const params = new URLSearchParams(location.search)
        return params.get(ExtensionsList.URL_QUERY_PARAM) || ''
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.queryChanges.subscribe(query => {
                this.setState({ query })
            })
        )

        const debouncedQueryChanges = this.queryChanges.pipe(debounceTime(50))

        // Update URL when query field changes.
        this.subscriptions.add(
            debouncedQueryChanges.subscribe(query => {
                let search: string
                if (query) {
                    const searchParams = new URLSearchParams()
                    searchParams.set(ExtensionsList.URL_QUERY_PARAM, query)
                    search = searchParams.toString()
                } else {
                    search = ''
                }
                this.props.history.replace({
                    search,
                    hash: this.props.location.hash,
                })
            })
        )

        // Update query field when URL is changed manually.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    filter(({ history }) => history.action !== 'REPLACE'),
                    map(({ location }) => this.getQueryFromProps({ location })),
                    withLatestFrom(debouncedQueryChanges),
                    filter(([urlQuery, debouncedStateQuery]) => urlQuery !== debouncedStateQuery)
                )
                .subscribe(([urlQuery]) => this.setState({ query: urlQuery }))
        )

        this.subscriptions.add(
            combineLatest(debouncedQueryChanges)
                .pipe(
                    switchMap(([query]) => {
                        const resultOrError = this.queryRegistryExtensions({ query }).pipe(
                            catchError(err => [asError(err)])
                        )
                        return concat(
                            of(LOADING).pipe(
                                delay(250),
                                takeUntil(resultOrError)
                            ),
                            resultOrError
                        ).pipe(map(resultOrError => ({ data: { query, resultOrError } })))
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.componentUpdates.next(this.props)
        this.queryChanges.next(this.state.query)
    }

    public componentWillReceiveProps(nextProps: Props<S, C>): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="configured-extensions-list">
                <form onSubmit={this.onSubmit}>
                    <div className="form-group">
                        <input
                            className="form-control"
                            type="search"
                            placeholder="Search extensions..."
                            name="query"
                            value={this.state.query}
                            onChange={this.onQueryChange}
                            autoFocus={true}
                            autoComplete="off"
                            autoCorrect="off"
                            autoCapitalize="off"
                            spellCheck={false}
                        />
                    </div>
                </form>
                {this.state.data.resultOrError === LOADING ? (
                    <this.props.extensions.context.icons.Loader className="icon-inline" />
                ) : isErrorLike(this.state.data.resultOrError) ? (
                    <div className="alert alert-danger">{this.state.data.resultOrError.message}</div>
                ) : (
                    <>
                        {this.state.data.resultOrError.error && (
                            <div className="alert alert-danger my-2">{this.state.data.resultOrError.error}</div>
                        )}
                        {this.state.data.resultOrError.Extensions.length === 0 ? (
                            this.state.data.query ? (
                                <span className="text-muted">
                                    No extensions matching <strong>{this.state.data.query}</strong>
                                </span>
                            ) : (
                                this.props.emptyElement || <span className="text-muted">No extensions found</span>
                            )
                        ) : (
                            <div className="row mt-3">
                                {this.state.data.resultOrError.Extensions.map((e, i) => (
                                    <ExtensionCard
                                        key={i}
                                        subject={this.props.subject}
                                        node={e}
                                        onDidUpdate={this.onDidUpdateExtension}
                                        configurationCascade={this.props.configurationCascade}
                                        extensions={this.props.extensions}
                                    />
                                ))}
                            </div>
                        )}
                    </>
                )}
            </div>
        )
    }

    private onSubmit: React.FormEventHandler = e => e.preventDefault()

    private onQueryChange: React.FormEventHandler<HTMLInputElement> = e => this.queryChanges.next(e.currentTarget.value)

    private queryRegistryExtensions = (args: { query?: string }): Observable<ExtensionsResult> =>
        this.props.extensions.viewerConfiguredExtensions.pipe(
            // Avoid refreshing (and changing order) when the user merely interacts with an extension (e.g.,
            // toggling its enablement), to reduce UI jitter.
            take(1),

            switchMap(viewerExtensions =>
                from(
                    this.props.extensions.context.queryGraphQL(
                        gql`
                            query RegistryExtensions($query: String, $prioritizeExtensionIDs: [String!]!) {
                                extensionRegistry {
                                    extensions(query: $query, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                                        nodes {
                                            ...RegistryExtensionFields
                                        }
                                        error
                                    }
                                }
                            }
                            ${registryExtensionFragment}
                        `[graphQLContent],
                        {
                            ...args,
                            prioritizeExtensionIDs: viewerExtensions.map(({ id }) => id),
                        } as GQL.IExtensionsOnExtensionRegistryArguments,
                        false
                    )
                ).pipe(
                    map(({ data, errors }) => {
                        if (!data || !data.extensionRegistry || !data.extensionRegistry.extensions || errors) {
                            throw createAggregateError(errors)
                        }
                        return {
                            registryExtensions: data.extensionRegistry.extensions.nodes,
                            error: data.extensionRegistry.extensions.error,
                        }
                    })
                )
            ),
            switchMap(({ registryExtensions, error }) =>
                this.props.extensions
                    .withConfiguration(of(registryExtensions))
                    .pipe(map(Extensions => ({ Extensions, error })))
            )
        )

    private onDidUpdateExtension = () => this.updates.next()
}
