import * as sourcegraph from 'sourcegraph'
import { Diagnostic, DiagnosticRelatedInformation } from '@sourcegraph/extension-api-types'
import { fromRange, fromLocation, toLocation } from '../extension/api/types'
import { Range } from '@sourcegraph/extension-api-classes'
import { isEqual } from 'lodash'

export function fromDiagnostic(diag: sourcegraph.Diagnostic): Diagnostic {
    if (!(diag.resource instanceof URL)) {
        throw new Error(`invalid Diagnostic#resource`)
    }
    return {
        ...diag,
        resource: diag.resource.toString(),
        range: fromRange(diag.range),
        relatedInformation: diag.relatedInformation && diag.relatedInformation.map(fromDiagnosticRelatedInformation),
    }
}

export function toDiagnostic(diag: Diagnostic): sourcegraph.Diagnostic {
    return {
        ...diag,
        resource: new URL(diag.resource),
        range: Range.fromPlain(diag.range),
        relatedInformation: diag.relatedInformation && diag.relatedInformation.map(toDiagnosticRelatedInformation),
    }
}

function fromDiagnosticRelatedInformation(
    info: sourcegraph.DiagnosticRelatedInformation
): DiagnosticRelatedInformation {
    return {
        ...info,
        location: fromLocation(info.location),
    }
}

function toDiagnosticRelatedInformation(info: DiagnosticRelatedInformation): sourcegraph.DiagnosticRelatedInformation {
    return {
        ...info,
        location: toLocation(info.location),
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

export const isDiagnosticQueryEqual = (a: sourcegraph.DiagnosticQuery, b: sourcegraph.DiagnosticQuery): boolean => {
    return (
        a.type === b.type &&
        a.tag === b.tag &&
        isEqual(a.document, b.document) &&
        (a.range && b.range ? a.range.isEqual(b.range) : false)
    )
}
