import { Node } from './parser'
import { visit } from './visitor'

/**
 * Returns true if a query is considered complex. A query is considered complex if there are more than one
 * of the following parameters in a query: {case, patterntype}.
 *
 * A query like `(Github case:yes) or (organisation case:no)` has this trait. It is valid
 * because the `case` parameters are descendents of an `or` operator.
 *
 * Knowing whether a query is complex helps parameterize UI states for toggle buttons. We only
 * do a simple pass to detect multiple parameters--full validation is done by the backend.
 */
export const isComplex = (nodes: Node[]): boolean => {
    const counts: Map<string, number> = new Map<string, number>()
    visit(nodes, {
        visitParameter(field: string) {
            if (field === 'case' || field === 'patterntype') {
                counts.set(field, (counts.get(field) || 0) + 1)
            }
        },
    })
    const count = (value: string, defaultValue: number = 0): number => counts.get(value) || defaultValue
    return count('case') > 1 || count('patterntype') > 1
}
