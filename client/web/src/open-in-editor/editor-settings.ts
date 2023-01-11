import { EditorId } from './editors'

const jetBrainsApis = ['protocolHandler', 'builtInServer']

export type EditorReplacements = Record<string, string>

export interface EditorSettings {
    editorIds?: EditorId[]
    'projectPaths.default'?: string
    'projectPaths.linux'?: string
    'projectPaths.mac'?: string
    'projectPaths.windows'?: string
    replacements?: EditorReplacements
    'jetbrains.forceApi'?: typeof jetBrainsApis[number]
    'vscode.isProjectPathUNCPath'?: boolean
    'vscode.useInsiders'?: boolean
    'vscode.useSSH'?: boolean
    'vscode.remoteHostForSSH'?: string
    'custom.urlPattern'?: string
}
