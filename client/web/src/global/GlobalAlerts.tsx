import * as React from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'
import differenceInDays from 'date-fns/differenceInDays'
import { Subscription } from 'rxjs'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { CodeHostScopeAlerts, GitLabScopeAlert } from '../site/CodeHostScopeAlerts/CodeHostScopeAlerts'
import { DockerForMacAlert } from '../site/DockerForMacAlert'
import { FreeUsersExceededAlert } from '../site/FreeUsersExceededAlert'
import { LicenseExpirationAlert } from '../site/LicenseExpirationAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'

import { GlobalAlert } from './GlobalAlert'
import { Notices } from './Notices'

import styles from './GlobalAlerts.module.scss'

interface Props extends SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
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
            <div className={classNames('test-global-alert', styles.globalAlerts)}>
                {this.state.siteFlags && (
                    <>
                        {this.state.siteFlags.needsRepositoryConfiguration && (
                            <NeedsRepositoryConfigurationAlert className={styles.alert} />
                        )}
                        {this.state.siteFlags.freeUsersExceeded && (
                            <FreeUsersExceededAlert
                                noLicenseWarningUserCount={
                                    this.state.siteFlags.productSubscription.noLicenseWarningUserCount
                                }
                                className={styles.alert}
                            />
                        )}
                        {/* Only show if the user has already added repositories; if not yet, the user wouldn't experience any Docker for Mac perf issues anyway. */}
                        {window.context.likelyDockerOnMac && window.context.deployType === 'docker-container' && (
                            <DockerForMacAlert className={styles.alert} />
                        )}
                        {window.context.sourcegraphDotComMode && (
                            <CodeHostScopeAlerts authenticatedUser={this.props.authenticatedUser} />
                        )}
                        {window.context.sourcegraphDotComMode && (
                            <GitLabScopeAlert authenticatedUser={this.props.authenticatedUser} />
                        )}
                        {this.state.siteFlags.alerts.map((alert, index) => (
                            <GlobalAlert key={index} alert={alert} className={styles.alert} />
                        ))}
                        {this.state.siteFlags.productSubscription.license &&
                            (() => {
                                const expiresAt = parseISO(this.state.siteFlags.productSubscription.license.expiresAt)
                                return (
                                    differenceInDays(expiresAt, Date.now()) <= 7 && (
                                        <LicenseExpirationAlert
                                            expiresAt={expiresAt}
                                            daysLeft={Math.floor(differenceInDays(expiresAt, Date.now()))}
                                            className={styles.alert}
                                        />
                                    )
                                )
                            })()}
                    </>
                )}
                {isSettingsValid<Settings>(this.props.settingsCascade) &&
                    this.props.settingsCascade.final.motd &&
                    Array.isArray(this.props.settingsCascade.final.motd) &&
                    this.props.settingsCascade.final.motd.map(motd => (
                        <DismissibleAlert
                            key={motd}
                            partialStorageKey={`motd.${motd}`}
                            variant="info"
                            className={styles.alert}
                        >
                            <Markdown dangerousInnerHTML={renderMarkdown(motd)} />
                        </DismissibleAlert>
                    ))}
                {process.env.SOURCEGRAPH_API_URL && (
                    <DismissibleAlert
                        key="dev-web-server-alert"
                        partialStorageKey="dev-web-server-alert"
                        variant="danger"
                        className={styles.alert}
                    >
                        <div>
                            <strong>Warning!</strong> This build uses data from the proxied API:{' '}
                            <Link target="__blank" to={process.env.SOURCEGRAPH_API_URL}>
                                {process.env.SOURCEGRAPH_API_URL}
                            </Link>
                        </div>
                        .
                    </DismissibleAlert>
                )}
                <Notices alertClassName={styles.alert} location="top" settingsCascade={this.props.settingsCascade} />
            </div>
        )
    }
}
