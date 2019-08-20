import { Id, MonikerKind } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

export interface DocumentBlob {
    ranges: Map<Id, number>
    orderedRanges: RangeData[]
    resultSets: Map<Id, ResultSetData>
    definitionResults: Map<Id, DefinitionResultData>
    referenceResults: Map<Id, ReferenceResultData>
    hovers: Map<Id, HoverData>
    monikers: Map<Id, MonikerData>
    packageInformation: Map<Id, PackageInformationData>
}

export interface RangeData {
    id: Id
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
