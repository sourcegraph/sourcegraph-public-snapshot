import ViewIcon from '@sourcegraph/icons/lib/View'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { RegistryAreaPageProps } from './RegistryArea'
import { ExtensionsListViewMode, RegistryExtensionsList } from './RegistryExtensionsPage'
import { RegistryPublishersList } from './RegistryPublishersList'

interface ViewerRegistryInfo {
    configuredExtensionsURL: string | null
    registryExtensionsURL: string | null
}

function queryViewerRegistryInfo(): Observable<ViewerRegistryInfo | null> {
    return queryGraphQL(gql`
        query ViewerRegistryInfo {
            currentUser {
                configuredExtensions {
                    url
                }
                registryExtensions {
                    url
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            if (!data.currentUser) {
                return null
            }
            return {
                configuredExtensionsURL: data.currentUser.configuredExtensions.url,
                registryExtensionsURL: data.currentUser.registryExtensions.url,
            }
        })
    )
}

interface Props extends RegistryAreaPageProps, RouteComponentProps<{}> {}

interface State {
    /** The viewer's registry-related info, undefined while loading, null if unauthenticated, or an error. */
    viewerInfoOrError: 'loading' | ViewerRegistryInfo | ErrorLike

    extensionsListViewMode: ExtensionsListViewMode
}

/** A page that displays overview information about the registry. */
export class RegistryOverviewPage extends React.PureComponent<Props, State> {
    private static STORAGE_KEY = 'RegistryOverviewPage.extensionsListViewMode'
    private static getExtensionListViewMode(): ExtensionsListViewMode {
        const v = localStorage.getItem(RegistryOverviewPage.STORAGE_KEY)
        if (v === ExtensionsListViewMode.Cards || v === ExtensionsListViewMode.List) {
            return v
        }
        return ExtensionsListViewMode.Cards
    }
    private static setExtensionListViewMode(value: ExtensionsListViewMode): void {
        localStorage.setItem(RegistryOverviewPage.STORAGE_KEY, value)
    }

    public state: State = {
        viewerInfoOrError: 'loading',
        extensionsListViewMode: RegistryOverviewPage.getExtensionListViewMode(),
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryOverview')

        this.subscriptions.add(
            concat(
                [{ viewerInfoOrError: 'loading' }],
                queryViewerRegistryInfo().pipe(
                    map(result => ({ viewerInfoOrError: result })),
                    catchError(error => [{ viewerInfoOrError: asError(error) }])
                )
            ).subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-overview-page">
                <PageTitle title="Registry" />
                <DismissibleAlert className="alert-info" partialStorageKey="registry-overview-help0">
                    <span>
                        <strong>Experimental feature:</strong> Extensions add features to Sourcegraph and other
                        connected tools (such as editors, code hosts, and code review tools).
                    </span>
                </DismissibleAlert>
                <div className="row">
                    <div className="col-sm-9 mt-3">
                        <RegistryExtensionsList
                            {...this.props}
                            mode={this.state.extensionsListViewMode}
                            publisher={null}
                            showExtensionID="extensionIDWithoutRegistry"
                            showSource={true}
                            showUserActions={true}
                            filters={RegistryExtensionsList.FILTERS}
                        />
                    </div>
                    <div className="col-sm-3 mt-3">
                        {this.state.viewerInfoOrError &&
                            this.state.viewerInfoOrError !== 'loading' &&
                            !isErrorLike(this.state.viewerInfoOrError) && (
                                <>
                                    {this.state.viewerInfoOrError.configuredExtensionsURL && (
                                        <Link
                                            className="w-100 mb-2 d-block"
                                            to={this.state.viewerInfoOrError.configuredExtensionsURL}
                                        >
                                            Your in-use extensions
                                        </Link>
                                    )}
                                    {this.state.viewerInfoOrError.registryExtensionsURL && (
                                        <Link
                                            className="w-100 mb-2 d-block"
                                            to={this.state.viewerInfoOrError.registryExtensionsURL}
                                        >
                                            Your published extensions
                                        </Link>
                                    )}
                                </>
                            )}
                        <div className="card mt-3">
                            <div className="card-header">
                                <h4 className="mb-0">Publishers</h4>
                            </div>
                            <RegistryPublishersList {...this.props} />
                        </div>
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm d-flex align-items-center mt-3"
                            onClick={this.onExtensionsListViewModeButtonClick}
                        >
                            <ViewIcon className="icon-inline mr-1" /> Use{' '}
                            {this.state.extensionsListViewMode === ExtensionsListViewMode.Cards
                                ? ExtensionsListViewMode.List
                                : ExtensionsListViewMode.Cards}{' '}
                            view
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    private onExtensionsListViewModeButtonClick = () => {
        this.setState(
            prevState => ({
                extensionsListViewMode:
                    prevState.extensionsListViewMode === ExtensionsListViewMode.Cards
                        ? ExtensionsListViewMode.List
                        : ExtensionsListViewMode.Cards,
            }),
            () => RegistryOverviewPage.setExtensionListViewMode(this.state.extensionsListViewMode)
        )
    }
}
