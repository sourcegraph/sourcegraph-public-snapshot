/* --------------------------------------------------------------------------------------------
 * Copyright (c) Sourcegraph and Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License.
 * ------------------------------------------------------------------------------------------ */

import * as protocol from 'lsif-protocol'
import * as lsp from 'vscode-languageserver-protocol'

export enum CompressionKind {
    scalar = 'scalar',
    literal = 'literal',
    array = 'array',
    any = 'any',
    raw = 'raw',
}

export interface CompressorPropertyDescription {
    /**
     * The name of the property.
     */
    name: string

    /**
     * It's index in the array.
     */
    index: number

    /**
     * Whether the value is raw in case it was an object literal.
     */
    compressionKind: CompressionKind

    /**
     * Short form if the value is a string.
     */
    shortForm?: [string, string | number][]
}

export interface CompressorData {
    vertexCompressor: number
    edgeCompressor: number
    itemEdgeCompressor: number
    all: CompressorDescription[]
}

export interface CompressorDescription {
    /**
     * The compressor id.
     */
    id: number

    /**
     * The parent compressor or undefined.
     */
    parent: number | undefined

    /**
     * The compressed propeties.
     */
    properties: CompressorPropertyDescription[]
}

/**
 * The meta data vertex.
 */
export interface MetaData extends protocol.MetaData {
    /**
     * A description of the compressor used.
     */
    compressors?: CompressorData
}

namespace Is {
    export function string(value: any): value is string {
        return typeof value === 'string' || value instanceof String
    }

    export function number(value: any): value is number {
        return typeof value === 'number' || value instanceof Number
    }
}

export type BaseCompressValue = boolean | number | string | object
export type CompressValue = BaseCompressValue | undefined | CompressArray
export interface CompressArray extends Array<CompressValue> {}

export interface CompressorProperty {
    name: string | number | symbol
    index: number
    shortForm: [string, string | number][] | undefined
    compressionKind: CompressionKind
}

export namespace CompressorProperty {
    export function scalar<T>(
        name: keyof T,
        index: number,
        shortForm?: [string, string | number][]
    ): CompressorProperty {
        return { name, index, shortForm, compressionKind: CompressionKind.scalar }
    }
    export function literal<T>(name: keyof T, index: number): CompressorProperty {
        return { name, index, shortForm: undefined, compressionKind: CompressionKind.literal }
    }
    export function create<T>(
        name: keyof T,
        index: number,
        compressionKind: CompressionKind,
        shortForm?: [string, string | number][]
    ): CompressorProperty {
        return { name, index, shortForm, compressionKind }
    }
}

export interface CompressorOptions {
    mode: 'store' | 'hash'
}

export namespace CompressorOptions {
    export const defaults: CompressorOptions = Object.freeze({
        mode: 'store',
    })

    export function is(value: Partial<CompressorOptions>): value is CompressorOptions {
        let candidate: CompressorOptions = value as CompressorOptions
        return candidate !== undefined && typeof candidate.mode === 'string'
    }

    export function fillDefaults(value?: Partial<CompressorOptions>): CompressorOptions {
        if (value === undefined) {
            return Object.assign({}, CompressorOptions.defaults)
        }
        return Object.assign({}, CompressorOptions.defaults, value)
    }
}

export abstract class Compressor<T> {
    private static compressors: Compressor<any>[] = []
    private static vertexCompressors: Map<string, Compressor<any>> = new Map()
    private static edgeCompressors: Map<string, Compressor<any>> = new Map()

    public static addCompressor(compressor: Compressor<any>): void {
        this.compressors.push(compressor)
    }

    public static registerVertexCompressor(key: protocol.VertexLabels, compress: Compressor<any>) {
        this.vertexCompressors.set(key, compress)
    }

    public static registerEdgeCompressor(key: protocol.EdgeLabels, compress: Compressor<any>) {
        this.edgeCompressors.set(key, compress)
    }

