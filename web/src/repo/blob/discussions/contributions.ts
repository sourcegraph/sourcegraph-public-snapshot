import { Subscription, Unsubscribable } from 'rxjs'
import { registerDiscussionsMentionCompletionContributions } from './mentionCompletion'

/**
 * Registers contributions for discussions-related functionality.
 */
export function registerDiscussionsContributions(
    args: Parameters<typeof registerDiscussionsMentionCompletionContributions>[0]
): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(registerDiscussionsMentionCompletionContributions(args))
    return subscriptions
}
