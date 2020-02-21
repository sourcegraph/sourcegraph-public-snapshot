import * as sinon from 'sinon'
import { PathExistenceChecker, properAncestors } from './existence'
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
            mockGetDirectoryChildren: ({ dirname }) => Promise.resolve(new Set(children.get(dirname))),
        })

        // Test within root
        expect(await pathExistenceChecker.shouldIncludePath('foo.ts', false)).toBeTruthy()
        expect(await pathExistenceChecker.shouldIncludePath('bar.ts', false)).toBeFalsy()
        expect(await pathExistenceChecker.shouldIncludePath('shared/bonk.ts', false)).toBeTruthy()

        // Test outside root but within repo
        expect(await pathExistenceChecker.shouldIncludePath('../shared/bar.ts', false)).toBeTruthy()
        expect(await pathExistenceChecker.shouldIncludePath('../shared/bar.ts', true)).toBeFalsy()
        expect(await pathExistenceChecker.shouldIncludePath('../shared/bonk.ts', false)).toBeFalsy()
        expect(await pathExistenceChecker.shouldIncludePath('../node_modules/@types/quux.ts', false)).toBeFalsy()

        // Test outside repo
        expect(await pathExistenceChecker.shouldIncludePath('../../node_modules/@types/oops.ts', false)).toBeFalsy()
    })

    it('should cache directory contents', async () => {
        const children = new Map([['', Array.from(range(0, 100).map(i => `${i}.ts`))]])
        const mockGetDirectoryChildren = sinon.spy<typeof getDirectoryChildren>(({ dirname }) =>
            Promise.resolve(new Set(children.get(dirname)))
        )

        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: 42,
            commit: 'c',
            root: '',
            frontendUrl: 'frontend',
            mockGetDirectoryChildren,
        })

        for (let i = 0; i < 100; i++) {
            expect(await pathExistenceChecker.shouldIncludePath(`${i}.ts`, false)).toBeTruthy()
            expect(await pathExistenceChecker.shouldIncludePath(`${i}.js`, false)).toBeFalsy()
        }

        expect(mockGetDirectoryChildren.callCount).toEqual(1)
    })

    it('should early out on untracked ancestors', async () => {
        const children = new Map([['', ['not_node_modules']]])
        const mockGetDirectoryChildren = sinon.spy<typeof getDirectoryChildren>(({ dirname }) =>
            Promise.resolve(new Set(children.get(dirname)))
        )

        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: 42,
            commit: 'c',
            root: '',
            frontendUrl: 'frontend',
            mockGetDirectoryChildren,
        })

        for (let i = 0; i < 100; i++) {
            const path = `node_modules/${i}/deeply/nested/lib/file.ts`
            expect(await pathExistenceChecker.shouldIncludePath(path, false)).toBeFalsy()
        }

        // Should only check children of / and /node_modules
        expect(mockGetDirectoryChildren.callCount).toEqual(2)
    })
})

describe('properAncestors', () => {
    it('should return all ancestor directories', () => {
        expect(properAncestors('foo/bar/baz/bonk')).toEqual(['', 'foo', 'foo/bar', 'foo/bar/baz'])
    })
})
