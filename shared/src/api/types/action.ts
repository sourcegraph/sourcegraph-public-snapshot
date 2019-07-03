import * as sourcegraph from 'sourcegraph'

export type Action = sourcegraph.Action

export type ActionType = {
    plan: { plan: sourcegraph.Plan }
    command: { command: sourcegraph.Command }
}

export function fromAction(action: sourcegraph.Action): Action {
    return action
}

export function toAction(action: Action): sourcegraph.Action {
    return action
}

export function isActionType(type: 'plan'): (action: Action) => action is { plan: sourcegraph.Plan }
export function isActionType(type: 'command'): (action: Action) => action is { command: sourcegraph.Command }
export function isActionType(type: string): (action: Action) => boolean {
    return action => !!(action as any)[type]
}
