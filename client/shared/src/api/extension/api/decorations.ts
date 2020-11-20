import { uniqueId } from 'lodash'
import { FileDecoration, TextDocumentDecorationType } from 'sourcegraph'
import * as z from 'zod'

// LINE DECORATIONS

export const createDecorationType = (): TextDocumentDecorationType => ({ key: uniqueId('TextDocumentDecorationType') })

// FILE DECORATIONS

/**
 * Returns whether the given value is a valid file decoration (uses zod)
 */
export function validateFileDecoration(fileDecoration: FileDecoration): boolean {
    // https://github.com/colinhacks/zod#safe-parse won't throw errors
    const result = fileDecorationSchema.safeParse(fileDecoration)

    if (!result.success) {
        console.error('invalid file decoration:', result.error)
        return false
    }

    // Make sure this type stays up to date. If the result is no longer assignable,
    // we have to update the zod schema.
    fileDecoration = result.data

    return true
}

const fileDecorationSchema = z.object({
    path: z.string(),
    component: z.union([z.undefined(), z.literal('panel'), z.literal('page')]),
    text: z.union([
        z.undefined(),
        z.object({
            value: z.string(),

            color: z.union([z.undefined(), z.string()]),
            light: z.union([z.undefined(), z.string()]),
            dark: z.union([z.undefined(), z.string()]),
            selected: z.union([z.undefined(), z.string()]),
            hoverMessage: z.union([z.undefined(), z.string()]),
            linkUrl: z.union([z.undefined(), z.string()]),
        }),
    ]),
    percentage: z.union([
        z.undefined(),
        z.object({
            value: z.number(),

            color: z.union([z.undefined(), z.string()]),
            light: z.union([z.undefined(), z.string()]),
            dark: z.union([z.undefined(), z.string()]),
            selected: z.union([z.undefined(), z.string()]),
            hoverMessage: z.union([z.undefined(), z.string()]),
            linkUrl: z.union([z.undefined(), z.string()]),
        }),
    ]),
})

// TODO(tj): evaluate whether or not we should validate line decorations. zod parsing takes ~1ms per
// file decoration. There are rarely enough file decorations at a time to cause problems, but a file could certainly have
// hundreds of line decorations. It might be worth the effort to write custom validators for each type (that could break UI) for performance
