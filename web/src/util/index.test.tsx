import * as assert from 'assert'
import { JSDOM } from 'jsdom'
import { highlightNode } from './dom'
import { getPathExtension } from './index'
import { parseHash, toBlobURL, toPrettyBlobURL, toTreeURL } from './url'

describe('util module', () => {

    describe('getPathExtension', () => {
        it('returns extension if normal path', () => {
            assert.deepEqual(getPathExtension('/foo/baz/bar.go'), 'go')
        })

        it('returns empty string if no extension', () => {
            assert.deepEqual(getPathExtension('README'), '')
        })

        it('returns empty string if hidden file with no extension', () => {
            assert.deepEqual(getPathExtension('.gitignore'), '')
        })

        it('returns extension for path with multiple dot separators', () => {
            assert.deepEqual(getPathExtension('.baz.bar.go'), 'go')
        })
    })

    describe('url module', () => {
        const linePosition = { line: 1, character: undefined }
        const lineCharPosition = { line: 1, character: 1 }
        const localRefMode = { ...lineCharPosition, modal: 'references', modalMode: 'local' }
        const externalRefMode = { ...lineCharPosition, modal: 'references', modalMode: 'external' }
        const ctx = {
            repoPath: 'github.com/gorilla/mux',
            rev: '',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            filePath: 'mux.go'
        }

        describe('parseHash', () => {
            it('parses empty hash', () => {
                assert.deepEqual(parseHash(''), {})
            })

            it('parses unexpectedly formatted hash', () => {
                assert.deepEqual(parseHash('#L53'), {})
                assert.deepEqual(parseHash('L-53'), {})
                assert.deepEqual(parseHash('L53:'), {})
                assert.deepEqual(parseHash('L53:a'), {})
                assert.deepEqual(parseHash('L53:36$'), {})
                assert.deepEqual(parseHash('L53:36$referencess'), {})
                assert.deepEqual(parseHash('L53:36$references:'), {})
                assert.deepEqual(parseHash('L53:36$references:trexternal'), {})
                assert.deepEqual(parseHash('L53:36$references:local_'), {})
            })

            it('parses hash with line', () => {
                assert.deepEqual(parseHash('L1'), linePosition)
            })

            it('parses hash with line and character', () => {
                assert.deepEqual(parseHash('L1:1'), lineCharPosition)
            })

            it('parses hash with local references', () => {
                assert.deepEqual(parseHash('L1:1$references'), localRefMode)
                assert.deepEqual(parseHash('L1:1$references:local'), localRefMode)
            })

            it('parses hash with local references', () => {
                assert.deepEqual(parseHash('L1:1$references:external'), externalRefMode)
            })
        })

        describe('toPrettyBlobURL', () => {
            it('formats url for empty rev', () => {
                assert.deepEqual(toPrettyBlobURL(ctx), '/github.com/gorilla/mux/-/blob/mux.go')
            })

            it('formats url for specified rev', () => {
                assert.deepEqual(toPrettyBlobURL({ ...ctx, rev: 'branch' }), '/github.com/gorilla/mux@branch/-/blob/mux.go')
            })

            it('formats url with position', () => {
                assert.deepEqual(toPrettyBlobURL({ ...ctx, position: lineCharPosition }), '/github.com/gorilla/mux/-/blob/mux.go#L1:1')
            })

            it('formats url with references mode', () => {
                assert.deepEqual(
                    toPrettyBlobURL({ ...ctx, position: lineCharPosition, referencesMode: 'external' }),
                    '/github.com/gorilla/mux/-/blob/mux.go#L1:1$references:external'
                )
            })
        })

        describe('toBlobURL', () => {
            it('formats url if commitID is specified', () => {
                assert.deepEqual(toBlobURL(ctx), '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go')
            })

            it('formats url if commitID is ommitted', () => {
                assert.deepEqual(toBlobURL({ ...ctx, commitID: undefined }), '/github.com/gorilla/mux/-/blob/mux.go')
                assert.deepEqual(toBlobURL({ ...ctx, commitID: undefined, rev: 'branch' }), '/github.com/gorilla/mux@branch/-/blob/mux.go')
            })
        })

        describe('toAbsoluteBlobURL', () => {
            it('formats url', () => {
                assert.deepEqual(toBlobURL(ctx), '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go')
            })

            // other cases are gratuitious given tests for other URL functions
        })

        describe('toTreeURL', () => {
            it('formats url', () => {
                assert.deepEqual(toTreeURL(ctx), '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/tree/mux.go')
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
                assert.deepEqual(cell.innerHTML, cellInnerHTML) // no changes
            })

            it('handles invalid start position', () => {
                highlightNode(cell, -1, 3, jsdom.window)
                assert.deepEqual(cell.innerHTML, cellInnerHTML) // no changes
                highlightNode(cell, cell.textContent!.length, 3, jsdom.window)
                assert.deepEqual(cell.innerHTML, cellInnerHTML) // no changes
            })

            it('highlights a single node', () => {
                highlightNode(cell, 0, 1, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span><span><span class="selection-highlight">\t</span></span></span></span><span style="color:#fff3bf;"><span>ServeHTTP</span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span>ResponseWriter</span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.deepEqual(cell.innerHTML, newCell)
            })

            it('highlights multiple nodes', () => {
                highlightNode(cell, 0, 14, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span><span><span class="selection-highlight">\t</span></span></span></span><span style="color:#fff3bf;"><span><span><span class="selection-highlight">ServeHTTP</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">(</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">Res</span>ponseWriter</span></span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.deepEqual(cell.innerHTML, newCell)
            })

            it('highlights after offset', () => {
                highlightNode(cell, 2, 3, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span>\t</span></span><span style="color:#fff3bf;"><span><span>S<span class="selection-highlight">erv</span>eHTTP</span></span></span><span style="color:#c0c5ce;"><span>(</span></span><span style="color:#c0c5ce;"><span>ResponseWriter</span></span><span style="color:#c0c5ce;"><span>,</span></span><span style="color:#c0c5ce;"><span> </span></span><span style="color:#329af0;"><span>*</span></span><span style="color:#c0c5ce;"><span>Request</span></span><span style="color:#c0c5ce;"><span>)</span></span>`
                assert.deepEqual(cell.innerHTML, newCell)
            })

            it('highlights entire cell', () => {
                highlightNode(cell, 0, cell.textContent!.length, jsdom.window)
                // tslint:disable-next-line:max-line-length
                const newCell = `<span style="color:#c0c5ce;"><span><span><span class="selection-highlight">\t</span></span></span></span><span style="color:#fff3bf;"><span><span><span class="selection-highlight">ServeHTTP</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">(</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">ResponseWriter</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">,</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight"> </span></span></span></span><span style="color:#329af0;"><span><span><span class="selection-highlight">*</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">Request</span></span></span></span><span style="color:#c0c5ce;"><span><span><span class="selection-highlight">)</span></span></span></span>`
                assert.deepEqual(cell.innerHTML, newCell)
            })
        })
    })
})
