import * as H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'

import { ContributableMenu } from '@sourcegraph/client-api'
import { syncRemoteSubscription } from '@sourcegraph/shared/src/api/util'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'

export function registerSearchStatsContributions({
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'extHostAPI'>): Unsubscribable {
    const subscriptions = new Subscription()

    const ACTION_ID = 'search.stats'

    if (extensionsController !== null && window.context.enableLegacyExtensions) {
        subscriptions.add(
            syncRemoteSubscription(
                extensionsController.extHostAPI.then(extensionHostAPI =>
                    extensionHostAPI.registerContributions({
                        actions: [
                            {
                                id: ACTION_ID,
                                title: 'Statistics',
                                category: 'Search',
                                command: 'open',
                                // eslint-disable-next-line no-template-curly-in-string
                                commandArguments: ['/stats?q=${get(context, "searchQuery")}'],
                                actionItem: {
                                    label: 'Stats',
                                },
                            },
                        ],
                        menus: {
                            [ContributableMenu.SearchResultsToolbar]: [
                                {
                                    action: ACTION_ID,
                                    when:
                                        'config.experimentalFeatures && get(config.experimentalFeatures, "searchStats")',
                                },
                            ],
                        },
                    })
                )
            )
        )
    }

    return subscriptions
}
