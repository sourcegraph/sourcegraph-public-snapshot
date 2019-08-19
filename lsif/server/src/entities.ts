import { Id, MonikerKind } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

// TODO - remove these, convert to small int ids where possible
export interface LiteralMap<T> {
    [key: string]: T
    [key: number]: T
}

export interface DocumentBlob {
    // TODO - make searchable via two-phase binary search
    // TODO - prune things that can't be reached (see if there is redundancy here)
    ranges: LiteralMap<RangeData>
    resultSets: LiteralMap<ResultSetData>
    definitionResults: LiteralMap<DefinitionResultData>
    referenceResults: LiteralMap<ReferenceResultData>
    hovers: LiteralMap<HoverData>
    monikers: LiteralMap<MonikerData>
    packageInformation: LiteralMap<PackageInformationData>
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
