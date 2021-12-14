import assert from 'assert'

import { SourcegraphUri } from './SourcegraphUri'

function check(
    input: string,
    expected: { repositoryName: string; revision: string; path?: string },
    assertion?: (uri: SourcegraphUri) => void
) {
    it(`parseBrowserRepoURL('${input})'`, () => {
        const obtained = SourcegraphUri.parse(input)
        assert.deepStrictEqual(obtained.repositoryName, expected.repositoryName, 'repositoryName does not match')
        assert.deepStrictEqual(obtained.revision, expected.revision, 'revision does not match')
        assert.deepStrictEqual(obtained.path, expected.path, 'path does not match')
        if (assertion) {
            assertion(obtained)
        }
        const roundtrip = SourcegraphUri.parse(obtained.uri)
        assert.deepStrictEqual(obtained, roundtrip, 'roundtrip test failed, uri !== Sourcegraph.parse(uri.uri)')
    })
}

function checkParent(input: string, expected: string | undefined) {
    it(`checkParent('${input}')`, () => {
        const obtained = SourcegraphUri.parse(input).parentUri()
        assert.deepStrictEqual(obtained, expected)
    })
}

describe('SourcegraphUri', () => {
    check('sourcegraph://sourcegraph.com/jdk@v8/-/blob/java/lang/String.java', {
        repositoryName: 'jdk',
        revision: 'v8',
        path: 'java/lang/String.java',
    })
    checkParent(
        'sourcegraph://sourcegraph.com/jdk@v8/-/blob/java/lang/String.java',
        'sourcegraph://sourcegraph.com/jdk@v8/-/tree/java/lang'
    )
    checkParent(
        'sourcegraph://sourcegraph.com/github.com/sourcegraph@v8/-/blob/indexing/dependency_indexing_scheduler_test.go',
        'sourcegraph://sourcegraph.com/github.com/sourcegraph@v8/-/tree/indexing'
    )
    checkParent(
        'sourcegraph://sourcegraph.com/github.com/sourcegraph/-/blob/indexing/dependency_indexing_scheduler_test.go#L102:1',
        'sourcegraph://sourcegraph.com/github.com/sourcegraph/-/tree/indexing'
    )
    checkParent(
        'sourcegraph://sourcegraph.com/jdk@v8/-/tree/java/lang',
        'sourcegraph://sourcegraph.com/jdk@v8/-/tree/java'
    )
    checkParent('sourcegraph://sourcegraph.com/jdk@v8/-/tree/java', 'sourcegraph://sourcegraph.com/jdk@v8')
    checkParent('sourcegraph://sourcegraph.com/jdk@v8', undefined)
    check(
        'sourcegraph://sourcegraph.com/REPO/-/commit/COMMIT?visible=1',
        {
            repositoryName: 'REPO',
            revision: 'COMMIT',
            path: '',
        },
        uri => {
            assert.strictEqual(uri.isCommit(), true)
        }
    )
    check(
        'sourcegraph://sourcegraph.com/REPO/-/commit/COMMIT',
        { repositoryName: 'REPO', revision: 'COMMIT', path: '' },
        uri => assert.strictEqual(uri.isCommit(), true)
    )
    check(
        'sourcegraph://sourcegraph.com/REPO/-/compare/COMMIT1...COMMIT2',
        { repositoryName: 'REPO', revision: '', path: undefined },
        uri => {
            assert(uri.isCompare())
            assert.deepStrictEqual(uri.compareRange, { base: 'COMMIT1', head: 'COMMIT2' })
        }
    )
    check(
        'sourcegraph://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/compare/main...code-intel/cesar/22229/keyboard-shortcut-fuzzy-finder?visible=6',
        { repositoryName: 'github.com/sourcegraph/sourcegraph', revision: '', path: undefined },
        uri => {
            assert.deepStrictEqual(uri.compareRange, {
                base: 'main',
                head: 'code-intel/cesar/22229/keyboard-shortcut-fuzzy-finder',
            })
        }
    )
})
