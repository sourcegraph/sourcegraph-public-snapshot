import assert from 'assert'

import { FileTree } from './FileTree'
import { SourcegraphUri } from './SourcegraphUri'

const tree = new FileTree(SourcegraphUri.parse('https://sourcegraph.com/sourcegraph-vscode@v8'), [
    '.eslintrc.json',
    '.github/workflows/build.yml',
    '.gitignore',
    '.vscode/extensions.json',
    '.vscode/launch.json',
    '.vscode/settings.json',
    '.vscode/tasks.json',
    '.vscodeignore',
    'README.md',
    'images/logo.png',
    'renovate.json',
    'src/browse/BrowseFileSystemProvider.ts',
    'src/browse/browseCommand.ts',
    'src/browse/graphqlQuery.ts',
    'src/browse/parseRepoUri.test.ts',
    'src/browse/parseRepoUrl.ts',
    'src/config.ts',
    'src/extension.ts',
    'src/git/helpers.ts',
    'src/git/index.ts',
    'src/git/remoteNameAndBranch.test.ts',
    'src/git/remoteNameAndBranch.ts',
    'src/git/remoteUrl.test.ts',
    'src/git/remoteUrl.ts',
    'src/log.ts',
    'tests/config.json',
    'tsconfig.json',
])

function checkChildren(directory: string, expected: string[]) {
    it(`directChildren('${directory}')`, () => {
        const childUris = tree.directChildren(directory)
        const obtained: string[] = []
        for (const childUri of childUris) {
            const uri = SourcegraphUri.parse(childUri)
            if (uri.path) {
                const path = uri.isDirectory() ? uri.path + '/' : uri.path
                obtained.push(path)
            }
        }
        assert.deepStrictEqual(obtained, expected)
    })
}
describe('FileTree', () => {
    checkChildren('src', ['src/browse/', 'src/git/', 'src/config.ts', 'src/extension.ts', 'src/log.ts'])
    checkChildren('', [
        '.github/workflows/',
        '.vscode/',
        'images/',
        'src/',
        'tests/',
        '.eslintrc.json',
        '.gitignore',
        '.vscodeignore',
        'README.md',
        'renovate.json',
        'tsconfig.json',
    ])
})
