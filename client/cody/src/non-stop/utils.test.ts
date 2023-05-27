import { getFileNameAfterLastDash } from './utils'

// Test for getFileNameAfterLastDash
describe('getFileNameAfterLastDash', () => {
    test('gets the last part of the file path after the last slash', () => {
        const filePath = '/path/to/file.txt'
        const fileName = 'file.txt'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
    test('get file name when there is no slash', () => {
        const filePath = 'file.txt'
        const fileName = 'file.txt'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
    test('get file name when there is no extension', () => {
        const filePath = 'file'
        const fileName = 'file'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
})
