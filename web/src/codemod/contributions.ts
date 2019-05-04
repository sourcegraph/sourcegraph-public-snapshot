import H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'
import { ContributableMenu } from '../../../shared/src/api/protocol'
import { urlForOpenPanel } from '../../../shared/src/commands/commands'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'

export const CODEMOD_PANEL_VIEW_ID = 'codemod'

const REPLACE_ID = 'codemod.search.replace'

export function registerCodemodContributions({
    location,
    history,
    extensionsController,
}: {
    location: H.Location
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()

    subscriptions.add(
        extensionsController.services.commands.registerCommand({
            command: REPLACE_ID,
            run: async () => {
                const text = prompt('Enter replacement text:')
                if (text !== null) {
                    history.push(urlForOpenPanel(CODEMOD_PANEL_VIEW_ID, location.hash))
                }
            },
        })
    )

    subscriptions.add(
        extensionsController.services.contribution.registerContributions({
            contributions: {
                actions: [
                    {
                        id: REPLACE_ID,
                        title: 'Replace',
                        category: 'Codemod',
                        command: REPLACE_ID,
                        actionItem: {
                            label: 'Replace',
                            // TODO!(sqs): icon theme color doesn't update
                            iconURL:
                                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' style='width:24px;height:24px' viewBox='0 0 24 24'%3E%3Cpath fill='%23a2b0cd' d='M11,6C12.38,6 13.63,6.56 14.54,7.46L12,10H18V4L15.95,6.05C14.68,4.78 12.93,4 11,4C7.47,4 4.57,6.61 4.08,10H6.1C6.56,7.72 8.58,6 11,6M16.64,15.14C17.3,14.24 17.76,13.17 17.92,12H15.9C15.44,14.28 13.42,16 11,16C9.62,16 8.37,15.44 7.46,14.54L10,12H4V18L6.05,15.95C7.32,17.22 9.07,18 11,18C12.55,18 14,17.5 15.14,16.64L20,21.5L21.5,20L16.64,15.14Z' /%3E%3C/svg%3E",
                        },
                    },
                ],
                menus: {
                    [ContributableMenu.SearchResultsToolbar]: [{ action: REPLACE_ID }],
                },
            },
        })
    )

    return subscriptions
}
