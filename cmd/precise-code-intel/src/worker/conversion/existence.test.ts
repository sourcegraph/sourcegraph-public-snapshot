import * as sinon from 'sinon'
import { PathExistenceChecker } from './existence'
import { getDirectoryChildren } from '../../shared/gitserver/gitserver'
import { range } from 'lodash'

describe('PathExistenceChecker', () => {
    it('should test path existence in git tree', async () => {
        const children = new Map([
            ['', ['web', 'shared']],
            ['web', ['web/foo.ts']],
            ['web/shared', ['web/shared/bonk.ts']],
            ['shared', ['shared/bar.ts', 'shared/baz.ts']],
        ])

        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: 42,
            commit: 'c',
            root: 'web',
            frontendUrl: 'frontend',
            mockGetDirectoryChildren: ({ dirnames }) =>
                Promise.resolve(new Map(dirnames.map(dirname => [dirname, new Set(children.get(dirname))]))),
        })

        await pathExistenceChecker.warmCache([
            'foo.ts',
            'bar.ts',
            'shared/bonk.ts',
            '../shared/bar.ts',
            '../shared/bar.ts',
            '../shared/bonk.ts',
            '../node_modules/@types/quux.ts',
            '../../node_modules/@types/oops.ts',
        ])

        // Test within root
        expect(pathExistenceChecker.shouldIncludePath('foo.ts', false)).toBeTruthy()
        expect(pathExistenceChecker.shouldIncludePath('bar.ts', false)).toBeFalsy()
        expect(pathExistenceChecker.shouldIncludePath('shared/bonk.ts', false)).toBeTruthy()
        // Test outside root but within repo
        expect(pathExistenceChecker.shouldIncludePath('../shared/bar.ts', false)).toBeTruthy()
        expect(pathExistenceChecker.shouldIncludePath('../shared/bar.ts', true)).toBeFalsy()
        expect(pathExistenceChecker.shouldIncludePath('../shared/bonk.ts', false)).toBeFalsy()
        expect(pathExistenceChecker.shouldIncludePath('../node_modules/@types/quux.ts', false)).toBeFalsy()

        // Test outside repo
        expect(pathExistenceChecker.shouldIncludePath('../../node_modules/@types/oops.ts', false)).toBeFalsy()
    })

    it('should cache directory contents', async () => {
        const children = new Map([['', Array.from(range(100).map(i => `${i}.ts`))]])
        const mockGetDirectoryChildren = sinon.spy<typeof getDirectoryChildren>(({ dirnames }) =>
            Promise.resolve(new Map(dirnames.map(dirname => [dirname, new Set(children.get(dirname))])))
        )
        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: 42,
            commit: 'c',
            root: '',
            frontendUrl: 'frontend',
            mockGetDirectoryChildren,
        })

        await pathExistenceChecker.warmCache(Array.from(range(100).flatMap(i => [`${i}.ts`, `${i}.js`])))

        for (let i = 0; i < 100; i++) {
            expect(pathExistenceChecker.shouldIncludePath(`${i}.ts`, false)).toBeTruthy()
            expect(pathExistenceChecker.shouldIncludePath(`${i}.js`, false)).toBeFalsy()
        }

        expect(mockGetDirectoryChildren.callCount).toEqual(1)
    })

    it('should early out on untracked ancestors', async () => {
        const children = new Map([['', ['not_node_modules']]])
        const mockGetDirectoryChildren = sinon.spy<typeof getDirectoryChildren>(({ dirnames }) =>
            Promise.resolve(new Map(dirnames.map(dirname => [dirname, new Set(children.get(dirname))])))
        )

        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: 42,
            commit: 'c',
            root: '',
            frontendUrl: 'frontend',
            mockGetDirectoryChildren,
        })

        await pathExistenceChecker.warmCache(
            range(0, 100).flatMap(i => [`node_modules/${i}/deeply/nested/lib/file.ts`])
        )

        for (let i = 0; i < 100; i++) {
            const path = `node_modules/${i}/deeply/nested/lib/file.ts`
            expect(pathExistenceChecker.shouldIncludePath(path, false)).toBeFalsy()
        }

        // Should only check children of / and /node_modules
        expect(mockGetDirectoryChildren.callCount).toEqual(2)
    })
})
