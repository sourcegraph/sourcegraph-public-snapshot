import * as GQL from '../../../../shared/src/graphql/schema'
import { authenticatedUser } from '../../auth'
import { logUserEvent } from '../../user/settings/backend'

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
        logUserEvent(GQL.UserEvent.PAGEVIEW)
        if (this.isAuthenicated) {
            if (eventAction === 'ViewRepository' || eventAction === 'ViewBlob' || eventAction === 'ViewTree') {
                logUserEvent(GQL.UserEvent.STAGECODE)
            }
        }
    }

    public trackAction(eventAction: string): void {
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
            } else if (eventAction === 'findReferences') {
                logUserEvent(GQL.UserEvent.CODEINTELREFS)
            } else if (eventAction === 'SavedSearchEmailClicked' || eventAction === 'SavedSearchSlackClicked') {
                logUserEvent(GQL.UserEvent.STAGEVERIFY)
            } else if (eventAction === 'DiffSearchResultsQueried') {
                logUserEvent(GQL.UserEvent.STAGEMONITOR)
            }
        }
    }
}

export const serverAdmin = new ServerAdminWrapper()
