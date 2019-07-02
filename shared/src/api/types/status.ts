import * as sourcegraph from 'sourcegraph'
import { Subscribable } from 'rxjs'
import { DiagnosticsService } from '../client/services/diagnosticsService'

export interface TransferableStatus extends Pick<sourcegraph.Status, Exclude<keyof sourcegraph.Status, 'diagnostics'>> {
    _diagnosticCollectionName?: string
}

export const toTransferableStatus = (status: sourcegraph.Status): TransferableStatus => {
    const { diagnostics, ...other } = status
    return { ...other, _diagnosticCollectionName: diagnostics ? diagnostics.name : undefined }
}

export interface ClientStatus extends Pick<sourcegraph.Status, Exclude<keyof sourcegraph.Status, 'diagnostics'>> {
    diagnostics?: Subscribable<[URL, sourcegraph.Diagnostic[]][]>
}

export const fromTransferableStatus = (
    status: TransferableStatus,
    { observe }: Pick<DiagnosticsService, 'observe'>
): ClientStatus => {
    return {
        ...status,
        diagnostics: status._diagnosticCollectionName ? observe(status._diagnosticCollectionName) : undefined,
    }
}
