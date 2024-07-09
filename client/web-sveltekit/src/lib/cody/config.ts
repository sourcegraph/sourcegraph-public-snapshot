import { memoize } from 'lodash'
import { z } from 'zod'

// This is used in the site config to enable/disable cody for specific repositories.

/**
 * Re2Expression is a zod schema for a RE2 regular expression.
 */
const Re2Expression = z.string().transform(async (val, ctx) => {
    try {
        const { RE2JS } = await import('re2js')
        return RE2JS.compile(val)
    } catch (error) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: 'Failed to parse r2 regex',
        })
        return z.NEVER
    }
})

const CodyContextFilterItem = z.object({
    repoNamePattern: Re2Expression,
})

/**
 * CodyContextFilters is a zod schema for the filters used to enable/disable cody for specific repositories.
 * It imports the RE2 parsing library and therefore needs to be
 * called with parseSync or safeParseAsync.
 */
export const CodyContextFiltersSchema = z
    .object({
        include: z.array(CodyContextFilterItem),
        exclude: z.array(CodyContextFilterItem),
    })
    .partial()

/**
 * CodyContextFilters describes the filters used to enable/disable cody for specific repositories.
 */
export type CodyContextFilters = z.infer<typeof CodyContextFiltersSchema>

/**
 * getFiltersFromCodyContextFilters imports the RE2 parsing library and returns a validation function.
 * That function returns true if a repo matches any of the include filters and none of the exclude filters.
 *
 * If filters include repo name patterns that are not valid regexes the function
 * always returns false.
 */
export const getFiltersFromCodyContextFilters = memoize(
    ({ include, exclude }: CodyContextFilters): ((repoName: string) => boolean) =>
        (repoName: string): boolean => {
            const isIncluded = !include?.length || include.some(filter => filter.repoNamePattern.matches(repoName))
            const isExcluded = exclude?.some(filter => filter.repoNamePattern.matches(repoName))
            return isIncluded && !isExcluded
        },
    filters => filters
)
