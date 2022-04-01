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
                                when: 'config.experimentalFeatures && get(config.experimentalFeatures, "searchStats")',
                            },
                        ],
                    },
                })
            )
        )
    )

    return subscriptions
}

export function registerBatchChangesContributions({
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'extHostAPI'>): Unsubscribable {
    const subscriptions = new Subscription()

    const ACTION_ID = 'batches.replaceSymbol'
    subscriptions.add(
        syncRemoteSubscription(
            extensionsController.extHostAPI.then(extensionHostAPI =>
                extensionHostAPI.registerContributions({
                    actions: [
                        {
                            id: ACTION_ID,
                            // label: 'Replace Symbol',
                            category: 'batches',
                            command: 'open',
                            commandArguments: [
                                // eslint-disable-next-line no-template-curly-in-string
                                '/batch-changes/create?kind=replaceSymbol&repo=${get(resource, "commit") || "undefined"}&q=${get(component, "selection.start.line")}&r=${get(component, "selection") || "undefined"}&s=${json(config)}',
                            ],
                            actionItem: {
                                label: 'Replace Symbol',
                                description: 'Changes the symbol for all references in this repo',
                                iconDescription: 'Batch Changes logo',
                                iconURL:
                                    "data:image/svg+xml;charset=UTF-8,%3csvg width='16' height='16' viewBox='0 0 32 32' fill='none' xmlns='http://www.w3.org/2000/svg'%3e%3cpath fill-rule='evenodd' clip-rule='evenodd' d='M5.829 6.76a1.932 1.932 0 100-3.863 1.932 1.932 0 000 3.863zm0 2.898a4.829 4.829 0 100-9.658 4.829 4.829 0 000 9.658z' fill='%231C7CD6'/%3e%3cpath d='M22.473 1.867H30.2v7.726h-7.726V1.867zM22.473 13.07H30.2v7.727h-7.726V13.07zM22.473 24.274H30.2V32h-7.726v-7.726z' fill='%231C7CD6'/%3e%3cpath fill-rule='evenodd' clip-rule='evenodd' d='M12.014 5.795c0-.8.648-1.449 1.448-1.449h5.795a1.449 1.449 0 110 2.897h-5.795c-.8 0-1.448-.648-1.448-1.448zM6.544 11.047a1.449 1.449 0 00-1.6 1.28l1.44.16-1.44-.16v.011l-.003.023-.008.08c-.006.066-.015.162-.024.283-.018.242-.04.587-.055 1.013a28.23 28.23 0 00.087 3.36c.226 2.602.915 6.018 2.937 8.546 2.08 2.599 5.13 3.566 7.48 3.918a18.29 18.29 0 003.957.15c.111-.008.2-.017.263-.023l.076-.008.023-.003h.008l.003-.001s.002 0-.178-1.438l.18 1.438a1.449 1.449 0 00-.358-2.875M7.824 12.646l-.001.012-.006.055-.02.231a25.333 25.333 0 00.03 3.902c.212 2.43.835 5.14 2.314 6.987 1.42 1.776 3.62 2.56 5.646 2.863a15.408 15.408 0 003.303.127 7.78 7.78 0 00.193-.017l.043-.005h.006M6.544 11.046a1.449 1.449 0 011.28 1.6l-1.28-1.6z' fill='%231C7CD6'/%3e%3cpath fill-rule='evenodd' clip-rule='evenodd' d='M5.692 11.214a1.449 1.449 0 00-.58 1.965l1.272-.692-1.272.692v.002l.002.002.003.006.008.014.023.04a8.703 8.703 0 00.353.551 12.492 12.492 0 005.602 4.416c2.047.807 4.203 1.038 5.803 1.079a21.55 21.55 0 001.986-.04 16.55 16.55 0 00.742-.067l.047-.006.014-.002h.008l-.193-1.436.192 1.435a1.45 1.45 0 00-.383-2.871h-.002l-.027.003a13.35 13.35 0 01-.594.053c-.416.028-1.012.052-1.716.035-1.424-.037-3.205-.244-4.814-.878a9.594 9.594 0 01-4.286-3.373 5.756 5.756 0 01-.221-.345l-.005-.008a1.449 1.449 0 00-1.962-.575z' fill='%231C7CD6'/%3e%3c/svg%3e ",
                            },
                        },
                    ],
                    menus: {
                        [ContributableMenu.Hover]: [
                            {
                                action: ACTION_ID,
                                // when: 'config.experimentalFeatures && get(config.experimentalFeatures, "searchStats")',
                            },
                        ],
                    },
                })
            )
        )
    )

    return subscriptions
}
