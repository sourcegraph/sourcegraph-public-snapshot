import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { ExtensionsExploreSectionExtensionCard } from './ExtensionsExploreSectionExtensionCard'
import { ErrorAlert } from '../../../components/alerts'

interface Props {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The extensions, loading, or an error. */
    extensionsOrError: typeof LOADING | GQL.IRegistryExtensionConnection | ErrorLike
}

/**
 * An explore section that shows extensions.
 */
export class ExtensionsExploreSection extends React.PureComponent<Props, State> {
    private static QUERY_EXTENSIONS_ARG_FIRST = 4

    /**
     * Extension IDs to prioritize displaying. If the length of this array is >= than QUERY_EXTENSIONS_ARG_FIRST,
     * then these will be the only ones shown (which is intended).
     */
    private static QUERY_EXTENSIONS_ARG_EXTENSION_IDS = [
        'sourcegraph/codecov',
        'sourcegraph/datadog-metrics',
        'sourcegraph/git-extras',
    ]

    public state: State = { extensionsOrError: LOADING }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            queryExtensions({
                first: ExtensionsExploreSection.QUERY_EXTENSIONS_ARG_FIRST,
                prioritizeExtensionIDs: ExtensionsExploreSection.QUERY_EXTENSIONS_ARG_EXTENSION_IDS,
            })
                .pipe(catchError(err => [asError(err)]))
                .subscribe(extensionsOrError => this.setState({ extensionsOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const extensionsOrError: (typeof LOADING | GQL.IRegistryExtension)[] | ErrorLike =
            this.state.extensionsOrError === LOADING
                ? Array(ExtensionsExploreSection.QUERY_EXTENSIONS_ARG_FIRST).fill(LOADING)
                : isErrorLike(this.state.extensionsOrError)
                ? this.state.extensionsOrError
                : this.state.extensionsOrError.nodes

        return (
            <div className="card">
                <h3 className="card-header">Top Sourcegraph extensions</h3>
                {isErrorLike(extensionsOrError) ? (
                    <ErrorAlert error={extensionsOrError} />
                ) : extensionsOrError.length === 0 ? (
                    <p>No extensions are available.</p>
                ) : (
                    <div className="list-group list-group-flush">
                        {extensionsOrError
                            .slice(0, ExtensionsExploreSection.QUERY_EXTENSIONS_ARG_FIRST)
                            .filter((e): e is GQL.IRegistryExtension => e !== LOADING)
                            .map((extension, i) => (
                                <ExtensionsExploreSectionExtensionCard
                                    key={extension.id}
                                    extensionID={extension.extensionIDWithoutRegistry}
                                    description={(extension.manifest && extension.manifest.description) || undefined}
                                    url={extension.url}
                                    className="list-group-item list-group-item-action"
                                />
                            ))}
                    </div>
                )}
                <div className="card-footer">
                    <Link to="/extensions">
                        View all extensions
                        <ChevronRightIcon className="icon-inline" />
                    </Link>
                </div>
            </div>
        )
    }
}

function queryExtensions(
    args: Pick<GQL.IExtensionsOnExtensionRegistryArguments, 'first' | 'prioritizeExtensionIDs'>
): Observable<GQL.IRegistryExtensionConnection> {
    return queryGraphQL(
        gql`
            query ExploreExtensions($first: Int, $prioritizeExtensionIDs: [String!]) {
                extensionRegistry {
                    extensions(first: $first, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                        nodes {
                            id
                            extensionIDWithoutRegistry
                            url
                            manifest {
                                description
                            }
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.extensionRegistry ||
                !data.extensionRegistry.extensions ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extensions
        })
    )
}
