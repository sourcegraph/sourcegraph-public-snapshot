import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
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
import { ConfiguredExtension, toConfiguredExtensions } from '../../../shared/src/extensions/extension'
import { viewerConfiguredExtensions } from '../../../shared/src/extensions/helpers'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps, SettingsSubject } from '../../../shared/src/settings/settings'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'
import { Form } from '../components/Form'
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
        isWorkInProgress
        viewerCanAdminister
    }
`

interface Props extends SettingsCascadeProps, PlatformContextProps, RouteComponentProps<{}> {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    emptyElement?: React.ReactFragment
}

const LOADING: 'loading' = 'loading'

interface ExtensionsResult {
    /** The configured extensions. */
    extensions: ConfiguredExtension<GQL.IRegistryExtension>[]

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
 * Displays a list of all extensions used by a settings subject.
 */
export class ExtensionsList extends React.PureComponent<Props, State> {
    private static URL_QUERY_PARAM = 'query'

    private updates = new Subject<void>()

    private componentUpdates = new Subject<Props>()
    private queryChanges = new Subject<string>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            query: this.getQueryFromProps(props),
            data: { query: '', resultOrError: LOADING },
        }
    }

    private getQueryFromProps(props: Pick<Props, 'location'>): string {
        const params = new URLSearchParams(props.location.search)
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

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="configured-extensions-list">
                <Form onSubmit={this.onSubmit}>
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
                </Form>
                {this.state.data.resultOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.data.resultOrError) ? (
                    <div className="alert alert-danger">{this.state.data.resultOrError.message}</div>
                ) : (
                    <>
                        {this.state.data.resultOrError.error && (
                            <div className="alert alert-danger my-2">{this.state.data.resultOrError.error}</div>
                        )}
                        {this.state.data.resultOrError.extensions.length === 0 ? (
                            this.state.data.query ? (
                                <span className="text-muted">
                                    No extensions matching <strong>{this.state.data.query}</strong>
                                </span>
                            ) : (
                                this.props.emptyElement || <span className="text-muted">No extensions found</span>
                            )
                        ) : (
                            <div className="row mt-3">
                                {this.state.data.resultOrError.extensions.map((e, i) => (
                                    <ExtensionCard
                                        key={i}
                                        subject={this.props.subject}
                                        node={e}
                                        onDidUpdate={this.onDidUpdateExtension}
                                        settingsCascade={this.props.settingsCascade}
                                        platformContext={this.props.platformContext}
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
        viewerConfiguredExtensions(this.props.platformContext).pipe(
            // Avoid refreshing (and changing order) when the user merely interacts with an extension (e.g.,
            // toggling its enablement), to reduce UI jitter.
            take(1),

            switchMap(viewerExtensions =>
                from(
                    queryGraphQL(
                        gql`
                            query RegistryExtensions($query: String, $prioritizeExtensionIDs: [String!]!) {
                                extensionRegistry {
                                    extensions(
                                        query: $query
                                        prioritizeExtensionIDs: $prioritizeExtensionIDs
                                        includeWIP: true
                                    ) {
                                        nodes {
                                            ...RegistryExtensionFields
                                        }
                                        error
                                    }
                                }
                            }
                            ${registryExtensionFragment}
                        `,
                        {
                            ...args,
                            prioritizeExtensionIDs: viewerExtensions.map(({ id }) => id),
                        } as GQL.IExtensionsOnExtensionRegistryArguments
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
            map(({ registryExtensions, error }) => ({
                extensions: toConfiguredExtensions(registryExtensions),
                error,
            }))
        )

    private onDidUpdateExtension = () => this.updates.next()
}
