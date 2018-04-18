import { matchPath } from 'react-router'
import { currentUser } from '../auth'
import * as GQL from '../backend/graphqlschema'
import { parseBrowserRepoURL } from '../repo'
import { repoRevRoute } from '../routes'
import { getPathExtension } from '../util'
import {
    browserExtensionMessageReceived,
    browserExtensionServerConfigurationMessageReceived,
    handleQueryEvents,
    pageViewQueryParameters,
} from './analyticsUtils'
import { serverAdmin } from './services/serverAdminWrapper'
import { telligent } from './services/telligentWrapper'

class EventLogger {
    private hasStrippedQueryParameters = false
    private user?: GQL.IUser | null

    constructor() {
        browserExtensionMessageReceived.subscribe(isInstalled => {
            telligent.setUserProperty('installed_chrome_extension', 'true')

            if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
                console.debug('%cBrowser extension detected, sync completed', 'color: #aaa')
            }

            // Subscribe to the currentUser Subject to send a success response back to the extension
            // right now, and on any future user changes.
            currentUser.subscribe(user => {
                const detail = { deviceId: telligent.getTelligentDuid(), userId: user ? user.email : undefined }
                document.dispatchEvent(
                    new CustomEvent('sourcegraph:identify', {
                        detail,
                    })
                )
            })
        })

        browserExtensionServerConfigurationMessageReceived.subscribe(() => {
            this.log('SourcegraphServerBrowserExtensionConfigureClicked')
        })

        currentUser.subscribe(
            user => {
                this.user = user
                if (user) {
                    this.updateUser(user)
                    this.log('UserProfileFetched')
                }
            },
            error => {
                /* noop */
            }
        )
    }

    /**
     * Set user-level properties in all external tracking services
     */
    public updateUser(
        user:
            | GQL.IUser
            | {
                  externalID: string
                  sourcegraphID: number | null
                  username: string | null
                  email: string | null
              }
    ): void {
        this.setUserIds(user.sourcegraphID, user.externalID || '', user.username)
        if (user.email) {
            this.setUserEmail(user.email)
        }
    }

    /**
     * Set user ID in Telligent tracker script.
     * @param uniqueSourcegraphId Unique Sourcegraph user ID (corresponds to User.ID from backend)
     * @param uniqueExternalID Unique user external auth provider ID
     * @param username Human-readable user identifier, not guaranteed to always stay the same
     */
    public setUserIds(uniqueSourcegraphId: number | null, uniqueExternalID: string, username: string | null): void {
        if (username) {
            telligent.setUserProperty('username', username)
        }
        if (uniqueSourcegraphId) {
            telligent.setUserProperty('user_id', uniqueSourcegraphId)
        }
        telligent.setUserProperty('internal_user_id', uniqueExternalID)
    }

    public setUserEmail(primaryEmail: string): void {
        telligent.setUserProperty('email', primaryEmail)
    }

    public setUserInvited(invitingUserId: string, invitedToOrg: string): void {
        telligent.setUserProperty('invited_by_user', invitingUserId)
        telligent.setUserProperty('org_invite', invitedToOrg)
    }

    /**
     * Log a pageview.
     * Page titles should be specific and human-readable in pascal case, e.g. "SearchResults" or "Blob" or "NewOrg"
     */
    public logViewEvent(pageTitle: string, eventProperties?: any, logUserEvent = true): void {
        if (window.context.userAgentIsBot || !pageTitle) {
            return
        }
        pageTitle = `View${pageTitle}`

        const decoratedProps = {
            ...this.decorateEventProperties(eventProperties),
            ...pageViewQueryParameters(window.location.href),
            page_name: pageTitle,
            page_title: pageTitle,
        }
        telligent.track('view', decoratedProps)
        if (logUserEvent) {
            serverAdmin.trackPageView()
        }
        this.logToConsole(pageTitle, decoratedProps)

        // Use flag to ensure URL query params are only stripped once
        if (!this.hasStrippedQueryParameters) {
            handleQueryEvents(window.location.href)
            this.hasStrippedQueryParameters = true
        }
    }

    /**
     * Log a user action or event.
     * Event labels should be specific and follow a ${noun}${verb} structure in pascal case, e.g. "ButtonClicked" or "SignInInitiated"
     */
    public log(eventLabel: string, eventProperties?: any): void {
        if (window.context.userAgentIsBot || !eventLabel) {
            return
        }

        const decoratedProps = {
            ...this.decorateEventProperties(eventProperties),
            eventLabel,
        }
        telligent.track(eventLabel, decoratedProps)
        serverAdmin.trackAction(eventLabel, decoratedProps)
        this.logToConsole(eventLabel, decoratedProps)
    }

    private logToConsole(eventLabel: string, object?: any): void {
        if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
            console.debug('%cEVENT %s', 'color: #aaa', eventLabel, object)
        }
    }

    private decorateEventProperties(eventProperties: any): any {
        const props = {
            ...eventProperties,
            is_authed: this.user ? 'true' : 'false',
            path_name: window.location && window.location.pathname ? window.location.pathname.slice(1) : '',
        }

        const match = matchPath<{ repoRev?: string; filePath?: string }>(window.location.pathname, repoRevRoute)
        if (match) {
            const u = parseBrowserRepoURL(window.location.href)
            props.repo = u.repoPath
            props.rev = u.rev
            if (u.filePath) {
                props.path = u.filePath
                props.language = getPathExtension(u.filePath)
            }
        }
        return props
    }

    /**
     * Access the Telligent unique user ID stored in a cookie on the user's computer. Cookie TTL is 2 years
     * https://sourcegraph.com/github.com/telligent-data/telligent-javascript-tracker@890d6a69b84fc0518a3e848f5469b34817da69fd/-/blob/src/js/tracker.js#L178
     */
    public uniqueUserCookieID(): string {
        return telligent.getTelligentDuid() || ''
    }
}

export const eventLogger = new EventLogger()
