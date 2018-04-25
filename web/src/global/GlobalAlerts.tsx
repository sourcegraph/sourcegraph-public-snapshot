import * as React from 'react'
import { Subscription } from 'rxjs'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { DockerForMacAlert } from '../site/DockerForMacAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { NoRepositoriesEnabledAlert } from '../site/NoRepositoriesEnabledAlert'
import { UpdateAvailableAlert } from '../site/UpdateAvailableAlert'

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
            if (
                this.props.isSiteAdmin &&
                this.state.siteFlags.updateCheck &&
                !this.state.siteFlags.updateCheck.errorMessage &&
                this.state.siteFlags.updateCheck.updateVersionAvailable
            ) {
                return (
                    <UpdateAvailableAlert
                        updateVersionAvailable={this.state.siteFlags.updateCheck.updateVersionAvailable}
                    />
                )
            }
        }

        // Last priority, we would show NeedsRepositoryConfigurationAlert or
        // NoRepositoriesEnabledAlert first above before the user would see
        // this. That's OK because they would need to do both of those before
        // they could see performance issues caused by Docker for Mac.
        if (window.context.likelyDockerOnMac) {
            return <DockerForMacAlert />
        }
        return null
    }
}
