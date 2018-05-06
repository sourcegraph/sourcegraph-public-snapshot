import marked from 'marked'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { Markdown } from '../components/Markdown'
import { Settings } from '../schema/settings.schema'
import { currentConfiguration } from '../settings/configuration'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { ConfigurationNoticeAlert } from '../site/ConfigurationNoticeAlert'
import { DockerForMacAlert } from '../site/DockerForMacAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { NoRepositoriesEnabledAlert } from '../site/NoRepositoriesEnabledAlert'
import { UpdateAvailableAlert } from '../site/UpdateAvailableAlert'

interface Props {
    isSiteAdmin: boolean
}

interface State {
    siteFlags?: SiteFlags
    mergedSettings?: Settings
}

/**
 * Fetches and displays relevant global alerts at the top of the page
 */
export class GlobalAlerts extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(siteFlags.subscribe(siteFlags => this.setState({ siteFlags })))
        this.subscriptions.add(currentConfiguration.subscribe(mergedSettings => this.setState({ mergedSettings })))
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
                            <NeedsRepositoryConfigurationAlert className="global-alerts__alert" />
                        ) : (
                            this.state.siteFlags.noRepositoriesEnabled && (
                                <NoRepositoriesEnabledAlert className="global-alerts__alert" />
                            )
                        )}

                        {this.props.isSiteAdmin &&
                            this.state.siteFlags.updateCheck &&
                            !this.state.siteFlags.updateCheck.errorMessage &&
                            this.state.siteFlags.updateCheck.updateVersionAvailable && (
                                <UpdateAvailableAlert
                                    className="global-alerts__alert"
                                    updateVersionAvailable={this.state.siteFlags.updateCheck.updateVersionAvailable}
                                />
                            )}

                        {/* Only show if the user has already enabled repositories; if not yet, the user wouldn't experience any Docker for Mac perf issues anyway. */}
                        {window.context.likelyDockerOnMac &&
                            !this.state.siteFlags.noRepositoriesEnabled && (
                                <DockerForMacAlert className="global-alerts__alert" />
                            )}

                        {this.state.siteFlags.configurationNotice && (
                            <ConfigurationNoticeAlert className="global-alerts__alert" />
                        )}
                    </>
                )}
                {this.state.mergedSettings &&
                    this.state.mergedSettings.motd &&
                    Array.isArray(this.state.mergedSettings.motd) &&
                    this.state.mergedSettings.motd.map(m => (
                        <DismissibleAlert
                            key={m}
                            partialStorageKey={`motd.${m}`}
                            className="alert alert-info global-alerts__alert global-alerts__motd"
                        >
                            <Markdown dangerousInnerHTML={marked(m, { gfm: true, breaks: true, sanitize: true })} />
                        </DismissibleAlert>
                    ))}
            </div>
        )
    }
}