    public static getCompressor(element: protocol.Edge | protocol.Vertex): Compressor<any> | undefined {
        if (element.type === protocol.ElementTypes.vertex) {
            return this.vertexCompressors.get(element.label)
        } else {
            return this.edgeCompressors.get(element.label)
        }
    }

    public static getVertexCompressor(label: protocol.VertexLabels): Compressor<any> | undefined {
        return this.vertexCompressors.get(label)
    }

    public static allCompressors(): Compressor<any>[] {
        let result = this.compressors.slice(0)
        for (let elem of this.vertexCompressors.values()) {
            result.push(elem)
        }
        for (let elem of this.edgeCompressors.values()) {
            result.push(elem)
        }
        return result
    }

    private static _id: number = 0

    public static nextId(): number {
        return this._id++
    }

    public abstract get id(): number

    public abstract get parent(): Compressor<any> | undefined

    public abstract get index(): number

    public abstract compress(element: T, options: CompressorOptions, level?: number): CompressValue

    public abstract metaData(): CompressorProperty[]
}

export interface GenericCompressorProperty<T, P = void> {
    name: keyof T
    index: number
    compressionKind: CompressionKind
    shortForm: Map<string, string | number> | undefined
    compressor: Compressor<any> | ((value: any) => Compressor<any> | undefined) | undefined
}

export namespace GenericCompressorProperty {
    export function scalar<T, P = void>(
        name: keyof T,
        index: number,
        shortForm?: Map<string, string | number>
    ): GenericCompressorProperty<T, P> {
        return { name, index, compressionKind: CompressionKind.scalar, shortForm, compressor: undefined }
    }
    export function raw<T, P = void>(name: keyof T, index: number): GenericCompressorProperty<T, P> {
        return { name, index, compressionKind: CompressionKind.raw, shortForm: undefined, compressor: undefined }
    }
    export function array<T, P = void>(
        name: keyof T,
        index: number,
        compressor?: Compressor<any> | ((value: any) => Compressor<any> | undefined)
    ): GenericCompressorProperty<T, P> {
        return { name, index, compressionKind: CompressionKind.array, shortForm: undefined, compressor }
    }
    export function literal<T, P = void>(
        name: keyof T,
        index: number,
        compressor: Compressor<any> | ((value: any) => Compressor<any> | undefined)
    ): GenericCompressorProperty<T, P> {
        return { name, index, compressionKind: CompressionKind.literal, shortForm: undefined, compressor }
    }
    export function any<T, P = void>(
        name: keyof T,
        index: number,
        compressor: (value: any) => Compressor<any> | undefined
    ): GenericCompressorProperty<T, P> {
        return { name, index, compressionKind: CompressionKind.any, shortForm: undefined, compressor }
    }
}

export class GenericCompressor<T> extends Compressor<T> {
    private _parent: Compressor<any> | undefined
    private _id: number
    private _index: number
    private _properties: GenericCompressorProperty<T, any>[]

    public constructor(
        parent: Compressor<any> | undefined,
        id: number,
        properties: (next: () => number, that: GenericCompressor<T>) => GenericCompressorProperty<T, any>[]
    ) {
        super()
        this._parent = parent
        this._id = id
        this._index = parent === undefined ? 1 : parent.index
        this._properties = properties(() => this._index++, this)
    }

    public get id(): number {
        return this._id
    }

    public get parent(): Compressor<any> | undefined {
        return this._parent
    }

    public get index(): number {
        return this._index
    }

