import { ExtensionsList } from '@sourcegraph/extensions-client-common/lib/extensions/manager/ExtensionsList'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExtensionsAreaPageProps } from './ExtensionsArea'
import { ExtensionsEmptyState } from './ExtensionsEmptyState'

interface Props extends ExtensionsAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays overview information about viewer's configured extensions. */
export class ExtensionsOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('ExtensionsOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="extensions-overview-page container px-2 px-xl-0">
                <PageTitle title="Extensions" />
                <div className="py-3">
                    <ExtensionsList
                        {...this.props}
                        emptyElement={<ExtensionsEmptyState />}
                        subject={this.props.authenticatedUser.id}
                    />
                </div>
            </div>
        )
    }
}
