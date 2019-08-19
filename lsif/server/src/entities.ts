import { Id, MonikerKind } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

export interface DocumentBlob {
    // TODO - make searchable via two-phase binary search
    // TODO - prune things that can't be reached (see if there is redundancy here)
    ranges: Map<Id, RangeData>
    resultSets: Map<Id, ResultSetData>
    definitionResults: Map<Id, DefinitionResultData>
    referenceResults: Map<Id, ReferenceResultData>
    hovers: Map<Id, HoverData>
    monikers: Map<Id, MonikerData>
    packageInformation: Map<Id, PackageInformationData>
}

export interface RangeData {
    start: lsp.Position // TODO - flatten these
    end: lsp.Position // TODO - flatten these
    monikers: Id[]
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
    next?: Id
}

export interface ResultSetData {
    monikers: Id[]
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
    next?: Id
}

export interface DefinitionResultData {
    values: Id[]
}

export interface ReferenceResultData {
    definitions: Id[]
    references: Id[]
}

export interface HoverData {
    // TODO - normalize content
    contents: lsp.MarkupContent | lsp.MarkedString | lsp.MarkedString[]
}

export interface MonikerData {
    id: Id
    kind: MonikerKind
    scheme: string
    identifier: string
    packageInformation?: Id
}

export interface PackageInformationData {
    name: string
    version: string
}