    public compress(element: T, options: CompressorOptions, level: number = 0): CompressValue {
        let result: CompressArray =
            this._parent !== undefined ? (this._parent.compress(element, options, level + 1) as CompressArray) : []
        if (level === 0) {
            result.unshift(options.mode === 'store' ? this._id : -1)
        }
        let undefinedElements: number = 0
        const recordUndefined = () => {
            undefinedElements++
        }
        const pushUndefined = () => {
            for (let i = 0; i < undefinedElements; i++) {
                result.push(undefined)
            }
            undefinedElements = 0
        }
        function getCompressor(
            value: Compressor<any> | ((value: any) => Compressor<any> | undefined) | undefined,
            prop: any,
            force: true
        ): Compressor<any>
        function getCompressor(
            value: Compressor<any> | ((value: any) => Compressor<any> | undefined) | undefined,
            prop: any,
            force?: boolean
        ): Compressor<any> | undefined
        function getCompressor(
            value: Compressor<any> | ((value: any) => Compressor<any> | undefined) | undefined,
            prop: any,
            force: boolean = false
        ): Compressor<any> | undefined {
            if (value instanceof Compressor) {
                return value
            }
            if (value === undefined) {
                if (force === true) {
                    throw new Error('No compressor available')
                } else {
                    return undefined
                }
            }
            return value(prop)
        }

        for (let item of this._properties) {
            let value = element[item.name]
            if (value === undefined || value === null) {
                recordUndefined()
                continue
            }
            switch (item.compressionKind) {
                case CompressionKind.raw:
                    pushUndefined()
                    result.push((value as unknown) as CompressValue)
                    break
                case CompressionKind.scalar:
                    let convertedScalar: any = value
                    if (item.shortForm !== undefined && Is.string(value)) {
                        let short = item.shortForm.get(value)
                        if (short !== undefined) {
                            convertedScalar = short
                        }
                    }
                    pushUndefined()
                    result.push(convertedScalar as CompressValue)
                    break
                case CompressionKind.literal:
                    let c1 = getCompressor(item.compressor, value, true)
                    pushUndefined()
                    result.push(c1.compress(value, options))
                    break
                case CompressionKind.array:
                    if (!Array.isArray(value)) {
                        throw new Error('Type mismatch. Compressor property declares array but value is not an array')
                    }
                    let convertedArray: any[] = []
                    for (let element of value) {
                        let c2 = getCompressor(item.compressor, element, false)
                        if (c2 !== undefined) {
                            convertedArray.push(c2.compress(element, options))
                        } else {
                            convertedArray.push(element)
                        }
                    }
                    pushUndefined()
                    result.push(convertedArray)
                    break
                case CompressionKind.any:
                    const handleValue = (value: any): any => {
                        let compresor = getCompressor(item.compressor, value, false)
                        if (compresor !== undefined) {
                            return compresor.compress(value, options)
                        }
                        let type = typeof value
                        if (type === 'number' || type === 'string' || type === 'boolean') {
                            return value
                        }
                        throw new Error(`Any compression kind can't infer conversion for property ${item.name}`)
                    }
                    let convertedAny: any
                    if (Array.isArray(value)) {
                        convertedAny = []
                        for (let element of value) {
                            ;(convertedAny as any[]).push(handleValue(element))
                        }
                    } else {
                        convertedAny = handleValue(value)
                    }
                    pushUndefined()
                    result.push(convertedAny)
                    break
                default:
                    throw new Error(`Comresion kind ${item.compressionKind} unknown.`)
            }
        }
        return result
    }

    public metaData(): CompressorProperty[] {
        let result: CompressorProperty[] = []
        for (let item of this._properties) {
            let shortForm: [string, string | number][] | undefined = undefined
            if (item.shortForm !== undefined) {
                shortForm = []
                for (let entry of item.shortForm.entries()) {
                    shortForm.push(entry)
                }
            }
            result.push(CompressorProperty.create(item.name, item.index, item.compressionKind, shortForm))
        }
        return result
    }
}

class NoopCompressor extends Compressor<number | string | boolean> {
    private _id: number

    constructor() {
        super()
        this._id = Compressor.nextId()
    }

    public compress(value: number | string | boolean): CompressValue {
        return value
    }

    public metaData(): CompressorProperty[] {
        return []
    }

    public get id(): number {
        return this._id
    }

    public get parent(): undefined {
        return undefined
    }

    public get index(): number {
        return 0
    }
}
const noopCompressor = new NoopCompressor()
Compressor.addCompressor(noopCompressor)

const elementShortForms = (function() {
    return new Map<protocol.ElementTypes, number>([[protocol.ElementTypes.vertex, 1], [protocol.ElementTypes.edge, 2]])
})()

