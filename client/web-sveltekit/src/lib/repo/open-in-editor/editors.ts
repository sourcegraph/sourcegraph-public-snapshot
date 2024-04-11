export interface Editor {
    id: string
    telemetryID: number
    name: string
    urlPattern?: string
    isJetBrainsProduct: boolean
}

export const supportedEditors: Editor[] = [
    {
        id: 'appcode',
        name: 'AppCode',
        isJetBrainsProduct: true,
        telemetryID: 1,
    },
    {
        id: 'atom',
        name: 'Atom',
        urlPattern: 'atom://core/open/file?filename=%file:%line:%col',
        isJetBrainsProduct: false,
        telemetryID: 2,
    },
    {
        id: 'clion',
        name: 'CLion',
        isJetBrainsProduct: true,
        telemetryID: 3,
    },
    {
        id: 'goland',
        name: 'GoLand',
        isJetBrainsProduct: true,
        telemetryID: 4,
    },
    {
        id: 'idea',
        name: 'IntelliJ IDEA',
        isJetBrainsProduct: true,
        telemetryID: 5,
    },
    {
        id: 'phpstorm',
        name: 'PhpStorm',
        isJetBrainsProduct: true,
        telemetryID: 6,
    },
    {
        id: 'pycharm',
        name: 'PyCharm',
        isJetBrainsProduct: true,
        telemetryID: 7,
    },
    {
        id: 'rider',
        name: 'Rider',
        isJetBrainsProduct: true,
        telemetryID: 8,
    },
    {
        id: 'rubymine',
        name: 'RubyMine',
        isJetBrainsProduct: true,
        telemetryID: 9,
    },
    {
        id: 'sublime',
        name: 'Sublime Text',
        urlPattern: 'subl://open?url=%file&line=%line&column=%col',
        isJetBrainsProduct: false,
        telemetryID: 10,
    },
    {
        id: 'vscode',
        name: 'Visual Studio Code',
        isJetBrainsProduct: false,
        telemetryID: 11,
    },
    {
        id: 'webstorm',
        name: 'WebStorm',
        isJetBrainsProduct: true,
        telemetryID: 12,
    },
    { id: 'custom', name: 'Custom', urlPattern: '', isJetBrainsProduct: false, telemetryID: 13 },
]

export type EditorId = typeof supportedEditors[number]['id']

export function getEditor(editorId: EditorId): Editor | undefined {
    return supportedEditors.find(editor => editor.id === editorId)
}
