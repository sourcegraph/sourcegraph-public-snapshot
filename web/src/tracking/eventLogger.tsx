import { currentUser } from '../auth'
import { parseBrowserRepoURL } from '../repo'
import { getPathExtension } from '../util'
import { EventActions, EventCategories } from './analyticsConstants'
import { handleQueryEvents, hasBrowserExtensionInstalled } from './analyticsUtils'
import { telligent } from './services/telligentWrapper'

class EventLogger {
    private static PLATFORM = 'Web'
    private hasStrippedQueryParameters = false

    constructor() {
        if (window.context.user) {
            // TODO(dan): update with sourcegraphID from JS Context once available
            this.updateUser({ id: window.context.user.UID, sourcegraphID: null, username: null, email: null })
        }

        currentUser.subscribe(
            user => {
                if (user) {
                    this.updateUser(user)
                    this.logEvent(EventCategories.Auth, EventActions.Passive, 'UserProfileFetched')
                    // Since this function checks if the Chrome ext has injected an element,
                    // wait a few ms in case there's an unpredictable delay before checking.
                    setTimeout(() => this.updateTrackerWithIdentificationProps(user), 100)
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
        user: GQL.IUser | { id: string; sourcegraphID: number | null; username: string | null; email: string | null }
    ): void {
        this.setUserIds(user.sourcegraphID, user.id, user.username)
        if (user.email) {
            this.setUserEmail(user.email)
        }
    }

    /**
     * Function to sync our key user identification props across Telligent and user Chrome extension installations
     */
    public updateTrackerWithIdentificationProps(user: GQL.IUser): any {
        if (!telligent.isTelligentLoaded() || !hasBrowserExtensionInstalled()) {
            return null
        }

        this.setUserInstalledChromeExtension('true')

        const idProps = { detail: { deviceId: telligent.getTelligentDuid(), userId: user.email } }
        setTimeout(() => document.dispatchEvent(new CustomEvent('sourcegraph:identify', idProps)), 20)
    }

    /**
     * Set user ID in Telligent tracker script.
     * @param uniqueSourcegraphId Unique Sourcegraph user ID (corresponds to User.ID from backend)
     * @param uniqueAuth0Id Unique Auth0 user ID (corresponds to Actor.UID or User.Auth0ID from backend)
     * @param username Human-readable user identifier, not guaranteed to always stay the same
     */
    public setUserIds(uniqueSourcegraphId: number | null, uniqueAuth0Id: string, username: string | null): void {
        if (username) {
            telligent.setUsername(username)
        }
        if (uniqueSourcegraphId) {
            telligent.setUserProperty('user_id', uniqueSourcegraphId)
        }
        telligent.setUserProperty('internal_user_id', uniqueAuth0Id)
    }

    public setUserInstalledChromeExtension(installedChromeExtension: string): void {
        telligent.setUserProperty('installed_chrome_extension', installedChromeExtension)
    }

    public setUserEmail(primaryEmail: string): void {
        telligent.setUserProperty('email', primaryEmail)
    }

    public setUserInvited(invitingUserId: string, invitedToOrg: string): void {
        telligent.setUserProperty('invited_by_user', invitingUserId)
        telligent.setUserProperty('org_invite', invitedToOrg)
    }

    /**
     * Tracking call to analytics services on pageview events
     * Note: should NEVER be called outside of events.tsx
     */
    public logViewEvent(pageTitle: string, eventProperties?: any): void {
        if (window.context.userAgentIsBot || !pageTitle) {
            return
        }

        const decoratedProps = {
            ...this.decorateEventProperties(eventProperties),
            page_name: pageTitle,
            page_title: pageTitle,
        }
        telligent.track('view', decoratedProps)
        this.logToConsole(pageTitle, decoratedProps)

        if (!this.hasStrippedQueryParameters) {
            handleQueryEvents(window.location.href)
            this.hasStrippedQueryParameters = true
        }
    }

    /**
     * Tracking call to analytics services on user action events
     * Note: should NEVER be called outside of events.tsx
     */
    public logEvent(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any): void {
        if (window.context.userAgentIsBot || !eventLabel) {
            return
        }

        const decoratedProps = {
            ...this.decorateEventProperties(eventProperties),
            eventLabel,
            eventCategory,
            eventAction,
        }
        telligent.track(eventAction, decoratedProps)
        this.logToConsole(eventLabel, decoratedProps)
    }

    private logToConsole(eventLabel: string, object?: any): void {
        if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
            console.debug('%cEVENT %s', 'color: #aaa', eventLabel, object)
        }
    }

    private decorateEventProperties(platformProperties: any): any {
        const props = {
            ...platformProperties,
            platform: EventLogger.PLATFORM,
            is_authed: window.context.user ? 'true' : 'false',
            path_name: window.location && window.location.pathname ? window.location.pathname.slice(1) : '',
        }

        try {
            const u = parseBrowserRepoURL(window.location.href)
            props.repo = u.repoPath!
            props.rev = u.rev
            if (u.filePath) {
                props.path = u.filePath!
                props.language = getPathExtension(u.filePath)
            }
        } catch (error) {
            // no-op
        }

        return props
    }
}

export const eventLogger = new EventLogger()
