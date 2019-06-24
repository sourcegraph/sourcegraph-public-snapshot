import { getFixtureBody } from '../code_intelligence/code_intelligence_test_utils'
import { getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView } from './scrape'

describe('Bitbucket scrape.ts', () => {
    describe('getDiffFileInfoFromMultiFileDiffCodeView()', () => {
        it('should get the FileInfo for an added file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/new_file.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/added.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: undefined,
                baseRawRepoName: undefined,
                filePath: 'dir/new_file.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a modified file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/modified.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux.go',
                baseRawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a deleted file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/old_test.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/deleted.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/old_test.go',
                baseRawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/old_test.go', // TODO should really be undefined?
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a copied file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux.1.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/copied.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux.go',
                baseRawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux.1.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a renamed file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux_test_moved.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/renamed.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux_test.go',
                baseRawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux_test_moved.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a moved file', async () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/test-dir/route.go',
            })
            const codeView = await getFixtureBody({
                htmlFixturePath: `${__dirname}/__fixtures__/code-views/pull-request/split/moved.html`,
                isFullDocument: false,
            })
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/route.go',
                baseRawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/test-dir/route.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                rawRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
    })
})
