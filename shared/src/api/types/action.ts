import * as sourcegraph from 'sourcegraph'

export type Action = sourcegraph.Action

export function fromAction(action: sourcegraph.Action): Action {
    return action
}

export function toAction(action: Action): sourcegraph.Action {
    return action
}
