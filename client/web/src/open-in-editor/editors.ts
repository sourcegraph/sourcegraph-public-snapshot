export interface Editor {
    id: string
    name: string
    urlPattern?: string
    isJetBrainsProduct: boolean
}

export const supportedEditors: Editor[] = [
    {
        id: 'appcode',
        name: 'AppCode',
        isJetBrainsProduct: true,
    },
    {
        id: 'atom',
        name: 'Atom',
        urlPattern: 'atom://core/open/file?filename=%file:%line:%col',
        isJetBrainsProduct: false,
    },
    {
        id: 'clion',
        name: 'CLion',
        isJetBrainsProduct: true,
    },
    {
        id: 'goland',
        name: 'GoLand',
        isJetBrainsProduct: true,
    },
    {
        id: 'idea',
        name: 'IntelliJ IDEA',
        isJetBrainsProduct: true,
    },
    {
        id: 'phpstorm',
        name: 'PhpStorm',
        isJetBrainsProduct: true,
    },
    {
        id: 'pycharm',
        name: 'PyCharm',
        isJetBrainsProduct: true,
    },
    {
        id: 'rider',
        name: 'Rider',
        isJetBrainsProduct: true,
    },
    {
        id: 'rubymine',
        name: 'RubyMine',
        isJetBrainsProduct: true,
    },
    {
        id: 'sublime',
        name: 'Sublime Text',
        urlPattern: 'subl://open?url=%file&line=%line&column=%col',
        isJetBrainsProduct: false,
    },
    {
        id: 'vscode',
        name: 'Visual Studio Code',
        isJetBrainsProduct: false,
    },
    {
        id: 'webstorm',
        name: 'WebStorm',
        isJetBrainsProduct: true,
    },
    { id: 'custom', name: 'Custom', urlPattern: '', isJetBrainsProduct: false },
]

export type EditorId = typeof supportedEditors[number]['id']

export function getEditor(editorId: EditorId): Editor | undefined {
    return supportedEditors.find(editor => editor.id === editorId)
}
