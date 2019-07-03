import * as sourcegraph from 'sourcegraph'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { fromRange } from '../extension/api/types'
import { Range } from '@sourcegraph/extension-api-classes'

export function fromDiagnostic(diag: sourcegraph.Diagnostic): Diagnostic {
    return {
        ...diag,
        range: fromRange(diag.range),
    }
}

export function toDiagnostic(diag: Diagnostic): sourcegraph.Diagnostic {
    return {
        ...diag,
        range: Range.fromPlain(diag.range),
    }
}

/** The format for sending {@link Diagnostic} data between the client and extension host. */
export type DiagnosticData = [string /* URL */, Diagnostic[]][]

export const fromDiagnosticData = (data: DiagnosticData): [URL, sourcegraph.Diagnostic[]][] => {
    return data.map(([uri, diagnostics]) => [new URL(uri), diagnostics.map(toDiagnostic)])
}

export const toDiagnosticData = (data: readonly [URL, readonly sourcegraph.Diagnostic[]][]): DiagnosticData => {
    return data.map(([uri, diagnostics]) => [uri.toString(), diagnostics.map(fromDiagnostic)])
}
