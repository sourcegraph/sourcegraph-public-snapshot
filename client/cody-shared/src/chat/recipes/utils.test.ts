import * as assert from 'assert'

import { stripPathPrefix, getDirectoryPathsBetween } from './utils'

describe('getDirectoryPathsBetween', () => {
    it('should return empty array if ancestor is not actually an ancestor', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('/foo', '/bar'), [])
    })

    it('should return array with single path if ancestor and descendant are the same', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('/foo', '/foo'), ['/foo'])
    })

    it('should return array with paths from descendant to ancestor', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('/foo/bar/baz', '/foo/bar'), [])
    })

    it('should return array with paths from descendant to ancestor', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('/foo/bar', '/foo/bar/baz'), ['/foo/bar/baz', '/foo/bar'])
    })

    it('should handle root path correctly', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('/', '/foo/bar/baz'), ['/foo/bar/baz', '/foo/bar', '/foo', '/'])
    })

    it('empty path is treated as root', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('', '/foo/bar/baz'), ['/foo/bar/baz', '/foo/bar', '/foo', '/'])
    })

    it('relative paths', () => {
        assert.deepStrictEqual(getDirectoryPathsBetween('foo', 'foo/bar/baz'), ['foo/bar/baz', 'foo/bar', 'foo'])
    })
})

describe('stripPathPrefix', () => {
    it('should strip the prefix from a path', () => {
        expect(stripPathPrefix('/home/user', '/home/user/file.txt')).toEqual('file.txt')
    })

    it('should return null if the path does not start with the prefix', () => {
        expect(stripPathPrefix('/home/user', '/etc/file.txt')).toBeNull()
    })

    it('should handle prefixes that end in a separator', () => {
        expect(stripPathPrefix('/home/user/', '/home/user/file.txt')).toEqual('file.txt')
    })

    it('should handle prefixes that do not end in a separator', () => {
        expect(stripPathPrefix('/home/user', '/home/user/file.txt')).toEqual('file.txt')
    })

    it('should return a single separator if the path is the same as the prefix', () => {
        expect(stripPathPrefix('/home/user', '/home/user')).toEqual('/')
    })

    it('should return relative path with prefix stripped from start', () => {
        expect(stripPathPrefix('foo/bar', 'foo/bar/baz.txt')).toEqual('baz.txt')
    })

    it('should return null if relative path does not start with prefix', () => {
        expect(stripPathPrefix('foo/bar', 'baz/foo/bar/baz.txt')).toBeNull()
    })
})
