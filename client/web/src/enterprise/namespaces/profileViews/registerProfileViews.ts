import { Subscription, Unsubscribable } from 'rxjs'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import H from 'history'
import { namespaceSavedSearches } from './namespaceSavedSearches'
import { namespaceCampaigns } from './namespaceCampaigns'
import { namespaceGraphs } from './namespaceGraphs'

export const registerProfileViews = ({
    extensionsController: { services },
    history,
}: ExtensionsControllerProps & { history: H.History }): Unsubscribable => {
    const subscription = new Subscription()

    subscription.add(
        services.view.register('profileView.savedSearches', ContributableViewContainer.Profile, context =>
            namespaceSavedSearches(context)
        )
    )

    subscription.add(
        services.view.register('profileView.campaigns', ContributableViewContainer.Profile, context =>
            namespaceCampaigns(context)
        )
    )

    subscription.add(
        services.view.register('profileView.graphs', ContributableViewContainer.Profile, context =>
            namespaceGraphs(context)
        )
    )

    return subscription
}
