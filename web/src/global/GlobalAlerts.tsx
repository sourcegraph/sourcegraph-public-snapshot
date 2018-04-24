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
        return (
            <div className="global-alerts">
                {this.state.siteFlags && (
                    <>
                        {this.state.siteFlags.needsRepositoryConfiguration ? (
                            <NeedsRepositoryConfigurationAlert />
                        ) : (
                            this.state.siteFlags.noRepositoriesEnabled && <NoRepositoriesEnabledAlert />
                        )}

                        {this.props.isSiteAdmin &&
                            this.state.siteFlags.updateCheck &&
                            !this.state.siteFlags.updateCheck.errorMessage &&
                            this.state.siteFlags.updateCheck.updateVersionAvailable && (
                                <UpdateAvailableAlert
                                    updateVersionAvailable={this.state.siteFlags.updateCheck.updateVersionAvailable}
                                />
                            )}

                        {/* Only show if the user has already enabled repositories; if not yet, the user wouldn't experience any Docker for Mac perf issues anyway. */}
                        {window.context.likelyDockerOnMac &&
                            !this.state.siteFlags.noRepositoriesEnabled && <DockerForMacAlert />}
                    </>
                )}
            </div>
        )
    }
}
