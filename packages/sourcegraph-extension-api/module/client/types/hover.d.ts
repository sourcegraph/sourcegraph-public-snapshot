import { Hover, MarkupContent } from 'sourcegraph';
import { Hover as PlainHover, Range } from '../../protocol/plainTypes';
/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    /**
     * @todo Make this type *just* {@link MarkupContent} when all consumers are updated.
     */
    contents: MarkupContent | string | {
        language: string;
        value: string;
    } | (MarkupContent | string | {
        language: string;
        value: string;
    })[];
    range?: Range;
}
export declare namespace HoverMerged {
    /** Create a merged hover from the given individual hovers. */
    function from(values: (Hover | PlainHover | null | undefined)[]): HoverMerged | null;
}
