import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { discussionsExtensionID } from '.'
import { ErrorNotSupportedPage } from '../components/ErrorNotSupportedPage'
import { PageTitle } from '../components/PageTitle'
import { DiscussionsList } from '../discussions/DiscussionsList'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends RouteComponentProps<any> {
    history: H.History
    location: H.Location
    isExtensionEnabled: (extensionID: string) => boolean
}

interface State {}

/**
 * A page for viewing code discussions on this site.
 */
export class DiscussionsPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('Discussions')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.props.isExtensionEnabled(discussionsExtensionID)) {
            return <ErrorNotSupportedPage />
        }

        return (
            <div className="discussions-page area">
                <div className="area__content">
                    <PageTitle title="Discussions" />
                    <h2>All discussions</h2>
                    <DiscussionsList
                        repoID={undefined}
                        rev={undefined}
                        filePath={'/**'}
                        history={this.props.history}
                        location={this.props.location}
                        noun="discussion"
                        pluralNoun="discussions"
                        defaultFirst={6}
                    />
                </div>
            </div>
        )
    }
}
