import { EditorId } from './editors'

const jetBrainsApis = ['protocolHandler', 'builtInServer']

export type EditorReplacements = Record<string, string>

export interface EditorSettings {
    editorId?: EditorId
    projectPaths?: {
        default?: string
        linux?: string
        mac?: string
        windows?: string
    }
    replacements?: EditorReplacements
    jetbrains?: {
        forceApi?: typeof jetBrainsApis[number]
    }
    vscode?: {
        isProjectPathUNCPath?: boolean
        useInsiders?: boolean
        useSSH?: boolean
        remoteHostForSSH?: string
    }
    custom?: {
        urlPattern?: string
    }
}
