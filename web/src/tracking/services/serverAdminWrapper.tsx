import { currentUser } from '../../auth'
import { logUserEvent } from '../../user/settings/backend'

class ServerAdminWrapper {
    /**
     * isAuthenicated is a flag that indicates if a user is signed in.
     * We only log certain events (pageviews) if the user is not authenticated.
     */
    private isAuthenicated = false

    constructor() {
        if (!window.context.sourcegraphDotComMode) {
            currentUser.subscribe(user => {
                if (user) {
                    this.isAuthenicated = true
                }
            })
        }
    }

    public trackPageView(): void {
        logUserEvent('PAGEVIEW')
    }

    public trackAction(eventAction: string, eventProps: any): void {
        if (this.isAuthenicated) {
            if (eventAction === 'SearchSubmitted') {
                logUserEvent('SEARCHQUERY')
            } else if (
                eventAction === 'SymbolHovered' ||
                eventAction === 'FindRefsClicked' ||
                eventAction === 'GoToDefClicked'
            ) {
                logUserEvent('CODEINTEL')
            }
        }
    }
}

export const serverAdmin = new ServerAdminWrapper()
