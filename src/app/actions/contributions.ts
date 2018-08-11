import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { sortBy } from 'lodash'
import { ActionItemProps } from './ActionItem'

/** Collect all command contrbutions for the menu. */
export function getContributedActionItems(contributions: Contributions, menu: ContributableMenu): ActionItemProps[] {
    const allItems: ActionItemProps[] = []
    const menuItems = contributions.menus && contributions.menus[menu]
    if (menuItems) {
        for (const { command: commandID } of menuItems) {
            const command = contributions.commands && contributions.commands.find(c => c.command === commandID)
            if (command) {
                allItems.push({ contribution: command })
            }
        }
    }
    return sortBy(allItems, (item: ActionItemProps) => item.contribution.command)
}
