import { Id, MonikerKind, Uri } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

// TODO - document these
export interface DocumentData {
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
    // TODO - used MarkupContent, MarkedString is deprecated
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

//
//

export interface XrepoSymbols {
    packages: Package[]
    references: SymbolReferences[]
}

export interface Package {
    scheme: string
    name: string
    version: string
}

export interface SymbolReferences {
    package: Package
    identifiers: string[]
}

//
//

export interface DocumentMeta {
    id: Id
    uri: Uri
}

//
//

export interface FlattenedRange {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

//
//

export interface WrappedDocumentData extends DocumentData {
    id: Id
    uri: string
    contains: Id[]
    definitions: MonikerScopedResultData<DefinitionResultData>[]
    references: MonikerScopedResultData<ReferenceResultData>[]
}

export interface MonikerScopedResultData<T> {
    moniker: MonikerData
    data: T
}

//
//