class ElementCompressor extends Compressor<protocol.Element> {
    private _id: number

    public constructor() {
        super()
        this._id = Compressor.nextId()
    }

    public compress(element: protocol.Element, options: CompressorOptions): [protocol.Id, number] {
        if (options.mode === 'store') {
            return [element.id, elementShortForms.get(element.type)!]
        } else {
            return [-1, -1]
        }
    }

    public metaData(): CompressorProperty[] {
        return [CompressorProperty.scalar('id', 1), CompressorProperty.scalar('type', 2)]
    }

    public get id(): number {
        return this._id
    }

    public get parent(): undefined {
        return undefined
    }

    public get index(): number {
        return 3
    }
}

const elementCompressor = new ElementCompressor()
Compressor.addCompressor(elementCompressor)

export const vertexShortForms = (function() {
    let shortCounter: number = 1
    return new Map<protocol.VertexLabels, number>([
        [protocol.VertexLabels.metaData, shortCounter++],
        [protocol.VertexLabels.event, shortCounter++],
        [protocol.VertexLabels.project, shortCounter++],
        [protocol.VertexLabels.range, shortCounter++],
        [protocol.VertexLabels.location, shortCounter++],
        [protocol.VertexLabels.document, shortCounter++],
        [protocol.VertexLabels.moniker, shortCounter++],
        [protocol.VertexLabels.packageInformation, shortCounter++],
        [protocol.VertexLabels.resultSet, shortCounter++],
        [protocol.VertexLabels.documentSymbolResult, shortCounter++],
        [protocol.VertexLabels.foldingRangeResult, shortCounter++],
        [protocol.VertexLabels.documentLinkResult, shortCounter++],
        [protocol.VertexLabels.diagnosticResult, shortCounter++],
        [protocol.VertexLabels.declarationResult, shortCounter++],
        [protocol.VertexLabels.definitionResult, shortCounter++],
        [protocol.VertexLabels.typeDefinitionResult, shortCounter++],
        [protocol.VertexLabels.hoverResult, shortCounter++],
        [protocol.VertexLabels.referenceResult, shortCounter++],
        [protocol.VertexLabels.implementationResult, shortCounter++],
    ])
})()

export const vertexCompressor = new GenericCompressor<protocol.V>(elementCompressor, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('label', next(), vertexShortForms),
])
Compressor.addCompressor(vertexCompressor)

const resultCompressor = new GenericCompressor<protocol.ResultSet>(vertexCompressor, Compressor.nextId(), () => [])
Compressor.registerVertexCompressor(protocol.VertexLabels.resultSet, resultCompressor)

class LspRangeCompressor extends Compressor<lsp.Range> {
    private _id: number

    public constructor() {
        super()
        this._id = Compressor.nextId()
    }

    public compress(element: lsp.Range, options: CompressorOptions): [number, number, number, number, number] {
        return [
            options.mode === 'store' ? this.id : -1,
            element.start.line,
            element.start.character,
            element.end.line,
            element.end.character,
        ]
    }

    public metaData(): CompressorProperty[] {
        return [
            CompressorProperty.scalar('start.line', 1),
            CompressorProperty.scalar('start.character', 2),
            CompressorProperty.scalar('end.line', 3),
            CompressorProperty.scalar('end.character', 4),
        ]
    }

    public get id(): number {
        return this._id
    }

    public get parent(): undefined {
        return undefined
    }

    public get index(): number {
        return 5
    }
}
const lspRangeCompressor = new LspRangeCompressor()
Compressor.addCompressor(lspRangeCompressor)

const rangeTagCompressor = new GenericCompressor<protocol.RangeTag>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar(
        'type',
        next(),
        new Map<protocol.RangeTagTypes, number>([
            [protocol.RangeTagTypes.declaration, 1],
            [protocol.RangeTagTypes.definition, 2],
            [protocol.RangeTagTypes.reference, 3],
            [protocol.RangeTagTypes.unknown, 4],
        ])
    ),
])
Compressor.addCompressor(rangeTagCompressor)

