import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import * as React from 'react'
import { combineLatest, concat, from, Observable, of, Subject, Subscription, timer } from 'rxjs'
import { catchError, debounce, delay, filter, map, switchMap, take, takeUntil, withLatestFrom } from 'rxjs/operators'
import {
    ConfiguredRegistryExtension,
    isExtensionEnabled,
    toConfiguredRegistryExtension,
} from '../../../shared/src/extensions/extension'
import { viewerConfiguredExtensions } from '../../../shared/src/extensions/helpers'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { Settings, SettingsCascadeProps, SettingsSubject } from '../../../shared/src/settings/settings'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'
import { Form } from '../components/Form'
import { extensionsQuery, isExtensionAdded } from './extension/extension'
import { ExtensionCard } from './ExtensionCard'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'

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

interface Props extends SettingsCascadeProps, PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'> {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    location: H.Location
    history: H.History
}

const LOADING: 'loading' = 'loading'

interface ExtensionsResult {
    /** The configured extensions. */
    extensions: ConfiguredRegistryExtension<GQL.IRegistryExtension>[]

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
 * Displays a list of extensions.
 */
export class ExtensionsList extends React.PureComponent<Props, State> {
    private static URL_QUERY_PARAM = 'query'

    private componentUpdates = new Subject<Props>()
    private queryChanges = new Subject<{ query: string; immediate?: boolean }>()
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
            this.queryChanges.subscribe(({ query }) => {
                this.setState({ query })
            })
        )

        const debouncedQueryChanges = this.queryChanges.pipe(debounce(({ immediate }) => timer(immediate ? 0 : 50)))

        // Update URL when query field changes.
        this.subscriptions.add(
            debouncedQueryChanges.subscribe(({ query }) => {
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
                    filter(([urlQuery, { query: debouncedStateQuery }]) => urlQuery !== debouncedStateQuery)
                )
                .subscribe(([urlQuery]) => this.setState({ query: urlQuery }))
        )

        this.subscriptions.add(
            combineLatest(debouncedQueryChanges)
                .pipe(
                    switchMap(([{ query, immediate }]) => {
                        const resultOrError = this.queryRegistryExtensions({ query }).pipe(
                            catchError(err => [asError(err)])
                        )
                        return concat(
                            of(LOADING).pipe(
                                delay(immediate ? 0 : 250),
                                takeUntil(resultOrError)
                            ),
                            resultOrError
                        ).pipe(map(resultOrError => ({ data: { query, resultOrError } })))
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.componentUpdates.next(this.props)
        this.queryChanges.next({ query: this.state.query, immediate: true })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="extensions-list">
                <Form onSubmit={this.onSubmit} className="form-inline">
                    <input
                        className="form-control flex-1 mr-2 mb-2"
                        type="search"
                        placeholder="Search extensions..."
                        name="query"
                        value={this.state.query}
                        onChange={this.onQueryChangeEvent}
                        autoFocus={true}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                    />
                    <div className="d-flex mb-2">
                        <ExtensionsQueryInputToolbar
                            query={this.state.query}
                            onQueryChange={this.onQueryChangeImmediate}
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
                            <div className="alert alert-danger mb-2">{this.state.data.resultOrError.error}</div>
                        )}
                        {this.state.data.resultOrError.extensions.length === 0 ? (
                            this.state.data.query ? (
                                <div className="text-muted">
                                    No extensions match <strong>{this.state.data.query}</strong>.
                                </div>
                            ) : (
                                <span className="text-muted">No extensions found</span>
                            )
                        ) : (
                            <div className="extensions-list__cards mt-1">
                                {this.state.data.resultOrError.extensions.map((e, i) => (
                                    <ExtensionCard
                                        key={i}
                                        subject={this.props.subject}
                                        node={e}
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

    private onQueryChangeEvent: React.FormEventHandler<HTMLInputElement> = e =>
        this.onQueryChange({ query: e.currentTarget.value })

    private onQueryChangeImmediate = (query: string) => this.queryChanges.next({ query, immediate: true })

    private onQueryChange = ({ query, immediate }: { query: string; immediate?: boolean }) =>
        this.queryChanges.next({ query, immediate })

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
                                    extensions(query: $query, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
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
                        if (!data || !data.extensionRegistry || !data.extensionRegistry.extensions) {
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
                extensions: applyExtensionsQuery(
                    args.query || '',
                    this.props.settingsCascade.final && !isErrorLike(this.props.settingsCascade.final)
                        ? this.props.settingsCascade.final
                        : {},
                    registryExtensions
                ).map(x => toConfiguredRegistryExtension(x)),
                error,
            }))
        )
}

/**
 * Applies the query's client-side extensions search keywords #installed, #enabled, and #disabled by filtering
 * {@link registryExtensions}.
 *
 * @internal Exported for testing only.
 */
export function applyExtensionsQuery<X extends { extensionID: string }>(
    query: string,
    settings: Pick<Settings, 'extensions'>,
    registryExtensions: X[]
): X[] {
    const installed = query.includes(extensionsQuery({ installed: true }))
    const enabled = query.includes(extensionsQuery({ enabled: true }))
    const disabled = query.includes(extensionsQuery({ disabled: true }))
    return registryExtensions.filter(
        x =>
            (!installed || isExtensionAdded(settings, x.extensionID)) &&
            (!enabled || isExtensionEnabled(settings, x.extensionID)) &&
            (!disabled || (isExtensionAdded(settings, x.extensionID) && !isExtensionEnabled(settings, x.extensionID)))
    )
}
