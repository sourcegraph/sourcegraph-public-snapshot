import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props {}

interface State {}

/**
 * A page displaying information about telemetry pings for the site.
 */
export class SiteAdminPingsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminPings')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nonCriticalTelemetryDisabled = window.context.site.disableNonCriticalTelemetry === true
        const updatesDisabled = window.context.site['update.channel'] !== 'release'

        return (
            <div className="site-admin-pings-page">
                <PageTitle title="Pings - Admin" />
                <h2>Pings</h2>
                <p>
                    Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It
                    sends only the high-level data below. It never sends code, repository names, usernames, or any other
                    specific data.
                </p>
                <h3>Critical telemetry</h3>
                <p>
                    Critical telemetry includes only the high-level data below required for billing, support, updates,
                    and security notices. This cannot be disabled.
                </p>
                <ul>
                    <li>Randomly generated site identifier</li>
                    <li>
                        The email address of the initial site installer (or if deleted, the first active site admin), to
                        know who to contact regarding sales, product updates, security updates, and policy updates
                    </li>
                    <li>Sourcegraph version string (e.g. "vX.X.X")</li>
                    <li>
                        Deployment type (single Docker image, Docker Compose, Kubernetes cluster, or pure Docker
                        cluster)
                    </li>
                    <li>License key associated with your Sourcegraph subscription</li>
                    <li>Aggregate count of current monthly users</li>
                    <li>Total count of existing user accounts</li>
                </ul>
                <h3>Other telemetry</h3>
                <p>
                    By default, Sourcegraph also aggregates usage and performance metrics for some product features. No
                    personal or specific information is ever included. Starting in May 2020 (Sourcegraph version 3.16),
                    Sourcegraph admins can disable the telemetry items below by setting the{' '}
                    <code>DisableNonCriticalTelemetry</code> setting to <code>true</code> on the{' '}
                    <Link to="/site-admin/configuration">Site configuration page</Link>.
                </p>
                <ul>
                    <li>Whether the instance is deployed on localhost (true/false)</li>
                    <li>
                        Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy,
                        SAML, GitHub, GitLab)
                    </li>
                    <li>
                        Which code hosts are in use (GitHub, Bitbucket Server, GitLab, Phabricator, Gitolite, AWS
                        CodeCommit, Other)
                    </li>
                    <li>Whether new user signup is allowed (true/false)</li>
                    <li>Whether a repository has ever been added (true/false)</li>
                    <li>Whether a code search has ever been executed (true/false)</li>
                    <li>Whether code intelligence has ever been used (true/false)</li>
                    <li>Aggregate counts of current daily, weekly, and monthly users</li>
                    <li>
                        Aggregate counts of current daily, weekly, and monthly users, by:
                        <ul>
                            <li>Whether they are using code host integrations</li>
                            <li>
                                Product area (site management, code search and navigation, code review, saved searches,
                                diff searches)
                            </li>
                            <li>Search modes used (interactive search, plain-text search)</li>
                            <li>Search filters used (e.g. "type:", "repo:", "file:", "lang:", etc.)</li>
                        </ul>
                    </li>
                    <li>
                        Aggregate daily, weekly, and monthly latencies (in ms) of search queries
                    </li>
                    <li>
                        Aggregate daily, weekly, and monthly counts of:
                        <ul>
                            <li>Code intelligence events (e.g., hover tooltips)</li>
                            <li>Searches using each search mode (interactive search, plain-text search)</li>
                            <li>Searches using each search filter (e.g. "type:", "repo:", "file:", "lang:", etc.)</li>
                        </ul>
                    </li>
                    <li>
                        Campaign usage data
                        <ul>
                            <li>Total count of created campaigns</li>
                            <li>Total count of changesets created by campaigns</li>
                            <li>Total count of changesets created by campaigns that have been merged</li>
                            <li>Total count of changesets manually added to a campaign</li>
                            <li>Total count of changesets manually added to a campaign that have been merged</li>
                        </ul>
                    </li>
                    <li>
                        Monthly aggregated user state changes
                        <ul>
                            <li>Count of users created</li>
                            <li>Count of users deleted</li>
                            <li>Count of users retained</li>
                            <li>Count of users resurrected</li>
                            <li>Count of users churned</li>
                        </ul>
                    </li>
                    <li>
                        Saved searches usage data
                        <ul>
                            <li>Count of saved searches</li>
                            <li>Count of users using saved searches</li>
                            <li>Count of notifications triggered</li>
                            <li>Count of notifications clicked</li>
                            <li>Count of saved search views</li>
                        </ul>
                    </li>
                    <li>
                        Aggregated repository statistics
                        <ul>
                            <li>Total size of git repositories stored in bytes</li>
                            <li>Total number of lines of code stored in text search index</li>
                        </ul>
                    </li>
                </ul>
                {updatesDisabled ? (
                    <p>All telemetry is disabled.</p>
                ) : (
                    nonCriticalTelemetryDisabled && <p>Non-critical telemetry is disabled.</p>
                )}
            </div>
        )
    }
}
