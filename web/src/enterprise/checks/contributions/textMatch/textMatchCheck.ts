import AlphabeticalIcon from 'mdi-react/AlphabeticalIcon'
import { Subscription, Unsubscribable } from 'rxjs'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { CHECK_TYPES } from '../sampleCheckTypes'

const PROVIDER_ID = 'check.textMatch'

/**
 * Registers the text match check provider.
 *
 * @internal
 */
export function registerTextMatchCheckProviderContributions({
    extensionsController,
}: ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        extensionsController.services.checkTemplates.registerProvider(
            {},
            {
                id: 'check.textMatch',
                title: 'Text find/replace',
                description: 'Find a string (and optionally replace it with another string)',
                icon: AlphabeticalIcon,
                settings: {
                    providers: PROVIDER_ID,
                },
            }
        )
    )
    for (const t of CHECK_TYPES) {
        subscriptions.add(extensionsController.services.checkTemplates.registerProvider({}, t))
    }
    return subscriptions
}