const declarationTagCompressor = new GenericCompressor<protocol.DeclarationTag>(
    rangeTagCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('text', next()),
        GenericCompressorProperty.scalar('kind', next()),
        GenericCompressorProperty.scalar('deprecated', next()),
        GenericCompressorProperty.literal('fullRange', next(), lspRangeCompressor),
        GenericCompressorProperty.scalar('detail', next()),
    ]
)
Compressor.addCompressor(declarationTagCompressor)

const definitionTagCompressor = new GenericCompressor<protocol.DefinitionTag>(
    rangeTagCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('text', next()),
        GenericCompressorProperty.scalar('kind', next()),
        GenericCompressorProperty.scalar('deprecated', next()),
        GenericCompressorProperty.literal('fullRange', next(), lspRangeCompressor),
        GenericCompressorProperty.scalar('detail', next()),
    ]
)
Compressor.addCompressor(definitionTagCompressor)

const referenceTagCompressor = new GenericCompressor<protocol.ReferenceTag>(
    rangeTagCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.scalar('text', next())]
)
Compressor.addCompressor(referenceTagCompressor)

const unknownTagCompressor = new GenericCompressor<protocol.UnknownTag>(
    rangeTagCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.scalar('text', next())]
)
Compressor.addCompressor(unknownTagCompressor)

class RangeCompressor extends Compressor<protocol.Range> {
    private _id: number
    private _index: number

    public constructor() {
        super()
        this._id = Compressor.nextId()
        this._index = vertexCompressor.index
    }

    public compress(element: protocol.Range, options: CompressorOptions): CompressValue {
        let result = vertexCompressor.compress(element, options, 1) as CompressArray
        result.unshift(options.mode === 'store' ? this._id : -1)
        result.push(element.start.line, element.start.character, element.end.line, element.end.character)
        if (element.tag) {
            switch (element.tag.type) {
                case protocol.RangeTagTypes.declaration:
                    result.push(declarationTagCompressor.compress(element.tag, options))
                    break
                case protocol.RangeTagTypes.definition:
                    result.push(definitionTagCompressor.compress(element.tag, options))
                    break
                case protocol.RangeTagTypes.reference:
                    result.push(referenceTagCompressor.compress(element.tag, options))
                    break
                case protocol.RangeTagTypes.unknown:
                    result.push(unknownTagCompressor.compress(element.tag, options))
                    break
            }
        } else {
            result.push(undefined)
        }
        return result
    }

    public metaData(): CompressorProperty[] {
        return [
            CompressorProperty.scalar('start.line', this._index++),
            CompressorProperty.scalar('start.character', this._index++),
            CompressorProperty.scalar('end.line', this._index++),
            CompressorProperty.scalar('end.character', this._index++),
            CompressorProperty.literal('tag', this._index++),
        ]
    }

    public get id(): number {
        return this._id
    }

    public get parent(): Compressor<any> | undefined {
        return vertexCompressor
    }

    public get index(): number {
        return this._index
    }
}
const rangeCompressor = new RangeCompressor()
Compressor.registerVertexCompressor(protocol.VertexLabels.range, rangeCompressor)

const locationCompressor = new GenericCompressor<protocol.Location>(vertexCompressor, Compressor.nextId(), next => [
    GenericCompressorProperty.literal('range', next(), lspRangeCompressor),
])
Compressor.registerVertexCompressor(protocol.VertexLabels.location, locationCompressor)

const projectCompressor = new GenericCompressor<protocol.Project>(vertexCompressor, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('kind', next()),
    GenericCompressorProperty.scalar('resource', next()),
    GenericCompressorProperty.scalar('contents', next()),
])
Compressor.registerVertexCompressor(protocol.VertexLabels.project, projectCompressor)

const documentCompressor = new GenericCompressor<protocol.Document>(vertexCompressor, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('uri', next()),
    GenericCompressorProperty.scalar('languageId', next()),
    GenericCompressorProperty.scalar('contents', next()),
])
Compressor.registerVertexCompressor(protocol.VertexLabels.document, documentCompressor)

