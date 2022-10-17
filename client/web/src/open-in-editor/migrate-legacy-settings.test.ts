import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { EditorSettings } from './editor-settings'
import { migrateLegacySettings } from './migrate-legacy-settings'

describe('migrate legacy editor settings tests', () => {
    it('migrates one legacy editor setting', () => {
        const editorId = 'vscode'
        const newSettings = migrateLegacySettings({ 'openineditor.editor': editorId })
        expect(newSettings).toHaveProperty(['editorIds'])
        expect((newSettings as EditorSettings).editorIds).toHaveLength(1)
        expect((newSettings as EditorSettings).editorIds?.[0]).toBe(editorId)
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
        const settings: Settings = {
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

        const openInEditor = migrateLegacySettings(settings)

        expect((openInEditor as EditorSettings).editorIds?.[0]).toBe(editorId)
        expect(openInEditor).toHaveProperty(['custom.urlPattern'], customUrlPattern)
        expect(openInEditor).toHaveProperty(['vscode.useInsiders'], true)
        expect(openInEditor).toHaveProperty(['vscode.remoteHostForSSH'], vscodeRemoteHost)
        expect(openInEditor).toHaveProperty(['jetbrains.forceApi'], 'builtInServer')
        expect(openInEditor).toHaveProperty(['projectPaths.default'], basePath)
        expect(openInEditor).toHaveProperty(['projectPaths.linux'], linuxBasePath)
        expect(openInEditor).toHaveProperty(['projectPaths.mac'], macBasePath)
        expect(openInEditor).toHaveProperty(['projectPaths.windows'], windowsBasePath)
        expect(openInEditor).toHaveProperty('replacements', replacements)
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
        const openInEditor = migrateLegacySettings({
            'vscode.open.basePath': '/home/user/projectsVsCode',
            'openineditor.basePath': basePath,
            'openInIntellij.basePath': '/home/user/projectsVsCode',
            'openInAtom.basePath': '/home/user/projectsVsCode',
        })
        expect(openInEditor).toHaveProperty(['projectPaths.default'], basePath)
    })

    it('doesn’t do anything if there are new settings available', () => {
        const originalSettings = { test: 1, 'openineditor.editor': 'vscode', openInEditor: {} }
        const newSettings = migrateLegacySettings(originalSettings)
        expect(newSettings).toBe(null)
    })
})
