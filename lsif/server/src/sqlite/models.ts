import { Id, RangeTag, MonikerKind, Uri } from 'lsif-protocol'
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
    hovers: LiteralMap<lsp.Hover>
    monikers: LiteralMap<MonikerData>
    packageInformation: LiteralMap<PackageInformationData>
}

export interface RangeData {
    start: lsp.Position
    end: lsp.Position
    tag?: RangeTag
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface ResultSetData {
    monikers?: Id[]
    next?: Id
    hoverResult?: Id
    definitionResult?: Id
    referenceResult?: Id
}

export interface DefinitionResultData {
    values: Id[]
}

export interface ReferenceResultData {
    definitions?: Id[]
    references?: Id[]
}

export interface MonikerData {
    id: Id
    scheme: string
    identifier: string
    kind?: MonikerKind
    packageInformation?: Id
}

export interface PackageInformationData {
    name: string
    manager: string
    uri?: Uri
    version?: string
    repository?: {
        type: string
        url: string
        commitId?: string
    }
}
