import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import * as React from 'react'
import { concat, from, Observable, of, Subject, Subscription, timer } from 'rxjs'
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
import { ErrorAlert } from '../components/alerts'

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
        that.state = {
            query: that.getQueryFromProps(props),
            data: { query: '', resultOrError: LOADING },
        }
    }

    private getQueryFromProps(props: Pick<Props, 'location'>): string {
        const params = new URLSearchParams(props.location.search)
        return params.get(ExtensionsList.URL_QUERY_PARAM) || ''
    }

    public componentDidMount(): void {
        that.subscriptions.add(
            that.queryChanges.subscribe(({ query }) => {
                that.setState({ query })
            })
        )

        const debouncedQueryChanges = that.queryChanges.pipe(debounce(({ immediate }) => timer(immediate ? 0 : 50)))

        // Update URL when query field changes.
        that.subscriptions.add(
            debouncedQueryChanges.subscribe(({ query }) => {
                let search: string
                if (query) {
                    const searchParams = new URLSearchParams()
                    searchParams.set(ExtensionsList.URL_QUERY_PARAM, query)
                    search = searchParams.toString()
                } else {
                    search = ''
                }
                that.props.history.replace({
                    search,
                    hash: that.props.location.hash,
                })
            })
        )

        // Update query field when URL is changed manually.
        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    filter(({ history }) => history.action !== 'REPLACE'),
                    map(({ location }) => that.getQueryFromProps({ location })),
                    withLatestFrom(debouncedQueryChanges),
                    filter(([urlQuery, { query: debouncedStateQuery }]) => urlQuery !== debouncedStateQuery)
                )
                .subscribe(([urlQuery]) => that.setState({ query: urlQuery }))
        )

        that.subscriptions.add(
            debouncedQueryChanges
                .pipe(
                    switchMap(({ query, immediate }) => {
                        const resultOrError = that.queryRegistryExtensions({ query }).pipe(
                            catchError(err => [asError(err)])
                        )
                        return concat(
                            of(LOADING).pipe(delay(immediate ? 0 : 250), takeUntil(resultOrError)),
                            resultOrError
                        ).pipe(map(resultOrError => ({ data: { query, resultOrError } })))
                    })
                )
                .subscribe(stateUpdate => that.setState(stateUpdate))
        )

        that.componentUpdates.next(that.props)
        that.queryChanges.next({ query: that.state.query, immediate: true })
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="extensions-list">
                <Form onSubmit={that.onSubmit} className="form-inline">
                    <input
                        className="form-control flex-grow-1 mr-1 mb-2"
                        type="search"
                        placeholder="Search extensions..."
                        name="query"
                        value={that.state.query}
                        onChange={that.onQueryChangeEvent}
                        autoFocus={true}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                    />
                    <div className="mb-2">
                        <ExtensionsQueryInputToolbar
                            query={that.state.query}
                            onQueryChange={that.onQueryChangeImmediate}
                        />
                    </div>
                </Form>
                {that.state.data.resultOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(that.state.data.resultOrError) ? (
                    <ErrorAlert error={that.state.data.resultOrError} />
                ) : (
                    <>
                        {that.state.data.resultOrError.error && (
                            <ErrorAlert className="mb-2" error={that.state.data.resultOrError.error} />
                        )}
                        {that.state.data.resultOrError.extensions.length === 0 ? (
                            that.state.data.query ? (
                                <div className="text-muted">
                                    No extensions match <strong>{that.state.data.query}</strong>.
                                </div>
                            ) : (
                                <span className="text-muted">No extensions found</span>
                            )
                        ) : (
                            <div className="extensions-list__cards mt-1">
                                {that.state.data.resultOrError.extensions.map((e, i) => (
                                    <ExtensionCard
                                        key={i}
                                        subject={that.props.subject}
                                        node={e}
                                        settingsCascade={that.props.settingsCascade}
                                        platformContext={that.props.platformContext}
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
        that.onQueryChange({ query: e.currentTarget.value })

    private onQueryChangeImmediate = (query: string): void => that.queryChanges.next({ query, immediate: true })

    private onQueryChange = ({ query, immediate }: { query: string; immediate?: boolean }): void =>
        that.queryChanges.next({ query, immediate })

    private queryRegistryExtensions = (args: { query?: string }): Observable<ExtensionsResult> =>
        viewerConfiguredExtensions(that.props.platformContext).pipe(
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
                    that.props.settingsCascade.final && !isErrorLike(that.props.settingsCascade.final)
                        ? that.props.settingsCascade.final
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