const monikerCompressor = new GenericCompressor<protocol.Moniker>(vertexCompressor, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('scheme', next()),
    GenericCompressorProperty.scalar('identifier', next()),
])
Compressor.registerVertexCompressor(protocol.VertexLabels.moniker, monikerCompressor)

const packageInformationCompressor = new GenericCompressor<protocol.PackageInformation>(
    vertexCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('name', next()),
        GenericCompressorProperty.scalar('manager', next()),
        GenericCompressorProperty.scalar('version', next()),
        GenericCompressorProperty.scalar('uri', next()),
        GenericCompressorProperty.scalar('contents', next()),
    ]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.packageInformation, packageInformationCompressor)

export const rangeBasedDocumentSymbolCompressor = new GenericCompressor<protocol.RangeBasedDocumentSymbol>(
    undefined,
    Compressor.nextId(),
    (next, that) => [
        GenericCompressorProperty.scalar('id', next()),
        GenericCompressorProperty.array('children', next(), that),
    ]
)
Compressor.addCompressor(rangeBasedDocumentSymbolCompressor)

const documentSymbolResultCompressor = new GenericCompressor<protocol.DocumentSymbolResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => [
        // @todo find solution for inline lsp.DocumentSymbol.
        GenericCompressorProperty.array('result', next(), rangeBasedDocumentSymbolCompressor),
    ]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.documentSymbolResult, documentSymbolResultCompressor)

const lspLocationCompressor = new GenericCompressor<lsp.Location>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('uri', next()),
    GenericCompressorProperty.literal('range', next(), lspRangeCompressor),
])
Compressor.addCompressor(lspLocationCompressor)

const diagnosticRelatedInformationCompressor = new GenericCompressor<lsp.DiagnosticRelatedInformation>(
    undefined,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.literal('location', next(), lspLocationCompressor),
        GenericCompressorProperty.scalar('message', next()),
    ]
)
Compressor.addCompressor(diagnosticRelatedInformationCompressor)

export const diagnosticCompressor = new GenericCompressor<lsp.Diagnostic>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.literal('range', next(), lspRangeCompressor),
    GenericCompressorProperty.scalar('severity', next()),
    GenericCompressorProperty.scalar('code', next()),
    GenericCompressorProperty.scalar('source', next()),
    GenericCompressorProperty.scalar('message', next()),
    GenericCompressorProperty.literal('relatedInformation', next(), diagnosticRelatedInformationCompressor),
])
Compressor.addCompressor(diagnosticCompressor)

const diagnosticResultCompressor = new GenericCompressor<protocol.DiagnosticResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.array('result', next(), diagnosticCompressor)]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.diagnosticResult, diagnosticResultCompressor)

export const foldingRangeCompressor = new GenericCompressor<lsp.FoldingRange>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('startLine', next()),
    GenericCompressorProperty.scalar('startCharacter', next()),
    GenericCompressorProperty.scalar('endLine', next()),
    GenericCompressorProperty.scalar('endCharacter', next()),
    GenericCompressorProperty.scalar('kind', next()),
])
Compressor.addCompressor(foldingRangeCompressor)

const foldingRangeResultCompressor = new GenericCompressor<protocol.FoldingRangeResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.array('result', next(), foldingRangeCompressor)]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.foldingRangeResult, foldingRangeResultCompressor)

const documentLinkCompressor = new GenericCompressor<lsp.DocumentLink>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.literal('range', next(), lspRangeCompressor),
    GenericCompressorProperty.scalar('target', next()),
    GenericCompressorProperty.raw('data', next()),
])
Compressor.addCompressor(documentLinkCompressor)

const documentLinkResultCompressor = new GenericCompressor<protocol.DocumentLinkResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.array('result', next(), documentLinkCompressor)]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.documentLinkResult, documentLinkResultCompressor)

