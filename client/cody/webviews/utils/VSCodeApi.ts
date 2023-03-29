declare const acquireVsCodeApi: () => VSCodeApi

interface VSCodeApi {
    getState: () => any
    setState: (newState: any) => any
    postMessage: (message: any) => void
}

class VSCodeWrapper {
    private readonly vscodeApi: VSCodeApi = acquireVsCodeApi()

    public postMessage(message: WebviewMessage): void {
        this.vscodeApi.postMessage(message)
    }

    public onMessage(callback: (message: any) => void): () => void {
        window.addEventListener('message', callback)
        return () => window.removeEventListener('message', callback)
    }
}

export const vscodeAPI: VSCodeWrapper = new VSCodeWrapper()

interface IntializedWebviewMessage {
    command: 'initialized'
}

interface ResetWebviewMessage {
    command: 'reset'
}

interface RemoveTokenWebviewMessage {
    command: 'removeToken'
}

interface SettingsWebviewMessage {
    command: 'settings'
    serverEndpoint: string
    accessToken: string
}

interface SubmitWebviewMessage {
    command: 'submit'
    text: string
}

interface ExecuteRecipeWebviewMessage {
    command: 'executeRecipe'
    recipe: string
}

interface RemoveChatHistoryWebviewMessage {
    command: 'removeHistory'
}

interface OpenFile {
    command: 'openFile'
    filePath: string
}

type WebviewMessage =
    | IntializedWebviewMessage
    | ResetWebviewMessage
    | RemoveTokenWebviewMessage
    | SettingsWebviewMessage
    | SubmitWebviewMessage
    | ExecuteRecipeWebviewMessage
    | RemoveChatHistoryWebviewMessage
    | OpenFile
