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
            <div className="extensions-explore-section">
                <h2 className="extensions-explore-section__section-title">Top Sourcegraph extensions</h2>
                {isErrorLike(extensionsOrError) ? (
                    <div className="alert alert-danger">Error: {extensionsOrError.message}</div>
                ) : extensionsOrError.length === 0 ? (
                    <p>No extensions are available.</p>
                ) : (
                    <>
                        <div className="extensions-explore-section__row">
                            {extensionsOrError.slice(0, ExtensionsExploreSection.QUERY_EXTENSIONS_ARG_FIRST).map((
                                extension /* or loading */,
                                i
                            ) => (
                                <div key={i} className="extensions-explore-section__card">
                                    {extension === LOADING ? (
                                        <ExtensionsExploreSectionExtensionCard
                                            extensionID=""
                                            // Spacer to reduce loading jitter.
                                            description=""
                                        />
                                    ) : (
                                        <ExtensionsExploreSectionExtensionCard
                                            extensionID={extension.extensionIDWithoutRegistry}
                                            description={
                                                (extension.manifest && extension.manifest.description) || undefined
                                            }
                                            url={extension.url}
                                        />
                                    )}
                                </div>
                            ))}
                        </div>
                        <div className="text-right mt-2">
                            <Link to="/extensions">
                                View all extensions
                                <ChevronRightIcon className="icon-inline" />
                            </Link>
                        </div>
                    </>
                )}
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
