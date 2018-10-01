import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAreaRouteContext } from './UserArea'

interface Props extends UserAreaRouteContext, RouteComponentProps<{}> {}

interface State {}

/**
 * The user overview page.
 */
export class UserOverviewPage extends React.PureComponent<Props, State> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('UserOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-page user-overview-page">
                <PageTitle title={this.props.user.username} />
                <p>
                    {this.props.user.displayName ? (
                        <>
                            {this.props.user.displayName} ({this.props.user.username})
                        </>
                    ) : (
                        this.props.user.username
                    )}{' '}
                    started using Sourcegraph <Timestamp date={this.props.user.createdAt} />.
                </p>
            </div>
        )
    }
}
