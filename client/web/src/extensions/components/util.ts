import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'

export function generateActionItemIcons(actionItems: ActionItemAction[]): ActionItemAction[] {
    for (const actionItem of actionItems) {
        actionItem.action.iconURL
    }
}
