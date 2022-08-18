export interface Editor {
    id: string
    name: string
    urlPattern: string
    isJetBrainsProduct: boolean
}

export const supportedEditors: Editor[] = [
    {
        id: 'appcode',
        name: 'AppCode',
        urlPattern: 'appcode://open?file=%file&line=%line&column=%col',
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
        urlPattern: 'clion://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'goland',
        name: 'GoLand',
        urlPattern: 'goland://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'idea',
        name: 'IntelliJ IDEA',
        urlPattern: 'idea://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'phpstorm',
        name: 'PhpStorm',
        urlPattern: 'phpstorm://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'pycharm',
        name: 'PyCharm',
        urlPattern: 'pycharm://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'rider',
        name: 'Rider',
        urlPattern: 'rider://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    {
        id: 'rubymine',
        name: 'RubyMine',
        urlPattern: 'rubymine://open?file=%file&line=%line&column=%col',
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
        urlPattern: 'vscode://file%file:%line:%col',
        isJetBrainsProduct: false,
    },
    {
        id: 'webstorm',
        name: 'WebStorm',
        urlPattern: 'webstorm://open?file=%file&line=%line&column=%col',
        isJetBrainsProduct: true,
    },
    { id: 'custom', name: 'Custom', urlPattern: '', isJetBrainsProduct: false },
]

export type EditorId = typeof supportedEditors[number]['id']

export function getEditor(editorId: EditorId): Editor | undefined {
    return supportedEditors.find(editor => editor.id === editorId)
}
