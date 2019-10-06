import { ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { Unsubscribable } from 'rxjs'
import { ClientCommandsAPI } from '../../client/api/commands'
import { syncSubscription } from '../../util'
import { WorkspaceEdit } from '../../types/workspaceEdit'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { toDiagnostic } from '../../types/diagnostic'

interface CommandEntry {
    command: string
    callback: (...args: any[]) => any
}

/** @internal */
export class ExtCommands {
    constructor(private proxy: ProxyResult<ClientCommandsAPI>) {}

    /**
     * Extension API method invoked directly when an extension wants to execute a command. It calls to the client
     * to execute the command because the desired command might be implemented on the client (or otherwise not in
     * this extension host).
     */
    public executeCommand(command: string, args: any[]): Promise<any> {
        return this.proxy.$executeCommand(command, args)
    }

    public registerCommand(entry: CommandEntry): Unsubscribable {
        return syncSubscription(this.proxy.$registerCommand(entry.command, proxyValue(entry.callback)))
    }

    public registerActionEditCommand(entry: CommandEntry): Unsubscribable {
        return syncSubscription(
            this.proxy.$registerCommand(
                entry.command,
                proxyValue(async (diagnostic: Diagnostic | null, ...args: any[]) => {
                    const edit: WorkspaceEdit = await Promise.resolve(
                        entry.callback(diagnostic !== null ? toDiagnostic(diagnostic) : null, ...args)
                    )
                    return edit.toJSON()
                })
            )
        )
    }
}
