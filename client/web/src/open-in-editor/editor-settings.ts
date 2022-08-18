import { EditorId } from './editors'

const jetBrainsApis = ['protocolHandler', 'builtInServer']
const vsCodeModes = ['standard', 'insiders', 'ssh']

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
        mode?: typeof vsCodeModes[number]
        remoteHostForSSH?: string
    }
    custom?: {
        urlPattern?: string
    }
}
