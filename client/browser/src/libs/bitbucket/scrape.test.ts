import { getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView } from './scrape'
import * as testCodeViews from './test-code-views'

describe('Bitbucket scrape.ts', () => {
    describe('getDiffFileInfoFromMultiFileDiffCodeView()', () => {
        it('should get the FileInfo for an added file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/new_file.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.added
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: undefined,
                baseRepoName: undefined,
                filePath: 'dir/new_file.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a modified file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.modified
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux.go',
                baseRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a deleted file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/old_test.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.deleted
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/old_test.go',
                baseRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/old_test.go', // TODO should really be undefined?
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a copied file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux.1.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.copied
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux.go',
                baseRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux.1.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a renamed file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/mux_test_moved.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.renamed
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/mux_test.go',
                baseRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/mux_test_moved.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
        it('should get the FileInfo for a moved file', () => {
            jsdom.reconfigure({
                url: 'https://bitbucket.test/projects/SOURCEGRAPH/repos/mux/pull-requests/1/diff#dir/test-dir/route.go',
            })
            const codeView = document.createElement('div')
            codeView.innerHTML = testCodeViews.pr.moved
            const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
            expect(fileInfo).toStrictEqual({
                baseFilePath: 'dir/route.go',
                baseRepoName: 'bitbucket.test/SOURCEGRAPH/mux',
                filePath: 'dir/test-dir/route.go',
                project: 'SOURCEGRAPH',
                repoSlug: 'mux',
                repoName: 'bitbucket.test/SOURCEGRAPH/mux',
            })
        })
    })
})
