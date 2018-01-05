import * as assert from 'assert'
import { JSDOM } from 'jsdom'
import { highlightNode } from './dom'
import { getPathExtension } from './index'
import { parseHash, toBlobURL, toPrettyBlobURL, toTreeURL } from './url'

describe('util module', () => {
    describe('getPathExtension', () => {
        it('returns extension if normal path', () => {
            assert.strictEqual(getPathExtension('/foo/baz/bar.go'), 'go')
        })

        it('returns empty string if no extension', () => {
            assert.strictEqual(getPathExtension('README'), '')
        })

        it('returns empty string if hidden file with no extension', () => {
            assert.strictEqual(getPathExtension('.gitignore'), '')
        })

        it('returns extension for path with multiple dot separators', () => {
            assert.strictEqual(getPathExtension('.baz.bar.go'), 'go')
        })
    })

    describe('url module', () => {
        const linePosition = { line: 1 }
        const lineCharPosition = { line: 1, character: 1 }
        const localRefMode = { ...lineCharPosition, modal: 'references', modalMode: 'local' }
        const externalRefMode = { ...lineCharPosition, modal: 'references', modalMode: 'external' }
        const ctx = {
            repoPath: 'github.com/gorilla/mux',
            rev: '',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            filePath: 'mux.go',
        }

        describe('parseHash', () => {
            it('parses empty hash', () => {
                assert.deepStrictEqual(parseHash(''), {})
            })

            it('parses unexpectedly formatted hash', () => {
                assert.deepStrictEqual(parseHash('L-53'), {})
                assert.deepStrictEqual(parseHash('L53:'), {})
                assert.deepStrictEqual(parseHash('L1:2-'), {})
                assert.deepStrictEqual(parseHash('L1:2-3'), {})
                assert.deepStrictEqual(parseHash('L1:2-3:'), {})
                assert.deepStrictEqual(parseHash('L1:-3:'), {})
                assert.deepStrictEqual(parseHash('L1:-3:4'), {})
                assert.deepStrictEqual(parseHash('L1-2:3'), {})
                assert.deepStrictEqual(parseHash('L1-2:'), {})
                assert.deepStrictEqual(parseHash('L1:-2'), {})
                assert.deepStrictEqual(parseHash('L1:2--3:4'), {})
                assert.deepStrictEqual(parseHash('L53:a'), {})
                assert.deepStrictEqual(parseHash('L53:36$'), {})
                assert.deepStrictEqual(parseHash('L53:36$referencess'), {})
                assert.deepStrictEqual(parseHash('L53:36$references:'), {})
                assert.deepStrictEqual(parseHash('L53:36$references:trexternal'), {})
                assert.deepStrictEqual(parseHash('L53:36$references:local_'), {})
            })

            it('parses hash with leading octothorpe', () => {
                assert.deepStrictEqual(parseHash('#L1'), linePosition)
            })

            it('parses hash with line', () => {
                assert.deepStrictEqual(parseHash('L1'), linePosition)
            })

            it('parses hash with line and character', () => {
                assert.deepStrictEqual(parseHash('L1:1'), lineCharPosition)
            })

            it('parses hash with range', () => {
                assert.deepStrictEqual(parseHash('L1-2'), { line: 1, endLine: 2 })
                assert.deepStrictEqual(parseHash('L1:2-3:4'), { line: 1, character: 2, endLine: 3, endCharacter: 4 })
            })

            it('parses hash with local references', () => {
                assert.deepStrictEqual(parseHash('L1:1$references'), localRefMode)
                assert.deepStrictEqual(parseHash('L1:1$references:local'), localRefMode)
            })

            it('parses hash with external references', () => {
                assert.deepStrictEqual(parseHash('L1:1$references:external'), externalRefMode)
            })
        })

        describe('toPrettyBlobURL', () => {
            it('formats url for empty rev', () => {
                assert.strictEqual(toPrettyBlobURL(ctx), '/github.com/gorilla/mux/-/blob/mux.go')
            })

            it('formats url for specified rev', () => {
                assert.strictEqual(
                    toPrettyBlobURL({ ...ctx, rev: 'branch' }),
                    '/github.com/gorilla/mux@branch/-/blob/mux.go'
                )
            })

            it('formats url with position', () => {
                assert.strictEqual(
                    toPrettyBlobURL({ ...ctx, position: lineCharPosition }),
                    '/github.com/gorilla/mux/-/blob/mux.go#L1:1'
                )
            })

            it('formats url with references mode', () => {
                assert.strictEqual(
                    toPrettyBlobURL({ ...ctx, position: lineCharPosition, referencesMode: 'external' }),
                    '/github.com/gorilla/mux/-/blob/mux.go#L1:1$references:external'
                )
            })
        })

        describe('toBlobURL', () => {
            it('formats url if commitID is specified', () => {
                assert.strictEqual(
                    toBlobURL(ctx),
                    '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
            })

            it('formats url if commitID is ommitted', () => {
                assert.strictEqual(toBlobURL({ ...ctx, commitID: undefined }), '/github.com/gorilla/mux/-/blob/mux.go')
                assert.strictEqual(
                    toBlobURL({ ...ctx, commitID: undefined, rev: 'branch' }),
                    '/github.com/gorilla/mux@branch/-/blob/mux.go'
                )
            })
        })

        describe('toAbsoluteBlobURL', () => {
            it('formats url', () => {
                assert.strictEqual(
                    toBlobURL(ctx),
                    '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
            })

            // other cases are gratuitious given tests for other URL functions
        })

        describe('toTreeURL', () => {
            it('formats url', () => {
                assert.strictEqual(
                    toTreeURL(ctx),
                    '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/tree/mux.go'
                )
            })

            // other cases are gratuitious given tests for other URL functions
        })
    })

    describe('dom module', () => {
        describe('highlightNode', () => {
            // tslint:disable-next-line:max-line-length
            const cellInnerHTML = `<span style="color:#c0c5ce;"><span>\t</span></span><span style="color:#fff3bf;"><span>ServeHTTP</span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span>ResponseWriter</span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
            let jsdom: JSDOM
            let cell: HTMLElement

            beforeEach(() => {
                jsdom = new JSDOM('<td>' + cellInnerHTML + '</td>')
                cell = jsdom.window.document.body
            })

            it('highlights no characters', () => {
                highlightNode(cell, 0, 0, jsdom.window)
                assert.strictEqual(cell.innerHTML, cellInnerHTML) // no changes
            })

            it('handles invalid start position', () => {
                highlightNode(cell, -1, 3, jsdom.window)
                assert.strictEqual(cell.innerHTML, cellInnerHTML) // no changes
                highlightNode(cell, cell.textContent!.length, 3, jsdom.window)
                assert.strictEqual(cell.innerHTML, cellInnerHTML) // no changes
            })

            it('handles invalid length', () => {
                highlightNode(cell, 0, 63, jsdom.window) // length longer than cell.innerText
                assert.strictEqual(cell.innerHTML, cellInnerHTML) // no changes
                highlightNode(cell, 22, 53, jsdom.window) // length longer than characters between start and end
                assert.strictEqual(cell.innerHTML, cellInnerHTML) // no changes
            })

            it('highlights a single node', () => {
                highlightNode(cell, 0, 1, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span><span><span class="selection-highlight">\t</span></span></span></span><span style="color:#fff3bf;"><span>ServeHTTP</span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span>ResponseWriter</span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.strictEqual(cell.innerHTML, newCell)
            })

            it('highlights multiple nodes', () => {
                highlightNode(cell, 2, 2, jsdom.window)
                highlightNode(cell, 23, 2, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span>\t</span></span><span style="color:#fff3bf;"><span><span>S<span class="selection-highlight">er</span>veHTTP</span></span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span><span>ResponseWrit<span class="selection-highlight">er</span></span></span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.strictEqual(cell.innerHTML, newCell)
            })

            it('highlights after offset', () => {
                highlightNode(cell, 2, 3, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span>\t</span></span><span style="color:#fff3bf;"><span><span>S<span class="selection-highlight">erv</span>eHTTP</span></span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span>ResponseWriter</span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.strictEqual(cell.innerHTML, newCell)
            })

            it('highlights entire cell', () => {
                highlightNode(cell, 0, cell.textContent!.length, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span><span><span class="selection-highlight">\t</span></span></span></span><span style="color:#fff3bf;"><span><span><span class="selection-highlight">ServeHTTP</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">(</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">ResponseWriter</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">,</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight"> </span></span></span></span><span style="color:#329af0;"><span><span><span class="selection-highlight">*</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">Request</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">)</span></span></span></span>`
                assert.strictEqual(cell.innerHTML, newCell)
            })
        })
    })
})
