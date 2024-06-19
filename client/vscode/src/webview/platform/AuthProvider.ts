import { commands, type SecretStorage } from 'vscode'

import { setEndpoint } from '../../settings/endpointSetting'

export const secretTokenKey = 'SOURCEGRAPH_AUTH'

export class SourcegraphAuthActions {
    constructor(private readonly secretStorage: SecretStorage) {}

    public async login(newtoken: string, newuri: string): Promise<void> {
        try {
            await this.secretStorage.store(secretTokenKey, newtoken)
            setEndpoint(newuri)
            return
        } catch (error) {
            console.error(error)
        }
    }

    public async logout(): Promise<void> {
        setEndpoint(undefined)
        await this.secretStorage.delete(secretTokenKey)
        await commands.executeCommand('workbench.action.reloadWindow')
        return
    }
}
