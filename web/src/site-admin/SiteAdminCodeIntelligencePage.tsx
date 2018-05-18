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
                <PageTitle title="Code intelligence" />
                <h2>Code intelligence</h2>
                <p>
                    Sourcegraph uses language servers built on the{' '}
                    <a href="https://langserver.org/">Language Server Protocol</a> (LSP) standard to provide code
                    intelligence. Enable and configure language servers to get hovers, definitions, references,
                    implementations, etc., on code.
                </p>
                <SiteAdminLangServers />
            </div>
        )
    }
}
