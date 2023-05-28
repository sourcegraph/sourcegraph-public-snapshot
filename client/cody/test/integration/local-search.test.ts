import * as assert from 'assert'
import path from 'path'

import * as vscode from 'vscode'

import { fastFilesExist } from '../../src/chat/fastFileFinder'
import { getRgPath } from '../../src/rg'

import { afterIntegrationTest, beforeIntegrationTest } from './helpers'

suite('Local search', function () {
    this.beforeEach(() => beforeIntegrationTest())
    this.afterEach(() => afterIntegrationTest())

    test('fast file finder', async () => {
        const workspaceFolders = vscode.workspace.workspaceFolders
        assert.ok(workspaceFolders)
        assert.ok(workspaceFolders.length >= 1)

        const rgPath = await getRgPath(path.join(__dirname, '..', '..', '..'))
        const filesExistMap = await fastFilesExist(rgPath, workspaceFolders[0].uri.fsPath, [
            'lib',
            'batches',
            'env',
            'var.go',
            'lib/batches',
            'batches/env',
            'lib/batches/env/var.go',
            'lib/batches/var.go',
            './lib/codeintel/tools/lsif-visualize/visualize.go',
        ])
        assert.deepStrictEqual(filesExistMap, {
            lib: true,
            batches: true,
            env: true,
            'var.go': true,
            'lib/batches': true,
            'batches/env': true,
            'lib/batches/env/var.go': true,
            'lib/batches/var.go': false,
            './lib/codeintel/tools/lsif-visualize/visualize.go': true,
        })
    })
})
