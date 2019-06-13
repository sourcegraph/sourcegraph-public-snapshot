import H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'
import { parseContributionExpressions } from '../../../../../shared/src/api/client/services/contribution'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createThread } from '../../../discussions/backend'
import { parseSearchURLQuery } from '../../../search'

/**
 * Registers contributions for checks functionality related to search.
 *
 * @internal
 */
export function registerChecksSearchContributions(
    args: Parameters<typeof registerSearchContextBarActions>[0]
): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(registerSearchContextBarActions(args))
    return subscriptions
}

function registerSearchContextBarActions({
    history,
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()

    const SAVE_ID = 'checks.search.saveAsCheck'
    subscriptions.add(
        extensionsController.services.commands.registerCommand({
            command: SAVE_ID,
            run: async () => {
                const title = prompt('Enter title to save:')
                if (title !== null) {
                    const query = parseSearchURLQuery(history.location.search)
                    const thread = await createThread({
                        title,
                        settings: JSON.stringify({ queries: [query] }),
                        contents: '',
                        type: GQL.ThreadType.CHECK,
                    }).toPromise()
                    history.push(thread.url)
                }
            },
        })
    )
    subscriptions.add(
        extensionsController.services.contribution.registerContributions({
            contributions: parseContributionExpressions({
                actions: [
                    {
                        id: SAVE_ID,
                        title: 'Save query to checks',
                        category: 'Checks',
                        command: SAVE_ID,
                        actionItem: {
                            label: 'Save query',
                            // TODO!(sqs): icon theme color doesn't update
                            iconURL:
                                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' style='width:24px;height:24px' viewBox='0 0 24 24' fill='%23a2b0cd'%3E%3Cpath d='M20,16V10H22V16C22,17.1 21.1,18 20,18H8C6.89,18 6,17.1 6,16V4C6,2.89 6.89,2 8,2H16V4H8V16H20M10.91,7.08L14,10.17L20.59,3.58L22,5L14,13L9.5,8.5L10.91,7.08M16,20V22H4C2.9,22 2,21.1 2,20V7H4V20H16Z'%3E%3C/path%3E%3C/svg%3E",
                        },
                    },
                ],
                menus: {
                    [ContributableMenu.SearchResultsToolbar]: [{ action: SAVE_ID }],
                },
            }),
        })
    )

    return subscriptions
}
