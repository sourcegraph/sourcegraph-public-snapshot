import { registerEditsBehaviorCommand } from './behaviors/edits'
import { CHANGESET_BEHAVIOR_BY_REPOSITORY_AND_BASE_BRANCH_COMMAND } from './behaviors/edits/byRepositoryAndBaseBranch'
import { Unsubscribable, Subscription } from 'rxjs'

export const register = (): Unsubscribable => {
    const subscription = new Subscription()
    subscription.add(registerEditsBehaviorCommand(...CHANGESET_BEHAVIOR_BY_REPOSITORY_AND_BASE_BRANCH_COMMAND))
    return subscription
}
