import { sortBy } from 'lodash-es'
import { ContributableMenu, Contributions } from 'sourcegraph/module/protocol'
import { ActionItemProps } from './ActionItem'

const MENU_ITEMS_PROP_SORT_ORDER = ['group', 'id']

/**
 * Collect all command contrbutions for the menu.
 *
 * @param prioritizeActions sort these actions first
 */
export function getContributedActionItems(contributions: Contributions, menu: ContributableMenu): ActionItemProps[] {
    if (!contributions.actions) {
        return []
    }

    const allItems: ActionItemProps[] = []
    const menuItems = contributions.menus && contributions.menus[menu]
    if (menuItems) {
        for (const { action: actionID, alt: altActionID } of sortBy(menuItems, MENU_ITEMS_PROP_SORT_ORDER)) {
            const action = contributions.actions.find(a => a.id === actionID)
            const altAction = contributions.actions.find(a => a.id === altActionID)
            if (action) {
                allItems.push({ action, altAction })
            }
        }
    }
    return allItems
}
