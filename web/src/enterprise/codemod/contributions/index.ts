import H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { registerCodemodSearchContributions } from './search'

export function registerCodemodContributions({
    history,
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(registerCodemodSearchContributions({ history, extensionsController }))
    return subscriptions
}
