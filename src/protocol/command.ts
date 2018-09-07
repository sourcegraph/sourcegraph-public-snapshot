import { RequestType } from '../jsonrpc2/messages'

export interface CommandClientCapabilities {
    workspace?: {
        executeCommand?: {
            dynamicRegistration?: boolean
        }
    }
}

/**
 * Execute command options.
 */
export interface ExecuteCommandOptions {
    /**
     * The commands to be executed on the server
     */
    commands: string[]
}

export interface ExecuteCommandParams {
    /**
     * The identifier of the actual command handler.
     */
    command: string
    /**
     * Arguments that the command should be invoked with.
     */
    arguments?: any[]
}

/**
 * Execute command registration options.
 */
export interface ExecuteCommandRegistrationOptions extends ExecuteCommandOptions {}

/**
 * A request send from the client to the server to execute a command. The request might return
 * a workspace edit which the client will apply to the workspace.
 */
export namespace ExecuteCommandRequest {
    export const type = new RequestType<ExecuteCommandParams, any | null, void, ExecuteCommandRegistrationOptions>(
        'workspace/executeCommand'
    )
}
