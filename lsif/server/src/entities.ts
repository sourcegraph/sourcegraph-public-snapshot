import { Id, MonikerKind, Uri } from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

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
    imported: SymbolReference[]
    exported: Package[]
}

export interface SymbolReference {
    scheme: string
    name: string
    version: string
    identifier: string
}

export interface Package {
    scheme: string
    name: string
    version: string
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
