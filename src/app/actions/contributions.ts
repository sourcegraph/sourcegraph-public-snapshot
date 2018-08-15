import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { sortBy } from 'lodash'
import { ActionItemProps } from './ActionItem'

/** Collect all command contrbutions for the menu. */
export function getContributedActionItems(contributions: Contributions, menu: ContributableMenu): ActionItemProps[] {
    const allItems: ActionItemProps[] = []
    const menuItems = contributions.menus && contributions.menus[menu]
    if (menuItems) {
        for (const { action: actionID } of menuItems) {
            const action = contributions.actions && contributions.actions.find(a => a.id === actionID)
            if (action) {
                allItems.push({ contribution: action })
            }
        }
    }
    return sortBy(allItems, (item: ActionItemProps) => item.contribution.id)
}
