import { from, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { isSettingsValid } from '../../../settings/settings'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtConfigurationAPI } from '../../extension/api/configuration'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { SettingsEdit, SettingsService } from '../services/settings'

/** @internal */
export interface ClientConfigurationAPI {
    $acceptConfigurationUpdate(edit: SettingsEdit): Promise<void>
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ClientConfiguration<C> implements ClientConfigurationAPI {
    private subscriptions = new Subscription()
    private proxy: ExtConfigurationAPI<C>

    constructor(connection: Connection, private settingsService: SettingsService<C>) {
        this.proxy = createProxyAndHandleRequests('configuration', connection, this)

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
