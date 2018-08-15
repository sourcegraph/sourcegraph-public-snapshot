import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { sortBy } from 'lodash-es'
import { ActionItemProps } from './ActionItem'

const MENU_ITEMS_PROP_SORT_ORDER = ['group', 'id']

/**
 * Collect all command contrbutions for the menu.
 *
 * @param prioritizeActions sort these actions first
 */
export function getContributedActionItems(contributions: Contributions, menu: ContributableMenu): ActionItemProps[] {
    const allItems: ActionItemProps[] = []
    const menuItems = contributions.menus && contributions.menus[menu]
    if (menuItems) {
        for (const { action: actionID } of sortBy(menuItems, MENU_ITEMS_PROP_SORT_ORDER)) {
            const action = contributions.actions && contributions.actions.find(a => a.id === actionID)
            if (action) {
                allItems.push({ contribution: action })
            }
        }
    }
    return allItems
}
