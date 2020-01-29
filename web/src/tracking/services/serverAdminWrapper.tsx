import * as GQL from '../../../../shared/src/graphql/schema'
import { authenticatedUser } from '../../auth'
import { logUserEvent, logEvent } from '../../user/settings/backend'

class ServerAdminWrapper {
    /**
     * isAuthenicated is a flag that indicates if a user is signed in.
     */
    private isAuthenicated = false

    constructor() {
        authenticatedUser.subscribe(user => {
            if (user) {
                this.isAuthenicated = true
            }
        })
    }

    public trackPageView(eventAction: string, logAsActiveUser: boolean = true): void {
        if (logAsActiveUser) {
            logUserEvent(GQL.UserEvent.PAGEVIEW)
        }
        if (this.isAuthenicated) {
            if (eventAction === 'ViewRepository' || eventAction === 'ViewBlob' || eventAction === 'ViewTree') {
                logUserEvent(GQL.UserEvent.STAGECODE)
            }
        }
        logEvent(eventAction)
    }

    public trackAction(eventAction: string, eventProperties?: any): void {
        if (this.isAuthenicated) {
            if (eventAction === 'SearchResultsQueried') {
                logUserEvent(GQL.UserEvent.SEARCHQUERY)
                logUserEvent(GQL.UserEvent.STAGECODE)
            } else if (
                eventAction === 'goToDefinition' ||
                eventAction === 'goToDefinition.preloaded' ||
                eventAction === 'hover'
            ) {
                logUserEvent(GQL.UserEvent.CODEINTEL)
            } else if (eventAction === 'SavedSearchEmailClicked' || eventAction === 'SavedSearchSlackClicked') {
                logUserEvent(GQL.UserEvent.STAGEVERIFY)
            } else if (eventAction === 'DiffSearchResultsQueried') {
                logUserEvent(GQL.UserEvent.STAGEMONITOR)
            }
        }
        logEvent(eventAction, eventProperties)
    }
}

export const serverAdmin = new ServerAdminWrapper()
