import escape from 'escape-html'

import { JsonDocument, Occurrence, SyntaxKind } from './lsif-typed'

class HtmlBuilder {
    public readonly buffer: string[] = []
    public plaintext(value: string): void {
        if (!value) {
            return
        }

        this.span('', value)
    }
    public span(attributes: string, value: string): void {
        this.element('span', attributes, value)
    }
    public element(element: string, attributes: string, value: string): void {
        this.openTag(element + ' ' + attributes)
        this.raw(escape(value))
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

function openLine(html: HtmlBuilder, lineNumber: number): void {
    html.openTag('tr')
    html.raw(`<td class="line" data-line="${lineNumber + 1}"></td>`)

    html.openTag('td class="code"')
    html.openTag('div')

    // Note: Originally, we passed hl-language for language specific overrides,
    // but that seems like a bad plan, so I'm not currently doing that.
    //
    // We could add this back if we wanted though:
    // html.openTag(`span class="hl-source hl-${language}"`)

    html.openTag('span class="hl-source"')
}

function closeLine(html: HtmlBuilder): void {
    html.closeTag('span')
    html.closeTag('div')
    html.closeTag('td')
    html.closeTag('tr')
}

function highlightSlice(html: HtmlBuilder, kind: SyntaxKind, slice: string): void {
    const kindName = SyntaxKind[kind]
    if (kindName) {
        html.span(`class="hl-typed-${kindName}"`, slice)
    } else {
        html.plaintext(slice)
    }
}

// Currently assumes that no ranges overlap in the occurrences.
export function render(lsif_json: string, content: string): string {
    const occurrences = (JSON.parse(lsif_json) as JsonDocument).occurrences.map(occ => new Occurrence(occ))

    // Sort by line, and then by start character.
    occurrences.sort((a, b) => {
        if (a.range.start.line !== b.range.start.line) {
            return a.range.start.line - b.range.start.line
        }

        return a.range.start.character - b.range.start.character
    })

    const lines = content.replaceAll('\r\n', '\n').split('\n')
    const html = new HtmlBuilder()

    let occIndex = 0

    html.openTag('table')
    html.openTag('tbody')
    for (let lineNumber = 0; lineNumber < lines.length; lineNumber++) {
        openLine(html, lineNumber)

        let line = lines[lineNumber]

        let startCharacter = 0
        while (occIndex < occurrences.length && occurrences[occIndex].range.start.line === lineNumber) {
            const occ = occurrences[occIndex]
            occIndex++

            const { start, end } = occ.range

            html.plaintext(line.slice(startCharacter, start.character))

            // For multiline ranges, move the line number forward as we handle the cases.
            // This currently assumes that there are no additional matches within this range.
            // At this time, the syntax highlighter only returns non-overlapping ranges so this
            // is OK.
            if (start.line !== end.line) {
                highlightSlice(html, occ.kind, line.slice(start.character))
                closeLine(html)

                // Move to the next line
                lineNumber++

                // Handle all the lines that completely owned by this occurrence
                while (lineNumber < end.line) {
                    line = lines[lineNumber]

                    openLine(html, lineNumber)
                    highlightSlice(html, occ.kind, lines[lineNumber])
                    closeLine(html)

                    lineNumber++
                }

                // Write as much of the line as the last occurrence owns
                line = lines[lineNumber]

                openLine(html, lineNumber)
                highlightSlice(html, occ.kind, line.slice(0, end.character))
            } else {
                highlightSlice(html, occ.kind, line.slice(start.character, end.character))
            }

            startCharacter = end.character
        }

        // Highlight the remainder of the line.
        //  This could be either that some characters at the end of the line didn't match any syntax kinds
        //  or that some line didn't have any matches at all.
        if (startCharacter !== line.length) {
            html.plaintext(line.slice(startCharacter))
        }

        closeLine(html)
    }
    html.closeTag('tbody')
    html.closeTag('table')

    return html.buffer.join('')
}
