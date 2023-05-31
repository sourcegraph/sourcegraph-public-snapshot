import { filePathContains } from './fastFileFinder'

describe('filePathContains', () => {
    it('should handle exact matches', () => {
        expect(filePathContains('/a/b/c', '/a/b/c')).toBe(true)
    })
    it('should handle child directories', () => {
        expect(filePathContains('/a/b/c', 'c')).toBe(true)
        expect(filePathContains('/a/b/c', 'b/c')).toBe(true)
    })
    it('should handle mid-level directories', () => {
        expect(filePathContains('/a/b/c', 'b')).toBe(true)
        expect(filePathContains('/a/b/c', 'a/b')).toBe(true)
        expect(filePathContains('lib/batches/env/var.go', 'batches/env')).toBe(true)
        expect(filePathContains('lib/batches/env/var.go', 'lib')).toBe(true)
        expect(filePathContains('lib/batches/env/var.go', 'lib/batches')).toBe(true)
    })
    it('should handle relative paths', () => {
        expect(filePathContains('/a/b/c', './c')).toBe(true)
    })
    it('should trim separators', () => {
        expect(filePathContains('/a/b/c/', 'c/')).toBe(true)
        expect(filePathContains('/a/b/c', '/c')).toBe(true)
    })
    it('should not match if the contained path is not actually contained', () => {
        expect(filePathContains('/a/b/c', 'd')).toBe(false)
        expect(filePathContains('/a/b/c', 'a/d')).toBe(false)
    })
})
