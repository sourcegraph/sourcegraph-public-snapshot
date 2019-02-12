import * as GQL from '../../../../shared/src/graphql/schema'
import { authenticatedUser } from '../../auth'
import { logUserEvent } from '../../user/account/backend'

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

    public trackPageView(): void {
        logUserEvent(GQL.UserEvent.PAGEVIEW)
    }

    public trackAction(eventAction: string, eventProps: any): void {
        if (this.isAuthenicated) {
            if (eventAction === 'SearchSubmitted') {
                logUserEvent(GQL.UserEvent.SEARCHQUERY)
            } else if (
                eventAction === 'goToDefinition' ||
                eventAction === 'goToDefinition.preloaded' ||
                eventAction === 'findReferences'
            ) {
                logUserEvent(GQL.UserEvent.CODEINTEL)
            }
        }
    }
}

export const serverAdmin = new ServerAdminWrapper()
