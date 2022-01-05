import { Position, Range } from 'sourcegraph'

import Parser, { Tree } from 'web-tree-sitter'

export interface Input {
    path: string
    text: string
}

export interface LsifDocument {
    occurrences: LsifOccurrence[]
}

export class LsifDocumentBuilder {
    public readonly occurrences: LsifOccurrence[] = []

    public pushHighlight(node: Parser.SyntaxNode, highlight: LsifHighlight): void {
        this.occurrences.push({ range: this.range(node), highlight })
    }

    private range(node: Parser.SyntaxNode): Range {
        return new Range(
            new Position(node.startPosition.row, node.startPosition.column),
            new Position(node.endPosition.row, node.endPosition.column)
        )
    }
}

export enum LsifRole {
    DEFINITION,
    REFERENCE,
}
export enum LsifHighlight {
    STRING_LITERAL = 'string',
    LOCAL_IDENTIFIER = 'variable',
    KEYWORD = 'keyword',
    NUMERIC_LITERAL = 'integer',
}

export interface LsifOccurrence {
    range: Range
    moniker?: string
    role?: LsifRole
    highlight?: LsifHighlight
}

class HtmlBuilder {
    public readonly buffer: string[] = []
    public plaintext(value: string): void {
        this.span('', value)
    }
    public span(attributes: string, value: string): void {
        this.element('span', attributes, value)
    }
    public element(element: string, attributes: string, value: string): void {
        this.openTag(element + ' ' + attributes)
        this.raw(value)
        this.closeTag(element)
    }
    public raw(html: string): void {
        this.buffer.push(html)
    }
    public openTag(tag: string): void {
        this.buffer.push('<')
        this.buffer.push(tag)
        this.buffer.push('>')
    }
    public closeTag(tag: string): void {
        this.buffer.push('</')
        this.buffer.push(tag)
        this.buffer.push('>')
    }
}

export abstract class Indexer {
    constructor(public readonly language: string) {}
    abstract index(input: Input): Promise<LsifDocument>
    public matchesFilePath(filePath: string): boolean {
        return filePath.endsWith('.' + this.language)
    }

    public async highlight(input: Input): Promise<string> {
        const document = await this.index(input)
        const lines = input.text.replaceAll('\r\n', '\n').split('\n')
        const html = new HtmlBuilder()
        html.openTag('table')
        html.openTag('tbody')
        let documentIndex = 0
        for (const [lineNumber, line] of lines.entries()) {
            html.openTag('tr')
            html.raw(`<td class="line" data-line="${lineNumber + 1}"></td>`)

            html.openTag('td class="code"')
            html.openTag('div')
            html.openTag(`span class="hl-source hl-${this.language}"`)
            const start = 0
            while (document.occurrences[documentIndex].range.start.line === lineNumber) {
                const occurrence = document.occurrences[documentIndex]
                if (occurrence.highlight) {
                    html.plaintext(line.slice(start, occurrence.range.start.character))
                    html.span(
                        LsifHighlight[occurrence.highlight],
                        line.slice(occurrence.range.start.character, occurrence.range.end.character)
                    )
                }
                documentIndex++
            }
            html.closeTag('span')
            html.closeTag('div')
            html.closeTag('td')

            html.closeTag('tr')
        }
        html.closeTag('tbody')
        html.closeTag('table')
        return html.buffer.join('')
    }

    public async parseInput(input: Input): Promise<Tree> {
        await Parser.init({
            locateFile() {
                // If you don't do this, then you will get a 404 encoded confusingly
                // which is why you get magic word complaints.
                return 'https://tree-sitter.github.io/tree-sitter.wasm'
            },
        })
        console.log('We got the parser loaded!')
        const parser = new Parser()
        const language = await Parser.Language.load(`https://tree-sitter.github.io/tree-sitter-${this.language}.wasm`)
        parser.setLanguage(language)
        return parser.parse(input.text)
    }
}
