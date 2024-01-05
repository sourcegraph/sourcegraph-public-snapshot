import { describe, expect, it } from 'vitest'

import { ExternalServiceKind } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import { buildEditorUrl, buildRepoBaseNameAndPath } from './build-url'
import type { EditorSettings } from './editor-settings'

function buildSettings(props: EditorSettings = {}): EditorSettings {
    return {
        editorIds: ['vscode'],
        'projectPaths.default': '/home/user/projects',
        ...props,
    }
}

describe('buildRepoBaseNameAndPath tests', () => {
    it('builds the correct string for "repositoryPathPattern": "{nameWithOwner}" config', () => {
        const url = 'https://sourcegraph.com/sourcegraph/sourcegraph/-/blob/tsconfig.json'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.GITHUB, filePath)

        expect(result).toEqual('sourcegraph/tsconfig.json')
    })

    it('builds the correct string for GitHub URLs', () => {
        const url = 'https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/tsconfig.json'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.GITHUB, filePath)

        expect(result).toEqual('sourcegraph/tsconfig.json')
    })

    it('builds the correct string for GitHub Enterprise URLs', () => {
        const url = 'https://k8s.sgdev.org/ghe.sgdev.org/sourcegraph/idan-test/-/blob/README.md'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.GITHUB, filePath)

        expect(result).toEqual('idan-test/README.md')
    })

    it('builds the correct string for GitLab URLs', () => {
        const url = 'https://sourcegraph.com/gitlab.com/gitlab-org/gitlab-foss/-/blob/.eslintignore'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.GITLAB, filePath)

        expect(result).toEqual('gitlab-foss/.eslintignore')
    })

    it('builds the correct string for self-hosted GitLab URLs', () => {
        const url =
            'https://k8s.sgdev.org/gitlab.sgdev.org/sg-repos/public-sg-repos/gophercon-2018-liveblog/-/blob/README.md'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.GITLAB, filePath)

        expect(result).toEqual('public-sg-repos/gophercon-2018-liveblog/README.md')
    })

    it('builds the correct string for Bitbucket Cloud URLs', () => {
        const url = 'https://sourcegraph.com/bitbucket.org/atlassian/stash-example-plugin/src/master/README.md'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, 'bitbucketCloud', filePath)

        expect(result).toEqual('stash-example-plugin/src/master/README.md')
    })

    it('builds the correct string for Perforce URLs', () => {
        const url =
            'https://cse-k8s.sgdev.org/perforce.beatrix.com/app/b200/patch/core/-/blob/test/1.js?toast=integrations'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.PERFORCE, filePath)

        expect(result).toEqual('app/b200/patch/core/test/1.js')
    })

    it('builds the correct string for other URLs', () => {
        const url = 'https://sourcegraph.com/maven/com.esotericsoftware.minlog/minlog/-/blob/lsif-java.json'
        const { repoName, filePath } = parseBrowserRepoURL(url)

        const result = buildRepoBaseNameAndPath(repoName, ExternalServiceKind.OTHER, filePath)

        expect(result).toEqual('com.esotericsoftware.minlog/minlog/lsif-java.json')
    })
})

