import * as lsp from 'vscode-languageserver-protocol'
import { internalLocationToLocation } from './lsif'

describe('internalLocationToLocation', () => {
    const dump = {
        id: 0,
        repository: 'github.com/sourcegraph/codeintellify',
        commit: 'deadbeef',
        root: '',
        visibleAtTip: false,
        uploadedAt: new Date(),
        processedAt: new Date(),
    }

    const range: lsp.Range = {
        start: {
            line: 1,
            character: 1,
        },
        end: {
            line: 2,
            character: 3,
        },
    }

    it('should generate a relative URI to the same repo', () => {
        const input = { dump, path: 'src/position.ts', range }
        const location = internalLocationToLocation('github.com/sourcegraph/codeintellify', input)
        const expected = lsp.Location.create('src/position.ts', range)
        expect(location).toEqual(expected)
    })

    it('should generate an absolute URI to another project', () => {
        const input = { dump, path: 'src/position.ts', range }
        const location = internalLocationToLocation('github.com/sourcegraph/lsif-go', input)
        const expected = lsp.Location.create(
            'git://github.com/sourcegraph/codeintellify?deadbeef#src/position.ts',
            range
        )
        expect(location).toEqual(expected)
    })
})
