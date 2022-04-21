import * as dom from './highlightNode'

describe('util/dom', () => {
    describe('highlightNode', () => {
        const cellInnerHTML = `<div><span class="hl-source hl-tsx"><span class="hl-meta hl-function hl-tsx"><span class="hl-keyword hl-control hl-export hl-tsx">export</span> <span class="hl-storage hl-type hl-function hl-tsx">function</span> <span class="hl-meta hl-definition hl-function hl-tsx"><span class="hl-entity hl-name hl-function hl-tsx">fetchSavedSearch</span></span><span class="hl-meta hl-parameters hl-tsx"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-tsx">(</span><span class="hl-variable hl-parameter hl-tsx">id</span><span class="hl-meta hl-type hl-annotation hl-tsx"><span class="hl-keyword hl-operator hl-type hl-annotation hl-tsx">:</span> <span class="hl-entity hl-name hl-type hl-tsx">Scalars</span><span class="hl-meta hl-type hl-tuple hl-tsx"><span class="hl-meta hl-brace hl-square hl-tsx">[</span><span class="hl-string hl-quoted hl-single hl-tsx"><span class="hl-punctuation hl-definition hl-string hl-begin hl-tsx">'</span>ID<span class="hl-punctuation hl-definition hl-string hl-end hl-tsx">'</span></span><span class="hl-meta hl-brace hl-square hl-tsx">]</span></span></span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-tsx">)</span></span><span class="hl-meta hl-return hl-type hl-tsx"><span class="hl-keyword hl-operator hl-type hl-annotation hl-tsx">:</span> <span class="hl-entity hl-name hl-type hl-tsx">Observable</span><span class="hl-meta hl-type hl-parameters hl-tsx"><span class="hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx">&lt;</span></span><span class="hl-meta hl-type hl-parameters hl-tsx"><span class="hl-entity hl-name hl-type hl-module hl-tsx">GQL</span><span class="hl-punctuation hl-accessor hl-tsx">.</span><span class="hl-entity hl-name hl-type hl-tsx">ISavedSearch</span></span><span class="hl-meta hl-type hl-parameters hl-tsx"><span class="hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx">&gt;</span></span> </span><span class="hl-meta hl-block hl-tsx"><span class="hl-punctuation hl-definition hl-block hl-tsx">{</span>
</span></span></span></div>`
        let cell: HTMLTableCellElement

        beforeEach(() => {
            document.body.innerHTML = `<table><tbody><td id="cell">${cellInnerHTML}</td></tbody></table>`
            cell = window.document.querySelector<HTMLTableCellElement>('#cell')!
        })

        test('highlights no characters', () => {
            dom.highlightNode(cell, 0, 0)
            expect(cell.innerHTML).toBe(cellInnerHTML) // no changes
        })

        test('handles invalid start position', () => {
            dom.highlightNode(cell, -1, 3)
            expect(cell.innerHTML).toBe(cellInnerHTML) // no changes
            dom.highlightNode(cell, cell.textContent!.length, 3)
            expect(cell.innerHTML).toBe(cellInnerHTML) // no changes
        })

        test('handles invalid length', () => {
            dom.highlightNode(cell, 0, 2000) // length longer than cell.innerText
            expect(cell.innerHTML).toBe(cellInnerHTML) // no changes
            dom.highlightNode(cell, 22, 80) // length longer than characters between start and end
            expect(cell.innerHTML).toBe(cellInnerHTML) // no changes
        })

        test('highlights a single node', () => {
            dom.highlightNode(cell, 0, 1)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\"><span><span class=\\"match-highlight a11y-ignore\\">e</span>xport</span></span> <span class=\\"hl-storage hl-type hl-function hl-tsx\\">function</span> <span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\">fetchSavedSearch</span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\">(</span><span class=\\"hl-variable hl-parameter hl-tsx\\">id</span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Scalars</span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">[</span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\">'</span>ID<span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\">'</span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">]</span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\">)</span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Observable</span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\">&lt;</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\">GQL</span><span class=\\"hl-punctuation hl-accessor hl-tsx\\">.</span><span class=\\"hl-entity hl-name hl-type hl-tsx\\">ISavedSearch</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\">&gt;</span></span> </span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\">{</span>
                </span></span></span></div>"
            `)
        })

        test('highlights multiple nodes', () => {
            dom.highlightNode(cell, 2, 2)
            dom.highlightNode(cell, 23, 2)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\"><span>ex<span class=\\"match-highlight a11y-ignore\\">po</span>rt</span></span> <span class=\\"hl-storage hl-type hl-function hl-tsx\\">function</span> <span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\"><span>fetchSa<span class=\\"match-highlight a11y-ignore\\">ve</span>dSearch</span></span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\">(</span><span class=\\"hl-variable hl-parameter hl-tsx\\">id</span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Scalars</span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">[</span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\">'</span>ID<span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\">'</span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">]</span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\">)</span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Observable</span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\">&lt;</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\">GQL</span><span class=\\"hl-punctuation hl-accessor hl-tsx\\">.</span><span class=\\"hl-entity hl-name hl-type hl-tsx\\">ISavedSearch</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\">&gt;</span></span> </span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\">{</span>
                </span></span></span></div>"
            `)
        })

        test('does not repeatedly highlight multiple nodes', () => {
            dom.highlightNode(cell, 0, 11)
            dom.highlightNode(cell, 0, 23)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">export</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-storage hl-type hl-function hl-tsx\\"><span><span class=\\"match-highlight a11y-ignore\\">func</span><span class=\\"match-highlight a11y-ignore\\">tion</span></span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\"><span><span class=\\"match-highlight a11y-ignore\\">fetchSa</span>vedSearch</span></span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\">(</span><span class=\\"hl-variable hl-parameter hl-tsx\\">id</span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Scalars</span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">[</span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\">'</span>ID<span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\">'</span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">]</span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\">)</span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Observable</span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\">&lt;</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\">GQL</span><span class=\\"hl-punctuation hl-accessor hl-tsx\\">.</span><span class=\\"hl-entity hl-name hl-type hl-tsx\\">ISavedSearch</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\">&gt;</span></span> </span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\">{</span>
                </span></span></span></div>"
            `)
        })

        test('highlights after offset', () => {
            dom.highlightNode(cell, 2, 3)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\"><span>ex<span class=\\"match-highlight a11y-ignore\\">por</span>t</span></span> <span class=\\"hl-storage hl-type hl-function hl-tsx\\">function</span> <span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\">fetchSavedSearch</span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\">(</span><span class=\\"hl-variable hl-parameter hl-tsx\\">id</span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Scalars</span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">[</span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\">'</span>ID<span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\">'</span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\">]</span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\">)</span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\">:</span> <span class=\\"hl-entity hl-name hl-type hl-tsx\\">Observable</span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\">&lt;</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\">GQL</span><span class=\\"hl-punctuation hl-accessor hl-tsx\\">.</span><span class=\\"hl-entity hl-name hl-type hl-tsx\\">ISavedSearch</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\">&gt;</span></span> </span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\">{</span>
                </span></span></span></div>"
            `)
        })

        test('highlights entire cell', () => {
            dom.highlightNode(cell, 0, cell.textContent!.length)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">export</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-storage hl-type hl-function hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">function</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">fetchSavedSearch</span></span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">(</span></span><span class=\\"hl-variable hl-parameter hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">id</span></span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">:</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">Scalars</span></span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">[</span></span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">'</span></span><span class=\\"match-highlight a11y-ignore\\">ID</span><span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">'</span></span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">]</span></span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">)</span></span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">:</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">Observable</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">&lt;</span></span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">GQL</span></span><span class=\\"hl-punctuation hl-accessor hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">.</span></span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">ISavedSearch</span></span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">&gt;</span></span></span><span class=\\"match-highlight a11y-ignore\\"> </span></span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">{</span></span><span class=\\"match-highlight a11y-ignore\\">
                </span></span></span></span></div>"
            `)
        })

        // https://github.com/sourcegraph/sourcegraph/issues/20510
        test('highlights deeply nested spans correctly', () => {
            dom.highlightNode(cell, 26, 57)
            expect(cell.innerHTML).toMatchInlineSnapshot(`
                "<div><span class=\\"hl-source hl-tsx\\"><span class=\\"hl-meta hl-function hl-tsx\\"><span class=\\"hl-keyword hl-control hl-export hl-tsx\\">export</span> <span class=\\"hl-storage hl-type hl-function hl-tsx\\">function</span> <span class=\\"hl-meta hl-definition hl-function hl-tsx\\"><span class=\\"hl-entity hl-name hl-function hl-tsx\\"><span>fetchSaved<span class=\\"match-highlight a11y-ignore\\">Search</span></span></span></span><span class=\\"hl-meta hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-parameters hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">(</span></span><span class=\\"hl-variable hl-parameter hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">id</span></span><span class=\\"hl-meta hl-type hl-annotation hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">:</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">Scalars</span></span><span class=\\"hl-meta hl-type hl-tuple hl-tsx\\"><span class=\\"hl-meta hl-brace hl-square hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">[</span></span><span class=\\"hl-string hl-quoted hl-single hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-string hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">'</span></span><span class=\\"match-highlight a11y-ignore\\">ID</span><span class=\\"hl-punctuation hl-definition hl-string hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">'</span></span></span><span class=\\"hl-meta hl-brace hl-square hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">]</span></span></span></span><span class=\\"hl-punctuation hl-definition hl-parameters hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">)</span></span></span><span class=\\"hl-meta hl-return hl-type hl-tsx\\"><span class=\\"hl-keyword hl-operator hl-type hl-annotation hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">:</span></span><span class=\\"match-highlight a11y-ignore\\"> </span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">Observable</span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-begin hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">&lt;</span></span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-entity hl-name hl-type hl-module hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">GQL</span></span><span class=\\"hl-punctuation hl-accessor hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">.</span></span><span class=\\"hl-entity hl-name hl-type hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">ISavedSearch</span></span></span><span class=\\"hl-meta hl-type hl-parameters hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-typeparameters hl-end hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">&gt;</span></span></span><span class=\\"match-highlight a11y-ignore\\"> </span></span><span class=\\"hl-meta hl-block hl-tsx\\"><span class=\\"hl-punctuation hl-definition hl-block hl-tsx\\"><span class=\\"match-highlight a11y-ignore\\">{</span></span>
                </span></span></span></div>"
            `)
        })
    })
})
