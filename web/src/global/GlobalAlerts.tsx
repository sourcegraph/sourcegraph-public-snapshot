import differenceInDays from 'date-fns/differenceInDays'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { Markdown } from '../../../shared/src/components/Markdown'
import { isSettingsValid, SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { Settings } from '../schema/settings.schema'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { DockerForMacAlert } from '../site/DockerForMacAlert'
import { LicenseExpirationAlert } from '../site/LicenseExpirationAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { NoRepositoriesEnabledAlert } from '../site/NoRepositoriesEnabledAlert'
import { UpdateAvailableAlert } from '../site/UpdateAvailableAlert'
import { GlobalAlert } from './GlobalAlert'
import { Notices } from './Notices'

interface Props extends SettingsCascadeProps {
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
                        {window.context.likelyDockerOnMac && !this.state.siteFlags.noRepositoriesEnabled && (
                            <DockerForMacAlert className="global-alerts__alert" />
                        )}

                        {this.state.siteFlags.alerts.map((alert, i) => (
                            <GlobalAlert key={i} alert={alert} className="global-alerts__alert" />
                        ))}

                        {this.state.siteFlags.productSubscription.license &&
                            differenceInDays(this.state.siteFlags.productSubscription.license.expiresAt, Date.now()) <=
                                365 && (
                                <LicenseExpirationAlert
                                    expiresAt={this.state.siteFlags.productSubscription.license.expiresAt}
                                    daysLeft={Math.floor(
                                        differenceInDays(
                                            this.state.siteFlags.productSubscription.license.expiresAt,
                                            Date.now()
                                        )
                                    )}
                                />
                            )}
                    </>
                )}
                {isSettingsValid<Settings>(this.props.settingsCascade) &&
                    this.props.settingsCascade.final.motd &&
                    Array.isArray(this.props.settingsCascade.final.motd) &&
                    this.props.settingsCascade.final.motd.map(m => (
                        <DismissibleAlert
                            key={m}
                            partialStorageKey={`motd.${m}`}
                            className="alert alert-info global-alerts__alert"
                        >
                            <Markdown dangerousInnerHTML={renderMarkdown(m)} />
                        </DismissibleAlert>
                    ))}
                <Notices
                    alertClassName="global-alerts__alert"
                    location="top"
                    settingsCascade={this.props.settingsCascade}
                />
            </div>
        )
    }
}
