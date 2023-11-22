import { sortBy } from 'lodash'

import type { ContributableMenu, Contributions, Evaluated } from '@sourcegraph/client-api'

import type { ActionItemAction } from '../actions/ActionItem'

const MENU_ITEMS_PROP_SORT_ORDER = ['group', 'id']

/**
 * Collect all command contributions for the menu.
 *
 * @param prioritizeActions sort these actions first
 */
export function getContributedActionItems(
    contributions: Evaluated<Contributions>,
    menu: ContributableMenu
): ActionItemAction[] {
    if (!contributions.actions) {
        return []
    }

    const allItems: ActionItemAction[] = []
    const menuItems = contributions.menus?.[menu]
    if (menuItems) {
        for (const { action: actionID, alt: altActionID, when, disabledWhen } of sortBy(
            menuItems,
            MENU_ITEMS_PROP_SORT_ORDER
        )) {
            const action = contributions.actions.find(a => a.id === actionID)
            const altAction = contributions.actions.find(a => a.id === altActionID)
            if (action) {
                allItems.push({ action, altAction, active: when === undefined || !!when, disabledWhen: !!disabledWhen })
            }
        }
    }

    return allItems
}
