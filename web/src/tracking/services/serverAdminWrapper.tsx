import { currentUser } from '../../auth'
import { logUserEvent } from '../../settings/backend'

class ServerAdminWrapper {
    /**
     * active is a flag to determine whether we want to log events
     */
    private active = false

    constructor() {
        if (window.context.onPrem) {
            currentUser.subscribe(user => {
                if (user) {
                    if (!this.active) {
                        // Handles initial page load
                        logUserEvent('PAGEVIEW').subscribe()
                    }
                    this.active = true
                }
            })
        }
    }

    public trackPageView(): void {
        if (this.active) {
            logUserEvent('PAGEVIEW').subscribe()
        }
    }

    public trackAction(eventAction: string, eventProps: any): void {
        if (this.active) {
            if (eventAction === 'SearchSubmitted') {
                logUserEvent('SEARCHQUERY').subscribe()
            }
        }
    }
}

export const serverAdmin = new ServerAdminWrapper()