const declarationResultCompressor = new GenericCompressor<protocol.DeclarationResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerVertexCompressor(protocol.VertexLabels.declarationResult, declarationResultCompressor)

const definitionResultCompressor = new GenericCompressor<protocol.DefinitionResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerVertexCompressor(protocol.VertexLabels.definitionResult, definitionResultCompressor)

const typeDefinitionResultCompressor = new GenericCompressor<protocol.TypeDefinitionResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerVertexCompressor(protocol.VertexLabels.typeDefinitionResult, typeDefinitionResultCompressor)

const markedStringCompressor = new GenericCompressor<{ language: string; value: string }>(
    undefined,
    Compressor.nextId(),
    next => [GenericCompressorProperty.scalar('language', next()), GenericCompressorProperty.scalar('value', next())]
)
Compressor.addCompressor(markedStringCompressor)

const markupContentCompressor = new GenericCompressor<lsp.MarkupContent>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.scalar('kind', next()),
    GenericCompressorProperty.scalar('value', next()),
])
Compressor.addCompressor(markupContentCompressor)

const rawHoverCompressor = new GenericCompressor<lsp.Hover>(undefined, Compressor.nextId(), next => [
    GenericCompressorProperty.any('contents', next(), value => {
        if (Is.string(value)) {
            return noopCompressor
        } else if (lsp.MarkedString.is(value)) {
            return markedStringCompressor
        } else {
            return markupContentCompressor
        }
    }),
    GenericCompressorProperty.literal('range', next(), lspRangeCompressor),
])
Compressor.addCompressor(rawHoverCompressor)

const hoverResultCompressor = new GenericCompressor<protocol.HoverResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.literal('result', next(), rawHoverCompressor)]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.hoverResult, hoverResultCompressor)

const referenceResultCompressor = new GenericCompressor<protocol.ReferenceResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerVertexCompressor(protocol.VertexLabels.referenceResult, referenceResultCompressor)

const implementationResultCompressor = new GenericCompressor<protocol.ImplementationResult>(
    vertexCompressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerVertexCompressor(protocol.VertexLabels.implementationResult, implementationResultCompressor)

const kindShortForm = (function() {
    let shortCounter: number = 1
    return new Map<protocol.EventKind, number>([
        [protocol.EventKind.begin, shortCounter++],
        [protocol.EventKind.end, shortCounter++],
    ])
})()

const scopeShortForm = (function() {
    let shortCounter: number = 1
    return new Map<protocol.EventScope, number>([
        [protocol.EventScope.project, shortCounter++],
        [protocol.EventScope.document, shortCounter++],
    ])
})()

const eventCompressor = new GenericCompressor<protocol.ProjectEvent | protocol.DocumentEvent>(
    vertexCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('kind', next(), kindShortForm),
        GenericCompressorProperty.scalar('scope', next(), scopeShortForm),
        GenericCompressorProperty.scalar('data', next(), scopeShortForm),
    ]
)
Compressor.registerVertexCompressor(protocol.VertexLabels.event, eventCompressor)

export const edgeShortForms = (function() {
    let shortCounter: number = 1
    return new Map<protocol.EdgeLabels, number>([
        [protocol.EdgeLabels.contains, shortCounter++],
        [protocol.EdgeLabels.item, shortCounter++],
        [protocol.EdgeLabels.next, shortCounter++],
        [protocol.EdgeLabels.moniker, shortCounter++],
        [protocol.EdgeLabels.packageInformation, shortCounter++],
        [protocol.EdgeLabels.textDocument_documentSymbol, shortCounter++],
        [protocol.EdgeLabels.textDocument_foldingRange, shortCounter++],
        [protocol.EdgeLabels.textDocument_documentLink, shortCounter++],
        [protocol.EdgeLabels.textDocument_diagnostic, shortCounter++],
        [protocol.EdgeLabels.textDocument_definition, shortCounter++],
        [protocol.EdgeLabels.textDocument_declaration, shortCounter++],
        [protocol.EdgeLabels.textDocument_typeDefinition, shortCounter++],
        [protocol.EdgeLabels.textDocument_hover, shortCounter++],
        [protocol.EdgeLabels.textDocument_references, shortCounter++],
        [protocol.EdgeLabels.textDocument_implementation, shortCounter++],
    ])
})()

