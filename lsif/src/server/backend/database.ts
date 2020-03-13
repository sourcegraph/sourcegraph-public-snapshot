import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'
import got from 'got'
import * as pgModels from '../../shared/models/pg'
import { InternalLocation, OrderedLocationSet } from './location'
import { TracingContext } from '../../shared/tracing'
import { uniqWith, isEqual } from 'lodash'
import { parseJSON } from '../../shared/encoding/json'

/** A wrapper around operations related to a single SQLite dump. */
export class Database {
    constructor(private dump: pgModels.LsifDump) {}

    /**
     * Retrieve all document paths from the database.
     *
     * @param ctx The tracing context.
     */
    public documentPaths(ctx: TracingContext = {}): Promise<string[]> {
        return this.request('documentPaths', new URLSearchParams(), ctx)
    }

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public exists(path: string, ctx: TracingContext = {}): Promise<boolean> {
        return this.request('exists', new URLSearchParams({ path }), ctx)
    }

    /**
     * Return a list of locations that define the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async definitions(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        return this.request(
            'definitions',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )
    }

    /**
     * Return a list of unique locations that reference the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async references(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<OrderedLocationSet> {
        const locations = new OrderedLocationSet()
        for (const location of await this.request<InternalLocation[]>(
            'references',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )) {
            locations.push(location)
        }

        return locations
    }

    /**
     * Return the hover content for the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async hover(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<{ text: string; range: lsp.Range } | null> {
        return this.request(
            'hover',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )
    }

    /**
     * Return a parsed document that describes the given path as well as the ranges
     * from that document that contains the given position. If multiple ranges are
     * returned, then the inner-most ranges will occur before the outer-most ranges.
     *
     * @param path The path of the document.
     * @param position The user's hover position.
     * @param ctx The tracing context.
     */
    public getRangeByPosition(
        path: string,
        position: lsp.Position,
        ctx: TracingContext
    ): Promise<{ document: sqliteModels.DocumentData | undefined; ranges: sqliteModels.RangeData[] }> {
        return this.request(
            'getRangeByPosition',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )
    }

    /**
     * Query the definitions or references table of `db` for items that match the given moniker.
     * Convert each result into an `InternalLocation`. The `pathTransformer` function is invoked
     * on each result item to modify the resulting locations.
     *
     * @param model The constructor for the model type.
     * @param moniker The target moniker.
     * @param pagination A limit and offset to use for the query.
     * @param ctx The tracing context.
     */
    public monikerResults(
        model: typeof sqliteModels.DefinitionModel | typeof sqliteModels.ReferenceModel,
        moniker: Pick<sqliteModels.MonikerData, 'scheme' | 'identifier'>,
        pagination: { skip?: number; take?: number },
        ctx: TracingContext
    ): Promise<{ locations: InternalLocation[]; count: number }> {
        let p: {} | { skip: string } | { take: string } | { skip: string; take: string } = {}
        if (pagination.skip !== undefined) {
            p = { skip: String(pagination.skip) }
        }
        if (pagination.take !== undefined) {
            p = { ...p, take: String(pagination.take) }
        }

        return this.request(
            'monikerResults',
            new URLSearchParams({
                modelType: model === sqliteModels.DefinitionModel ? 'definition' : 'reference',
                scheme: moniker.scheme,
                identifier: moniker.identifier,
                ...p,
            }),
            ctx
        )
    }

    /**
     * Return a parsed document that describes the given path. The result of this
     * method is cached across all database instances. If the document is not found
     * it returns undefined; other errors will throw.
     *
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public getDocumentByPath(path: string, ctx: TracingContext = {}): Promise<sqliteModels.DocumentData | undefined> {
        return this.request('getDocumentByPath', new URLSearchParams({ path }), ctx)
    }

    private async request<T>(method: string, searchParams: URLSearchParams, ctx: TracingContext): Promise<T> {
        const url = new URL(`http://localhost:3188/${this.dump.id}/${method}`)
        url.search = searchParams.toString()
        const resp = await got.get(url.href)
        return parseJSON(resp.body)
    }
}

// The order to present monikers in when organized by kinds
const monikerKindPreferences = ['import', 'local', 'export']

// A map from moniker schemes to schemes that subsume them. The schemes
// identified by keys should be removed from the sets of monikers that
// also contain the scheme identified by that key's value.
const subsumedMonikers = new Map([
    ['go', 'gomod'],
    ['tsc', 'npm'],
])

/**
 * Normalize the set of monikers by filtering, sorting, and removing
 * duplicates from the list based on the moniker kind and scheme values.
 *
 * @param monikers The list of monikers.
 */
export function sortMonikers(monikers: sqliteModels.MonikerData[]): sqliteModels.MonikerData[] {
    // Deduplicate monikers. This can happen with long chains of result
    // sets where monikers are applied several times to an aliased symbol.
    monikers = uniqWith(monikers, isEqual)

    // Remove monikers subsumed by the presence of another. For example,
    // if we have an `npm` moniker in this list, we want to remove all
    // `tsc` monikers as they are duplicate by construction in lsif-tsc.
    monikers = monikers.filter(a => {
        const by = subsumedMonikers.get(a.scheme)
        return !(by && monikers.some(b => b.scheme === by))
    })

    // Sort monikers by kind
    monikers.sort((a, b) => monikerKindPreferences.indexOf(a.kind) - monikerKindPreferences.indexOf(b.kind))
    return monikers
}
