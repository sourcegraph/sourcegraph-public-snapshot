import { SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

/**
 * Returns the corresponding CSS class name for the provided syntax kind.
 * Returns an empty string if the provided kind does not exist (which could
 * happen since this data usually comes from the server).
 */
export function toCSSClassName(kind: SyntaxKind): string {
    const kindName = SyntaxKind[kind]
    return kindName ? `hl-typed-${kindName}` : ''
}
