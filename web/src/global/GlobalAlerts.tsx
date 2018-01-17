import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { NoRepositoriesEnabledAlert } from '../site/NoRepositoriesEnabledAlert'

interface Props {
    isSiteAdmin: boolean
}

interface State {
    siteFlags?: SiteFlags
}

/**
 * Fetches and displays relevant global alerts at the top of the page
 */
export class GlobalAlerts extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(siteFlags.subscribe(siteFlags => this.setState({ siteFlags })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.siteFlags) {
            if (this.state.siteFlags.needsRepositoryConfiguration) {
                return <NeedsRepositoryConfigurationAlert />
            }
            if (this.state.siteFlags.noRepositoriesEnabled) {
                return <NoRepositoriesEnabledAlert />
            }
        }
        return null
    }
}
