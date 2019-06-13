import H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'

export function registerCodemodSampleProviderContributions({
    history,
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()
    // TODO!(sqs)
    return subscriptions
}
