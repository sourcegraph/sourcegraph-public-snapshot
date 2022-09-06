import { EditorSettings } from './editor-settings'
import { migrateLegacySettings } from './migrate-legacy-settings'

describe('migrate legacy editor settings tests', () => {
    it('migrates one legacy editor setting', () => {
        const editorId = 'vscode'
        const newSettings = migrateLegacySettings({ 'openineditor.editor': editorId })
        expect(newSettings).toHaveProperty(['openInEditor', 'editorIds'])
        expect((newSettings.openInEditor as EditorSettings).editorIds).toHaveLength(1)
        expect((newSettings.openInEditor as EditorSettings).editorIds?.[0]).toBe(editorId)
    })

    it('migrates all legacy editor settings', () => {
        const editorId = 'vscode'
        const customUrlPattern = 'vscode://file%file:%line:%col'
        const vscodeRemoteHost = 'user@test'
        const basePath = '/home/user/projects'
        const linuxBasePath = '/home/user/projectsLinux'
        const macBasePath = '/Users/user/projectsMac'
        const windowsBasePath = 'C:\\Users\\user\\projects'
        const replacements = { abc: 'def' }

        // noinspection SpellCheckingInspection
        const settings = {
            'openineditor.editor': editorId,
            'openineditor.customUrlPattern': customUrlPattern,
            'vscode.open.useMode': 'insiders',
            'vscode.open.remoteHost': vscodeRemoteHost,
            'openInIntellij.useBuiltin': true,
            'vscode.open.basePath': basePath,
            'vscode.open.osPaths.linux': linuxBasePath,
            'vscode.open.osPaths.mac': macBasePath,
            'vscode.open.osPaths.windows': windowsBasePath,
            'vscode.open.replacements': replacements,
        }

        const newSettings = migrateLegacySettings(settings)

        expect(newSettings).toHaveProperty('openInEditor')
        expect((newSettings.openInEditor as EditorSettings).editorIds?.[0]).toBe(editorId)
        expect(newSettings.openInEditor).toHaveProperty(['custom.urlPattern'], customUrlPattern)
        expect(newSettings.openInEditor).toHaveProperty(['vscode.useInsiders'], true)
        expect(newSettings.openInEditor).toHaveProperty(['vscode.remoteHostForSSH'], vscodeRemoteHost)
        expect(newSettings.openInEditor).toHaveProperty(['jetbrains.forceApi'], 'builtInServer')
        expect(newSettings.openInEditor).toHaveProperty(['projectPaths.default'], basePath)
        expect(newSettings.openInEditor).toHaveProperty(['projectPaths.linux'], linuxBasePath)
        expect(newSettings.openInEditor).toHaveProperty(['projectPaths.mac'], macBasePath)
        expect(newSettings.openInEditor).toHaveProperty(['projectPaths.windows'], windowsBasePath)
        expect(newSettings.openInEditor).toHaveProperty('replacements', replacements)
    })

    it('retains unrelated settings', () => {
        const someObject = { baz: 2 }
        const newSettings = migrateLegacySettings({ foo: 1, bar: someObject, 'openineditor.editor': 'vscode' })
        expect(newSettings).toHaveProperty('foo', 1)
        expect(newSettings).toHaveProperty('bar', someObject)
    })

    it('deletes legacy settings', () => {
        const newSettings = migrateLegacySettings({ 'openineditor.editor': 'vscode' })
        expect(newSettings).not.toHaveProperty('openineditor.editor')
    })

    it('doesn’t change the original settings object', () => {
        const originalSettings = { test: 1, 'openineditor.editor': 'vscode' }
        migrateLegacySettings(originalSettings)
        expect(Object.keys(originalSettings)).toHaveLength(2)
        expect(originalSettings).toHaveProperty('test')
        expect(originalSettings).toHaveProperty(['openineditor.editor'], 'vscode')
        expect(originalSettings).not.toHaveProperty('openInEditor')
    })

    it('uses the right precedence', () => {
        const basePath = '/home/user/projects'
        const newSettings = migrateLegacySettings({
            'vscode.open.basePath': '/home/user/projectsVsCode',
            'openineditor.basePath': basePath,
            'openInIntellij.basePath': '/home/user/projectsVsCode',
            'openInAtom.basePath': '/home/user/projectsVsCode',
        })
        expect(newSettings).toHaveProperty(['openInEditor', 'projectPaths.default'], basePath)
    })

    it('doesn’t do anything if there are new settings available', () => {
        const originalSettings = { test: 1, 'openineditor.editor': 'vscode', openInEditor: {} }
        const newSettings = migrateLegacySettings(originalSettings)
        expect(newSettings).toBe(originalSettings)
    })
})
