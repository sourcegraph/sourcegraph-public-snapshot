import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { from, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { isSettingsValid } from '../../../settings/settings'
import { ExtConfigurationAPI } from '../../extension/api/configuration'
import { SettingsEdit, SettingsService } from '../services/settings'

/** @internal */
export interface ClientConfigurationAPI extends ProxyValue {
    $acceptConfigurationUpdate(edit: SettingsEdit): void
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ClientConfiguration<C> implements ClientConfigurationAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()

    constructor(private proxy: ProxyResult<ExtConfigurationAPI<C>>, private settingsService: SettingsService<C>) {
        this.subscriptions.add(
            from(settingsService.data)
                .pipe(
                    switchMap(settings => {
                        // Only send valid settings.
                        //
                        // TODO(sqs): This could cause problems where the settings seen by extensions will lag behind
                        // settings seen by the client.
                        if (isSettingsValid(settings)) {
                            return this.proxy.$acceptConfigurationData(settings)
                        }
                        return []
                    })
                )
                .subscribe()
        )
    }

    public async $acceptConfigurationUpdate(edit: SettingsEdit): Promise<void> {
        await this.settingsService.update(edit)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
