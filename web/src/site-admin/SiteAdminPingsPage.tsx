import * as React from 'react'
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
        const pingsEnabled = window.context.site['update.channel'] === 'release'

        return (
            <div className="site-admin-pings-page">
                <PageTitle title="Pings - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">Pings</h2>
                </div>
                <p>
                    Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It
                    sends only the high-level data below. It never sends code, repository names, usernames, or any other
                    specific data.
                </p>
                <ul>
                    <li>Sourcegraph version string</li>
                    <li>Deployment type (Docker, Kubernetes, or dev build)</li>
                    <li>Whether the instance is deployed on localhost (true/false)</li>
                    <li>Randomly generated site identifier</li>
                    <li>
                        The email address of the initial site installer (or if deleted, the first active site admin), to
                        know who to contact regarding sales, product updates, and policy updates
                    </li>
                    <li>
                        Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy,
                        SAML, GitHub, GitLab)
                    </li>
                    <li>
                        Which categories of external service are in use (GitHub, Bitbucket Server, GitLab, Phabricator,
                        Gitolite, AWS CodeCommit, Other)
                    </li>
                    <li>Whether new user signup is allowed (true/false)</li>
                    <li>Whether a repository has ever been added (true/false)</li>
                    <li>Whether a code search has ever been executed (true/false)</li>
                    <li>Whether code intelligence has ever been used (true/false)</li>
                    <li>Total count of existing user accounts</li>
                    <li>Aggregate counts of current daily, weekly, and monthly users</li>
                    <li>Aggregate counts of current users using code host integrations</li>
                    <li>
                        Aggregate counts of current users by product feature (site management, code search and
                        navigation, code review, saved searches, diff searches)
                    </li>
                </ul>
                {!pingsEnabled ? (
                    <p>Pings are disabled.</p>
                ) : (
                    <p>
                        To disable pings please {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/contact/" target="_blank">
                            contact support
                        </a>
                        .
                    </p>
                )}
            </div>
        )
    }
}
