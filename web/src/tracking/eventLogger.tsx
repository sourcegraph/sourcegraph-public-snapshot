import { currentUser } from '../auth'
import { parseBrowserRepoURL } from '../repo'
import { getPathExtension } from '../util'
import { browserExtensionMessageReceived, handleQueryEvents, pageViewQueryParameters } from './analyticsUtils'
import { serverAdmin } from './services/serverAdminWrapper'
import { telligent } from './services/telligentWrapper'

class EventLogger {
    private hasStrippedQueryParameters = false

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

        // tslint:disable-next-line deprecation
        if (window.context.user) {
            // TODO(dan): update with sourcegraphID from JS Context once available
            //
            // tslint:disable-next-line deprecation
            this.updateUser({ auth0ID: window.context.user.UID, sourcegraphID: null, username: null, email: null })
        }

        currentUser.subscribe(
            user => {
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
            | { auth0ID: string; sourcegraphID: number | null; username: string | null; email: string | null }
    ): void {
        this.setUserIds(user.sourcegraphID, user.auth0ID, user.username)
        if (user.email) {
            this.setUserEmail(user.email)
        }
    }

    /**
     * Set user ID in Telligent tracker script.
     * @param uniqueSourcegraphId Unique Sourcegraph user ID (corresponds to User.ID from backend)
     * @param uniqueAuthId Unique user auth ID (corresponds to Actor.UID or User.AuthID from backend)
     * @param username Human-readable user identifier, not guaranteed to always stay the same
     */
    public setUserIds(uniqueSourcegraphId: number | null, uniqueAuthId: string, username: string | null): void {
        if (username) {
            telligent.setUserProperty('username', username)
        }
        if (uniqueSourcegraphId) {
            telligent.setUserProperty('user_id', uniqueSourcegraphId)
        }
        telligent.setUserProperty('internal_user_id', uniqueAuthId)
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
    public logViewEvent(pageTitle: string, eventProperties?: any): void {
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
        serverAdmin.trackPageView()
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
            // tslint:disable-next-line deprecation
            is_authed: window.context.user ? 'true' : 'false',
            path_name: window.location && window.location.pathname ? window.location.pathname.slice(1) : '',
        }

        try {
            // TODO: This will not work on repo pages like sourcegraph.mycompany.com/foo/bar
            // because it does not start with github. But removing the startsWith below means
            // sourcegraph.mycompany.com/c/01BYDGS45FJ1XE91M2GGR2WTAC would find a repository
            // "c/01BYDGS45FJ1XE91M2GGR2WTAC" which isn't what you want either. This code
            // should be factored out to only be invoked on real repo pages according to the
            // router.
            const u = parseBrowserRepoURL(window.location.href)
            if (u.repoPath.startsWith('github.com/')) {
                props.repo = u.repoPath
                props.rev = u.rev
                if (u.filePath) {
                    props.path = u.filePath
                    props.language = getPathExtension(u.filePath)
                }
            }
        } catch (error) {
            // no-op
        }

        return props
    }
}

export const eventLogger = new EventLogger()
