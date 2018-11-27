import { Observable, Subscription } from 'rxjs'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtConfigurationAPI } from '../../extension/api/configuration'
import { SettingsEdit } from '../../protocol'
import { Connection, ConnectionError, ConnectionErrors } from '../../protocol/jsonrpc2/connection'

/** @internal */
export interface ClientConfigurationAPI {
    $acceptConfigurationUpdate(params: SettingsEdit): Promise<void>
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ClientConfiguration<C> implements ClientConfigurationAPI {
    private subscriptions = new Subscription()
    private proxy: ExtConfigurationAPI<C>

    constructor(
        connection: Connection,
        environmentConfiguration: Observable<C>,
        private updateConfiguration: (params: SettingsEdit) => Promise<void>
    ) {
        this.proxy = createProxyAndHandleRequests('configuration', connection, this)

        this.subscriptions.add(
            environmentConfiguration.subscribe(config => {
                this.proxy.$acceptConfigurationData(config).catch(error => {
                    if (error instanceof ConnectionError && error.code === ConnectionErrors.Unsubscribed) {
                        // This error was probably caused by the user disabling
                        // an extension, which is a normal occurrence.
                        return
                    }
                    throw error
                })
            })
        )
    }

    public async $acceptConfigurationUpdate(params: SettingsEdit): Promise<void> {
        await this.updateConfiguration(params)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
