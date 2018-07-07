import { ConfigurationItem, ConfigurationParams, ConfigurationRequest } from '../protocol'
import { _RemoteWorkspace } from './features/workspace'
import { Feature } from './server'

export interface Configuration {
    getConfiguration(): Promise<any>
    // tslint:disable-next-line:unified-signatures
    getConfiguration(section: string): Promise<any>
    // tslint:disable-next-line:unified-signatures
    getConfiguration(item?: ConfigurationItem): Promise<any>
    getConfiguration(items?: ConfigurationItem[]): Promise<any[]>
}

export const ConfigurationFeature: Feature<_RemoteWorkspace, Configuration> = Base =>
    class extends Base {
        public getConfiguration(arg?: string | ConfigurationItem | ConfigurationItem[]): Promise<any> {
            if (!arg) {
                return this._getConfiguration({})
            } else if (typeof arg === 'string') {
                return this._getConfiguration({ section: arg })
            } else {
                return this._getConfiguration(arg)
            }
        }

        private _getConfiguration(arg: ConfigurationItem | ConfigurationItem[]): Promise<any> {
            const params: ConfigurationParams = {
                items: Array.isArray(arg) ? arg : [arg],
            }
            return this.connection
                .sendRequest(ConfigurationRequest.type, params)
                .then(result => (Array.isArray(arg) ? result : result[0]))
        }
    }