export const edge11Compressor = new GenericCompressor<protocol.E11<protocol.V, protocol.V, protocol.EdgeLabels>>(
    elementCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('label', next(), edgeShortForms),
        GenericCompressorProperty.scalar('outV', next()),
        GenericCompressorProperty.scalar('inV', next()),
    ]
)
Compressor.addCompressor(edge11Compressor)

export const edge1nCompressor = new GenericCompressor<protocol.E1N<protocol.V, protocol.V, protocol.EdgeLabels>>(
    elementCompressor,
    Compressor.nextId(),
    next => [
        GenericCompressorProperty.scalar('label', next(), edgeShortForms),
        GenericCompressorProperty.scalar('outV', next()),
        GenericCompressorProperty.array('inVs', next()),
    ]
)
Compressor.addCompressor(edge1nCompressor)

export const containsCompressor = new GenericCompressor<protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>>(
    edge11Compressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerEdgeCompressor(protocol.EdgeLabels.contains, containsCompressor)

export const nextCompressor = new GenericCompressor<protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>>(
    edge11Compressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerEdgeCompressor(protocol.EdgeLabels.next, nextCompressor)

export const monikerEdgeCompressor = new GenericCompressor<protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>>(
    edge11Compressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerEdgeCompressor(protocol.EdgeLabels.moniker, monikerEdgeCompressor)

export const nextMonikerCompressor = new GenericCompressor<protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>>(
    edge11Compressor,
    Compressor.nextId(),
    next => []
)
Compressor.registerEdgeCompressor(protocol.EdgeLabels.nextMoniker, nextMonikerCompressor)

export const packageInformationEdgeCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.packageInformation, packageInformationEdgeCompressor)

export const textDocumentDocumentSymbolCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_documentSymbol, textDocumentDocumentSymbolCompressor)

export const textDocumentFoldingRangeCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_foldingRange, textDocumentFoldingRangeCompressor)

export const textDocumentDocumentLinkCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_documentLink, textDocumentDocumentLinkCompressor)

export const textDocumentDiagnosticCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_diagnostic, textDocumentDiagnosticCompressor)

export const textDocumentDefinitionCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_definition, textDocumentDefinitionCompressor)

export const textDocumentDeclarationCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_declaration, textDocumentDeclarationCompressor)

export const textDocumentTypeDefinitionCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_typeDefinition, textDocumentTypeDefinitionCompressor)

export const textDocumentHoverCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_hover, textDocumentHoverCompressor)

export const textDocumentReferencesCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_references, textDocumentReferencesCompressor)

export const textDocumentImplementationCompressor = new GenericCompressor<
    protocol.E<protocol.V, protocol.V, protocol.EdgeLabels>
>(edge11Compressor, Compressor.nextId(), next => [])
Compressor.registerEdgeCompressor(protocol.EdgeLabels.textDocument_implementation, textDocumentImplementationCompressor)

export const itemPropertyShortForms = (function() {
    let shortCounter: number = 1
    return new Map<protocol.ItemEdgeProperties, number>([
        [protocol.ItemEdgeProperties.declarations, shortCounter++],
        [protocol.ItemEdgeProperties.definitions, shortCounter++],
        [protocol.ItemEdgeProperties.references, shortCounter++],
        [protocol.ItemEdgeProperties.referenceResults, shortCounter++],
    ])
})()

export const itemEdgeCompressor = new GenericCompressor<protocol.ItemEdge<protocol.V, protocol.V>>(
    edge1nCompressor,
    Compressor.nextId(),
    next => [GenericCompressorProperty.scalar('property', next(), itemPropertyShortForms)]
)
Compressor.registerEdgeCompressor(protocol.EdgeLabels.item, itemEdgeCompressor)
