import { EditorId } from './editors'

export const jetBrainsApis = ['protocolHandler', 'builtInServer']

export type Replacements = Record<string, string>

export interface EditorSettings {
    editorId?: EditorId
    projectsPaths?: {
        default?: string
        linux?: string
        mac?: string
        windows?: string
    }
    replacements?: Replacements
    jetbrains?: {
        forceApi?: typeof jetBrainsApis[number]
    }
    vscode?: {
        isBasePathUNCPath?: boolean
        useInsiders?: boolean
        useSSH?: boolean
        remoteHostForSSH?: string
    }
    custom?: {
        urlPattern?: string
    }
}
