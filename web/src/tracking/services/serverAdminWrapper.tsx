import { authenticatedUser } from '../../auth'
import { logUserEvent } from '../../user/settings/backend'
import { UserEvent } from '../../../../browser/src/shared/backend/userEvents'

class ServerAdminWrapper {
    /**
     * isAuthenicated is a flag that indicates if a user is signed in.
     * We only log certain events (pageviews) if the user is not authenticated.
     */
    private isAuthenicated = false

    constructor() {
        if (window.context && !window.context.sourcegraphDotComMode) {
            authenticatedUser.subscribe(user => {
                if (user) {
                    this.isAuthenicated = true
                }
            })
        }
    }

    public trackPageView(eventAction: string): void {
        logUserEvent(UserEvent.PAGEVIEW)
        if (this.isAuthenicated) {
            if (eventAction === 'ViewRepository' || eventAction === 'ViewBlob' || eventAction === 'ViewTree') {
                logUserEvent(UserEvent.STAGECODE)
            }
        }
    }

    public trackAction(eventAction: string): void {
        if (this.isAuthenicated) {
            if (eventAction === 'SearchResultsQueried') {
                logUserEvent(UserEvent.SEARCHQUERY)
                logUserEvent(UserEvent.STAGECODE)
            } else if (
                eventAction === 'goToDefinition' ||
                eventAction === 'goToDefinition.preloaded' ||
                eventAction === 'hover'
            ) {
                logUserEvent(UserEvent.CODEINTEL)
            } else if (eventAction === 'SavedSearchEmailClicked' || eventAction === 'SavedSearchSlackClicked') {
                logUserEvent(UserEvent.STAGEVERIFY)
            } else if (eventAction === 'DiffSearchResultsQueried') {
                logUserEvent(UserEvent.STAGEMONITOR)
            }
        }
    }
}

export const serverAdmin = new ServerAdminWrapper()
