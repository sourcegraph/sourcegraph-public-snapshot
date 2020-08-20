import AddIcon from 'mdi-react/AddIcon'
import React from 'react'
import { RouteComponentProps, Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { tap } from 'rxjs/operators'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../FilteredConnection'
import { PageTitle } from '../PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import * as H from 'history'
import { queryExternalServices } from './backend'
import { ExternalServiceNodeProps, ExternalServiceNode } from './ExternalServiceNode'
import { ListExternalServiceFields, ExternalServicesResult, Scalars } from '../../graphql-operations'

interface Props extends RouteComponentProps<{}>, ActivationProps {
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
    userID?: Scalars['ID']
}

interface State {
    noExternalServices?: boolean
}

/**
 * A page displaying the external services on this site.
 */
export class ExternalServicesPage extends React.PureComponent<Props, State> {
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {}
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalServices')
        this.subscriptions.add(
            this.queryExternalServices({ first: 1 }).subscribe(externalServicesResult =>
                this.setState({ noExternalServices: externalServicesResult.totalCount === 0 })
            )
        )
    }

    private completeConnectedCodeHostActivation = (
        externalServices: ExternalServicesResult['externalServices']
    ): void => {
        if (this.props.activation && externalServices.totalCount > 0) {
            this.props.activation.update({ ConnectedCodeHost: true })
        }
    }

    private queryExternalServices = (
        args: FilteredConnectionQueryArgs
    ): Observable<ExternalServicesResult['externalServices']> =>
        queryExternalServices({ first: args.first ?? null, namespace: this.props.userID ?? null }).pipe(
            tap(externalServices => this.completeConnectedCodeHostActivation(externalServices))
        )

    public render(): JSX.Element | null {
        const nodeProps: Omit<ExternalServiceNodeProps, 'node'> = {
            onDidUpdate: this.onDidUpdateExternalServices,
            history: this.props.history,
            routingPrefix: this.props.routingPrefix,
            afterDeleteRoute: this.props.afterDeleteRoute,
        }

        if (this.state.noExternalServices) {
            return <Redirect to={`${this.props.routingPrefix}/external-services/new`} />
        }
        return (
            <div className="site-admin-external-services-page">
                <PageTitle title="Manage repositories - Admin" />
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <h2 className="mb-0">Manage repositories</h2>
                    <Link
                        className="btn btn-primary test-goto-add-external-service-page"
                        to={`${this.props.routingPrefix}/external-services/new`}
                    >
                        <AddIcon className="icon-inline" /> Add repositories
                    </Link>
                </div>
                <p className="mt-2">Manage code host connections to sync repositories.</p>
                <FilteredConnection<ListExternalServiceFields, Omit<ExternalServiceNodeProps, 'node'>>
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

    private onDidUpdateExternalServices = (): void => this.updates.next()
}
