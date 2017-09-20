import { currentUser } from '../auth'
import { parseBrowserRepoURL } from '../repo'
import { getPathExtension } from '../util'
import { sourcegraphContext } from '../util/sourcegraphContext'
import { EventActions, EventCategories } from './analyticsConstants'
import { telligent } from './services/telligentWrapper'

class EventLogger {
    private static PLATFORM = 'Web'

    constructor() {
        if (sourcegraphContext.user) {
            this.updateUser({ id: sourcegraphContext.user.UID })
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
            error => { /* noop */ }
        )
    }

    /**
     * Set user-level properties in all external tracking services
     */
    public updateUser(user: GQL.IUser | { id: string, email?: string }): void {
        // TODO(dan): update with correct handle, if we want one
        this.setUserId(user.id.toString(), user.email || '')
        if (user.email) {
            this.setUserEmail(user.email)
        }
    }

    /**
     * Function to sync our key user identification props across Telligent and user Chrome extension installations
     */
    public updateTrackerWithIdentificationProps(user: GQL.IUser): any {
        if (!telligent.isTelligentLoaded() || !sourcegraphContext.hasBrowserExtensionInstalled()) {
            return null
        }

        this.setUserInstalledChromeExtension('true')

        const idProps = { detail: { deviceId: telligent.getTelligentDuid(), userId: user.email } }
        setTimeout(() => document.dispatchEvent(new CustomEvent('sourcegraph:identify', idProps)), 20)
    }

    /**
     * Set user ID in Telligent tracker script.
     * TODO(Dan): determine whether we continue to use handles at the user level, or
     * if they're fully replaced by org-level handles/usernames. For now, handles
     * have been replaced with emails.
     * @param uniqueSourcegraphId Unique Sourcegraph user ID (corresponds to User.UID from backend)
     * @param handle Human-readable user identifier, not guaranteed to always stay the same. TODO: determine if we choose to use usernames or emails.
     */
    public setUserId(uniqueSourcegraphId: string, handle?: string): void {
        if (handle) {
            telligent.setUserId(handle)
        }
        telligent.setUserProperty('internal_user_id', uniqueSourcegraphId)
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
        if (sourcegraphContext.userAgentIsBot || !pageTitle) {
            return
        }

        const decoratedProps = { ...this.decorateEventProperties(eventProperties), page_name: pageTitle, page_title: pageTitle }
        telligent.track('view', decoratedProps)
        this.logToConsole(pageTitle, decoratedProps)
    }

    /**
     * Tracking call to analytics services on user action events
     * Note: should NEVER be called outside of events.tsx
     */
    public logEvent(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any): void {
        if (sourcegraphContext.userAgentIsBot || !eventLabel) {
            return
        }

        const decoratedProps = { ...this.decorateEventProperties(eventProperties), eventLabel, eventCategory, eventAction }
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
            is_authed: sourcegraphContext.user ? 'true' : 'false',
            path_name: window.location && window.location.pathname ? window.location.pathname.slice(1) : ''
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
