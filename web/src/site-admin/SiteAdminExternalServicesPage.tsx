import AddIcon from 'mdi-react/AddIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { map, tap } from 'rxjs/operators'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { createInvalidGraphQLMutationResponseError, dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'

interface ExternalServiceNodeProps {
    node: GQL.IExternalService
    onDidUpdate: () => void
}

interface ExternalServiceNodeState {
    loading: boolean
    errorDescription?: string
}

class ExternalServiceNode extends React.PureComponent<ExternalServiceNodeProps, ExternalServiceNodeState> {
    public state: ExternalServiceNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        return (
            <li
                className="external-service-node list-group-item py-2"
                data-e2e-external-service-name={this.props.node.displayName}
            >
                <div className="d-flex align-items-center justify-content-between">
                    <div>{this.props.node.displayName}</div>
                    <div>
                        <Link
                            className="btn btn-secondary btn-sm"
                            to={`/site-admin/external-services/${this.props.node.id}`}
                            data-tooltip="External service settings"
                        >
                            <SettingsIcon className="icon-inline" /> Settings
                        </Link>{' '}
                        <button
                            className="btn btn-sm btn-danger e2e-delete-external-service-button"
                            onClick={this.deleteExternalService}
                            disabled={this.state.loading}
                            data-tooltip="Delete external service"
                        >
                            <DeleteIcon className="icon-inline" />
                        </button>
                    </div>
                </div>
                {this.state.errorDescription && (
                    <div className="alert alert-danger mt-2">{this.state.errorDescription}</div>
                )}
            </li>
        )
    }

    private deleteExternalService = () => {
        if (!window.confirm(`Delete the external service ${this.props.node.displayName}?`)) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        deleteExternalService(this.props.node.id)
            .toPromise()
            .then(
                () => {
                    // Refresh site flags so that global site alerts
                    // reflect the latest configuration.
                    refreshSiteFlags().subscribe(undefined, err => console.error(err))

                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }
}

function deleteExternalService(externalService: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteExternalService($externalService: ID!) {
                deleteExternalService(externalService: $externalService) {
                    alwaysNil
                }
            }
        `,
        { externalService }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteExternalService) {
                throw createInvalidGraphQLMutationResponseError('DeleteExternalService')
            }
        })
    )
}

interface Props extends RouteComponentProps<{}>, ActivationProps {}

class FilteredExternalServiceConnection extends FilteredConnection<
    GQL.IExternalService,
    Pick<ExternalServiceNodeProps, 'onDidUpdate'>
> {}

/**
 * A page displaying the external services on this site.
 */
export class SiteAdminExternalServicesPage extends React.PureComponent<Props, {}> {
    private updates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalServices')
    }

    private completeConnectedCodeHostActivation = (externalServices: GQL.IExternalServiceConnection) => {
        if (this.props.activation && externalServices.totalCount > 0) {
            this.props.activation.update({ ConnectedCodeHost: true })
        }
    }

    private queryExternalServices = (args: FilteredConnectionQueryArgs): Observable<GQL.IExternalServiceConnection> =>
        queryGraphQL(
            gql`
                query ExternalServices($first: Int) {
                    externalServices(first: $first) {
                        nodes {
                            id
                            kind
                            displayName
                            config
                        }
                        totalCount
                        pageInfo {
                            hasNextPage
                        }
                    }
                }
            `,
            {
                first: args.first,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.externalServices || errors) {
                    throw createAggregateError(errors)
                }
                return data.externalServices
            }),
            tap(this.completeConnectedCodeHostActivation)
        )

    public render(): JSX.Element | null {
        const nodeProps: Pick<ExternalServiceNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateExternalServices,
        }

        return (
            <div className="site-admin-external-services-page">
                <PageTitle title="External services - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                    <h2 className="mb-0">External services</h2>
                    <Link
                        className="btn btn-primary e2e-goto-add-external-service-page"
                        to="/site-admin/external-services/new"
                    >
                        <AddIcon className="icon-inline" /> Add external service
                    </Link>
                </div>
                <p className="mt-2">
                    Manage connections to external services, such as code hosts (to sync repositories).
                </p>
                <FilteredExternalServiceConnection
                    className="list-group list-group-flush mt-3"
                    noun="external service"
                    pluralNoun="external services"
                    queryConnection={this.queryExternalServices}
                    nodeComponent={ExternalServiceNode}
                    nodeComponentProps={nodeProps}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    updates={this.updates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onDidUpdateExternalServices = () => this.updates.next()
}
