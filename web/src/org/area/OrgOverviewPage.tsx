import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {}

interface State {}

/**
 * The organization overview page.
 */
export class OrgOverviewPage extends React.PureComponent<Props, State> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('OrgOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="org-page org-overview-page">
                <PageTitle title={that.props.org.name} />
                <p>
                    {that.props.org.displayName ? (
                        <>
                            {that.props.org.displayName} ({that.props.org.name})
                        </>
                    ) : (
                        that.props.org.name
                    )}{' '}
                    was created <Timestamp date={that.props.org.createdAt} />.
                </p>
            </div>
        )
    }
}
