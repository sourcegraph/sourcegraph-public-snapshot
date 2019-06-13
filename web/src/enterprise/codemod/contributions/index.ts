import H from 'history'
import { Subscription, Unsubscribable } from 'rxjs'
import { USE_CODEMOD } from '..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { registerCodemodSampleProviderContributions } from './sampleProviders'
import { registerCodemodSearchContributions } from './search'

export function registerCodemodContributions({
    history,
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    if (!USE_CODEMOD) {
        return { unsubscribe: () => void 0 }
    }

    const subscriptions = new Subscription()
    subscriptions.add(registerCodemodSearchContributions({ history, extensionsController }))
    subscriptions.add(registerCodemodSampleProviderContributions({ history, extensionsController }))
    return subscriptions
}
