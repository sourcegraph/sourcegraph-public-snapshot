import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { SiteAdminLangServers } from './SiteAdminLangServers'

/**
 * A page displaying information about code intelligence on this site.
 */
export class SiteAdminCodeIntelligencePage extends React.PureComponent {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCodeIntelligence')
    }

    public render(): JSX.Element {
        return (
            <div>
                <PageTitle title="Code Intelligence" />
                <SiteAdminLangServers />
            </div>
        )
    }
}
