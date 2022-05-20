import React, { useEffect, useMemo } from 'react'

import { isEmpty, noop } from 'lodash'
import * as Monaco from 'monaco-editor'
import { RouteComponentProps } from 'react-router-dom'
import { fromFetch } from 'rxjs/fetch'

import { checkOk } from '@sourcegraph/http-client'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, Typography, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends RouteComponentProps, ThemeProps {}

/**
 * A page displaying information about telemetry pings for the site.
 */
export const SiteAdminPingsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const latestPing = useObservable(
        useMemo(
            () => fromFetch<{}>('/site-admin/pings/latest', { selector: response => checkOk(response).json() }),
            []
        )
    )
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminPings')
    }, [])

    const nonCriticalTelemetryDisabled = window.context.site.disableNonCriticalTelemetry === true
    const updatesDisabled = window.context.site['update.channel'] !== 'release'

    const options: Monaco.editor.IStandaloneEditorConstructionOptions = {
        readOnly: true,
        minimap: {
            enabled: false,
        },
        lineNumbers: 'off',
        fontSize: 14,
        glyphMargin: false,
        overviewRulerBorder: false,
        rulers: [],
        overviewRulerLanes: 0,
        wordBasedSuggestions: false,
        quickSuggestions: false,
        fixedOverflowWidgets: true,
        renderLineHighlight: 'none',
        contextmenu: false,
        links: false,
        // Display the cursor as a 1px line.
        cursorStyle: 'line',
        cursorWidth: 1,
    }
    return (
        <div className="site-admin-pings-page">
            <PageTitle title="Pings - Admin" />
            <Typography.H2>Pings</Typography.H2>
            <p>
                Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It
                sends only the high-level data below. It never sends code, repository names, usernames, or any other
                specific data.
            </p>
            <Typography.H3>Most recent ping</Typography.H3>
            {latestPing === undefined ? (
                <p>
                    <LoadingSpinner />
                </p>
            ) : isEmpty(latestPing) ? (
                <p>No recent ping data to display.</p>
            ) : (
                <MonacoEditor
                    {...props}
                    language="json"
                    options={options}
                    height={300}
                    editorWillMount={noop}
                    value={JSON.stringify(latestPing, undefined, 4)}
                    className="mb-3"
                />
            )}
            <Typography.H3>Critical telemetry</Typography.H3>
            <p>
                Critical telemetry includes only the high-level data below required for billing, support, updates, and
                security notices. This cannot be disabled.
            </p>
            <ul>
                <li>Randomly generated site identifier</li>
                <li>
                    The email address of the initial site installer (or if deleted, the first active site admin), to
                    know who to contact regarding sales, product updates, security updates, and policy updates
                </li>
                <li>Sourcegraph version string (e.g. "vX.X.X")</li>
                <li>Dependency versions (e.g. "6.0.9" for Redis, or "13.0" for Postgres)</li>
                <li>
                    Deployment type (single Docker image, Docker Compose, Kubernetes cluster, Helm, or pure Docker
                    cluster)
                </li>
                <li>License key associated with your Sourcegraph subscription</li>
                <li>Aggregate count of current monthly users</li>
                <li>Total count of existing user accounts</li>
                <li>Code Insights: total count of insights</li>
            </ul>
            <Typography.H3>Other telemetry</Typography.H3>
            <p>
                By default, Sourcegraph also aggregates usage and performance metrics for some product features. No
                personal or specific information is ever included.
            </p>
            <ul>
                <li>Whether the instance is deployed on localhost (true/false)</li>
                <li>
                    Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy, SAML,
                    GitHub, GitLab)
                </li>
                <li>
                    Which code hosts are in use (GitHub, Bitbucket Server, GitLab, Phabricator, Gitolite, AWS
                    CodeCommit, Other)
                    <ul>
                        <li>Which versions of the code hosts are used</li>
                    </ul>
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
                        <li>Search modes used (interactive search, plain-text search)</li>
                        <li>Search filters used (e.g. "type:", "repo:", "file:", "lang:", etc.)</li>
                    </ul>
                </li>
                <li>Aggregate daily, weekly, and monthly latencies (in ms) of search queries</li>
                <li>
                    Aggregate daily, weekly, and monthly counts of:
                    <ul>
                        <li>Code intelligence events (e.g., hover tooltips)</li>
                        <li>Searches using each search mode (interactive search, plain-text search)</li>
                        <li>Searches using each search filter (e.g. "type:", "repo:", "file:", "lang:", etc.)</li>
                    </ul>
                </li>
                <li>
                    Code intelligence usage data
                    <ul>
                        <li>Total number of repositories with and without an uploaded LSIF index</li>
                        <li>
                            Total number of code intelligence queries (e.g., hover tooltips) per week grouped by
                            language
                        </li>
                        <li>
                            Number of users performing code intelligence queries (e.g., hover tooltips) per week grouped
                            by language
                        </li>
                    </ul>
                </li>
                <li>
                    Batch changes usage data
                    <ul>
                        <li>Total count of page views on the batch change apply page</li>
                        <li>
                            Total count of page views on the batch change details page after creating a batch change
                        </li>
                        <li>
                            Total count of page views on the batch change details page after updating a batch change
                        </li>
                        <li>Total count of created changeset specs</li>
                        <li>Total count of created batch specs</li>
                        <li>Total count of created batch changes</li>
                        <li>Total count of closed batch changes</li>
                        <li>Total count of changesets created by batch changes</li>
                        <li>Aggregate counts of lines changed, added, deleted in changeset</li>
                        <li>Total count of changesets created by batch changes that have been merged</li>
                        <li>Aggregate counts of lines changed, added, deleted in merged changeset</li>
                        <li>Total count of changesets manually added to a batch change</li>
                        <li>Total count of changesets manually added to a batch change that have been merged</li>
                        <li>
                            Aggregate counts of unique monthly users, by:
                            <ul>
                                <li>Whether they have contributed to batch changes</li>
                                <li>Whether they only viewed batch changes</li>
                                <li>Whether they have performed a bulk operation</li>
                            </ul>
                        </li>
                        <li>
                            Weekly batch change (open, closed) and changesets counts (imported, published, unpublished,
                            open, draft, merged, closed) for batch change cohorts created in the last 12 months
                        </li>
                        <li>Weekly bulk operations count (grouped by operation)</li>
                        <li>Total count of executors connected</li>
                        <li>Cumulative executor runtime monthly</li>
                        <li>Total count of publish bulk operation</li>
                        <li>Total count of bulk operations (grouped by operation type)</li>
                        <li>
                            Changeset distribution for batch change (grouped by batch change source: local or executor)
                        </li>
                        <li>Total count of users that ran a job on an executor monthly</li>
                        <li>
                            Total count of published changesets and batch changes created via:
                            <ul>
                                <li>executor</li>
                                <li>local (using src-cli)</li>
                            </ul>
                        </li>
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
                <li>
                    Homepage panel engagement
                    <ul>
                        <li>Percentage of panel clicks (out of total views)</li>
                        <li>Total count of unique users engaging with the panels</li>
                    </ul>
                </li>
                <li>Weekly retention rates for user cohorts created in the last 12 weeks</li>
                <li>
                    Search onboarding engagement
                    <ul>
                        <li>Total number of views of the onboarding tour</li>
                        <li>Total number of views of each step in the onboarding tour</li>
                        <li>Total number of tours closed</li>
                    </ul>
                </li>
                <li>
                    Sourcegraph extension activation statistics
                    <ul>
                        <li>Total number of users that use a given non-default Sourcegraph extension</li>
                        <li>
                            Average number of activations for users that use a given non-default Sourcegraph extension
                        </li>
                        <li>Total number of users that use non-default Sourcegraph extensions</li>
                        <li>
                            Average number of non-default extensions enabled for users that use non-default Sourcegraph
                            extensions
                        </li>
                    </ul>
                </li>
                <li>
                    Code insights usage data
                    <ul>
                        <li>Total count of page views on the insights page</li>
                        <li>Count of unique viewers on the insights page</li>
                        <li>Total counts of hovers, clicks, and drags of insights by type (e.g. search, code stats)</li>
                        <li>Total counts of edits, additions, and removals of insights by type</li>
                        <li>
                            Total count of clicks on the "Add more insights" and "Configure insights" buttons on the
                            insights page
                        </li>
                        <li>
                            Weekly count of users that have created an insight, and count of users that have created
                            their first insight this week
                        </li>
                        <li>
                            Weekly count of total and unique views to the Create new insight, Create search insight, and
                            Create language insight pages
                        </li>
                        <li>
                            Weekly count of total and unique clicks of the Create search insight, Create language usage
                            insight, and Explore the extensions buttons on the Create new insight page
                        </li>
                        <li>
                            Weekly count of total and unique clicks of the Create and Cancel buttons on the Create
                            search insight and Create language insight pages
                        </li>
                        <li>Total count of insights grouped by time interval (step size) in days</li>
                        <li>Total count of insights set organization visible grouped by insight type</li>
                        <li>
                            Total count of insight views grouped by presentation type, series type, and
                            presentation-series type
                        </li>
                    </ul>
                </li>
                <li>
                    Code monitoring usage data
                    <ul>
                        <li>Total number of views of the code monitoring page</li>
                        <li>Total number of views of the create code monitor page</li>
                        <li>
                            Total number of views of the create code monitor page with a pre-populated trigger query
                        </li>
                        <li>
                            Total number of views of the create code monitor page without a pre-populated trigger query
                        </li>
                        <li>Total number of views of the manage code monitor page</li>
                        <li>Total number of clicks on the code monitor email search link</li>
                    </ul>
                </li>
                <li>
                    Notebooks usage data
                    <ul>
                        <li>Total number of views of the notebook page</li>
                        <li>Total number of views of the notebooks list page</li>
                        <li>Total number of views of the embedded notebook page</li>
                        <li>Total number of created notebooks</li>
                        <li>Total number of added notebook stars</li>
                        <li>Total number of added notebook markdown blocks</li>
                        <li>Total number of added notebook query blocks</li>
                        <li>Total number of added notebook file blocks</li>
                        <li>Total number of added notebook symbol blocks</li>
                        <li>Total number of added notebook compute blocks</li>
                    </ul>
                </li>
                <li>
                    CTA usage data
                    <ul>
                        <li>
                            Browser extension
                            <ul>
                                <li>
                                    Number of users who viewed / clicked the "install browser extension" CTA on the file
                                    / search pages today
                                </li>
                                <li>
                                    Number of views / clicks on the "install browser extension" CTA on the file / search
                                    pages today
                                </li>
                            </ul>
                        </li>
                        <li>
                            IDE extension
                            <ul>
                                <li>
                                    Number of users who viewed / clicked the "install IDE extension" CTA on the file /
                                    search pages today
                                </li>
                                <li>
                                    Number of views / clicks on the "install IDE extension" CTA on the file / search
                                    pages today
                                </li>
                            </ul>
                        </li>
                    </ul>
                </li>
                <li>
                    Code Host integration usage data (Browser extension / Native Integration)
                    <ul>
                        <li>Aggregate counts of current daily, weekly, and monthly unique users and total events</li>
                        <li>
                            Aggregate counts of current daily, weekly, and monthly unique users and total events who
                            visited Sourcegraph instance from browser extension
                        </li>
                    </ul>
                </li>
                <li>
                    IDE extensions data
                    <ul>
                        Aggregate counts of current daily, weekly, and monthly searches performed:
                        <li>
                            <ul>Count of unique users who performed searches</ul>
                            <ul>Count of total searches performed</ul>
                        </li>
                    </ul>
                    <ul>
                        Aggregate counts of daily user state:
                        <li>
                            <ul>Count of unique users who installed the extension</ul>
                            <ul>Count of unique users who uninstalled the extension</ul>
                        </li>
                    </ul>
                    <ul>Aggregate count of daily redirects from extension to Sourcegraph instance</ul>
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