describe('buildEditorUrl tests', () => {
    const defaultPosition = { line: 43, character: 0 }
    const defaultPath = 'sourcegraph/.gitignore'
    const baseUrl = 'https://sourcegraph.com'
    describe('happy paths', () => {
        it('builds the correct URL for some basic settings and VS Code', () => {
            const url = buildEditorUrl(defaultPath, defaultPosition, buildSettings(), baseUrl)
            expect(url.toString()).toBe('vscode://file/home/user/projects/sourcegraph/.gitignore:43:0')
        })

        it('builds the correct URL for some basic settings and IDEA', () => {
            const url = buildEditorUrl(defaultPath, defaultPosition, buildSettings({ editorIds: ['idea'] }), baseUrl)
            expect(url.toString()).toBe('idea://open?file=/home/user/projects/sourcegraph/.gitignore&line=43&column=0')
        })

        it('builds the correct URL for some basic settings and Atom', () => {
            const url = buildEditorUrl(defaultPath, defaultPosition, buildSettings({ editorIds: ['atom'] }), baseUrl)
            expect(url.toString()).toBe(
                'atom://core/open/file?filename=/home/user/projects/sourcegraph/.gitignore:43:0'
            )
        })

        it('builds the correct URL for some basic settings and Sublime', () => {
            const url = buildEditorUrl(defaultPath, defaultPosition, buildSettings({ editorIds: ['sublime'] }), baseUrl)
            expect(url.toString()).toBe('subl://open?url=/home/user/projects/sourcegraph/.gitignore&line=43&column=0')
        })

        it('builds the correct URL for some basic settings and PyCharm', () => {
            const url = buildEditorUrl(defaultPath, defaultPosition, buildSettings({ editorIds: ['pycharm'] }), baseUrl)
            expect(url.toString()).toBe(
                'pycharm://open?file=/home/user/projects/sourcegraph/.gitignore&line=43&column=0'
            )
        })

        it('rewrites default project path with OS specific one', () => {
            const oldUserAgent = navigator.userAgent
            Object.defineProperty(navigator, 'userAgent', { value: 'MacOS', writable: true })
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    editorIds: ['goland'],
                    'projectPaths.default': '/home/user/projects',
                    'projectPaths.mac': '/Users/user/projects',
                }),
                baseUrl
            )
            expect(url.toString()).toBe(
                'goland://open?file=/Users/user/projects/sourcegraph/.gitignore&line=43&column=0'
            )
            Object.defineProperty(navigator, 'userAgent', { value: oldUserAgent, writable: true })
        })

        it('performs replacements', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({ replacements: { projects: 'new-projects' } }),
                baseUrl
            )
            expect(url.toString()).toBe('vscode://file/home/user/new-projects/sourcegraph/.gitignore:43:0')
        })

        it('forces JetBrains built-in server', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    editorIds: ['goland'],
                    'jetbrains.forceApi': 'builtInServer',
                }),
                baseUrl
            )
            expect(url.toString()).toBe(
                'http://localhost:63342/api/file/home/user/projects/sourcegraph/.gitignore:43:0'
            )
        })

        it('handles UNC paths for VS Code', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    'projectPaths.default': '/server/projects',
                    'vscode.isProjectPathUNCPath': true,
                }),
                baseUrl
            )
            expect(url.toString()).toBe('vscode://file//server/projects/sourcegraph/.gitignore:43:0')
        })

        it('handles Windows paths for VS Code', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({ 'projectPaths.default': 'C:\\Projects' }),
                baseUrl
            )
            expect(url.toString()).toBe('vscode://file/C:\\Projects/sourcegraph/.gitignore:43:0')
        })

        it('handles no range', () => {
            const url = buildEditorUrl(defaultPath, undefined, buildSettings(), baseUrl)
            expect(url.toString()).toBe('vscode://file/home/user/projects/sourcegraph/.gitignore:1:1')
        })

        it('can use insiders build of VS Code', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({ 'vscode.useInsiders': true }),
                baseUrl
            )
            expect(url.toString()).toBe('vscode-insiders://file/home/user/projects/sourcegraph/.gitignore:43:0')
        })

        it('can use SSH with VS Code', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    'vscode.useSSH': true,
                    'vscode.remoteHostForSSH': '127.0.0.1',
                }),
                baseUrl
            )
            expect(url.toString()).toBe(
                'vscode://vscode-remote/ssh-remote+127.0.0.1/home/user/projects/sourcegraph/.gitignore:43:0'
            )
        })

        it('can use SSH with VS Code Insiders', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    'vscode.useInsiders': true,
                    'vscode.useSSH': true,
                    'vscode.remoteHostForSSH': '127.0.0.1',
                }),
                baseUrl
            )
            expect(url.toString()).toBe(
                'vscode-insiders://vscode-remote/ssh-remote+127.0.0.1/home/user/projects/sourcegraph/.gitignore:43:0'
            )
        })

        it('can use a custom URL pattern', () => {
            const url = buildEditorUrl(
                defaultPath,
                defaultPosition,
                buildSettings({
                    editorIds: ['custom'],
                    'custom.urlPattern': 'idea://test?file=%file&line=%line&column=%col',
                }),
                baseUrl
            )
            expect(url.toString()).toBe('idea://test?file=/home/user/projects/sourcegraph/.gitignore&line=43&column=0')
        })
    })

    describe('unhappy paths', () => {
        it('recognizes missing editor settings', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, undefined, baseUrl)
            }).toThrow()
        })

        it('recognizes missing project path', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, { editorIds: ['vscode'] }, baseUrl)
            }).toThrow()
        })

        it('recognizes non-absolute project path', () => {
            expect(() => {
                buildEditorUrl(
                    defaultPath,
                    defaultPosition,
                    buildSettings({ 'projectPaths.default': '../projects' }),
                    baseUrl
                )
            }).toThrow()
        })

        it('recognizes missing editor ID', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, { 'projectPaths.default': '/home/user/projects' }, baseUrl)
            }).toThrow()
        })

        it('recognizes missing customUrlPattern in case of custom editor setting', () => {
            expect(() => {
                buildEditorUrl(
                    defaultPath,
                    defaultPosition,
                    buildSettings({
                        editorIds: ['custom'],
                        'projectPaths.default': '/home/user/projects',
                    }),
                    baseUrl
                )
            }).toThrow()
        })

        it('recognizes missing editor settings', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, { editorIds: ['vscode'] }, baseUrl)
            }).toThrow()
        })

        it('recognizes missing SSH remote setting if vscode SSH mode is enabled', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, buildSettings({ 'vscode.useSSH': true }), baseUrl)
            }).toThrow()
        })

        it('builds the right "Learn more" URL', () => {
            expect(() => {
                buildEditorUrl(defaultPath, defaultPosition, { editorIds: ['vscode'] }, baseUrl)
            }).toThrow(/https:\/\/docs\.sourcegraph\.com\/integration\/open_in_editor/)
        })
    })
})
