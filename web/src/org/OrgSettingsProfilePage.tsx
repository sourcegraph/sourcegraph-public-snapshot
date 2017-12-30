import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { OrgSettingsForm } from './OrgSettingsForm'

interface Props extends RouteComponentProps<any> {
    org: GQL.IOrg
    user: GQL.IUser
}

interface State {
    error?: string
}

/**
 * The organizations settings page
 */
export class OrgSettingsProfilePage extends React.PureComponent<Props, State> {
    public state: State = {}

    private orgChanges = new Subject<GQL.IOrg>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.orgChanges
                .pipe(
                    distinctUntilChanged(),
                    tap(org => eventLogger.logViewEvent('OrgProfile', { organization: { org_name: org.name } }))
                )
                .subscribe()
        )
        this.orgChanges.next(this.props.org)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.org !== this.props.org) {
            this.orgChanges.next(props.org)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="org-settings-profile-page">
                <PageTitle title={this.props.org.name} />
                <div className="org-settings-profile-page__header">
                    <h2>{this.props.org.name}</h2>
                </div>
                <OrgSettingsForm org={this.props.org} />
            </div>
        )
    }
}
