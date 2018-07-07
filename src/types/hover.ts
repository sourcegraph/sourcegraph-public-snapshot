import { Hover, MarkedString, MarkupContent, Range } from 'vscode-languageserver-types'

/** A hover that is merged from multiple Hover results and normalized. */
export type HoverMerged = Pick<Hover, Exclude<keyof Hover, 'contents'>> & {
    /** Also allows MarkupContent[]. */
    // tslint:disable-next-line:deprecation
    contents: (MarkupContent | MarkedString)[]
}

export namespace HoverMerged {
    /** Create a merged hover from the given individual hovers. */
    export function from(values: Hover[]): HoverMerged | null {
        const contents: HoverMerged['contents'] = []
        let range: HoverMerged['range']
        for (const result of values) {
            if (result) {
                if (Array.isArray(result.contents)) {
                    contents.push(...result.contents)
                } else {
                    contents.push(result.contents)
                }
                if (result.range && !range) {
                    range = result.range
                }
            }
        }
        return contents.length === 0 ? null : { contents, range }
    }

    /** Reports whether the value conforms to the HoverMerged interface. */
    export function is(value: any): value is HoverMerged {
        // Based on Hover.is from vscode-languageserver-types.
        return (
            value !== null &&
            typeof value === 'object' &&
            Array.isArray(value.contents) &&
            // tslint:disable-next-line:deprecation
            (value.contents as any[]).every(c => MarkupContent.is(c) || MarkedString.is(c)) &&
            (value.range === undefined || Range.is(value.range))
        )
    }
}
