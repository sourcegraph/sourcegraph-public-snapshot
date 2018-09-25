import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import { SourcegraphLicense } from './SourcegraphLicense'

interface Props extends RouteComponentProps<{}> {}

/**
 * Displays the Sourcegraph license information from the Sourcegraph license key in site configuration.
 */
export class SiteAdminLicensePage extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLicense')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-license-page">
                <PageTitle title="License" />
                <SourcegraphLicense />
            </div>
        )
    }
}
