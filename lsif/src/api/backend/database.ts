import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'
import * as pgModels from '../../shared/models/pg'
import { TracingContext } from '../../shared/tracing'
import { parseJSON } from '../../shared/encoding/json'
import * as settings from '../settings'
import got from 'got'
import { InternalLocation, OrderedLocationSet } from './location'

/** A wrapper around operations related to a single SQLite dump. */
export class Database {
    constructor(private dumpId: pgModels.DumpId) {}

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
        const locations = await this.request<{ path: string; range: lsp.Range }[]>(
            'definitions',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )

        return locations.map(location => ({ ...location, dumpId: this.dumpId }))
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
        const locations = await this.request<{ path: string; range: lsp.Range }[]>(
            'references',
            new URLSearchParams({ path, line: String(position.line), character: String(position.character) }),
            ctx
        )

        return new OrderedLocationSet(locations.map(location => ({ ...location, dumpId: this.dumpId })))
    }

    /**
     * Return the hover content for the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public hover(
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
     * Return all of the monikers attached to all ranges that contain the given position. The
     * resulting list is grouped by range. If multiple ranges contain this position, then the
     * list monikers for the inner-most ranges will occur before the outer-most ranges.
     *
     * @param path The path of the document.
     * @param position The user's hover position.
     * @param ctx The tracing context.
     */
    public monikersByPosition(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<sqliteModels.MonikerData[][]> {
        return this.request(
            'monikersByPosition',
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
    public async monikerResults(
        model: typeof sqliteModels.DefinitionModel | typeof sqliteModels.ReferenceModel,
        moniker: Pick<sqliteModels.MonikerData, 'scheme' | 'identifier'>,
        pagination: { skip?: number; take?: number },
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; count: number }> {
        let p: {} | { skip: string } | { take: string } | { skip: string; take: string } = {}
        if (pagination.skip !== undefined) {
            p = { ...p, skip: String(pagination.skip) }
        }
        if (pagination.take !== undefined) {
            p = { ...p, take: String(pagination.take) }
        }

        const { locations, count } = await this.request<{
            locations: { path: string; range: lsp.Range }[]
            count: number
        }>(
            'monikerResults',
            new URLSearchParams({
                modelType: model === sqliteModels.DefinitionModel ? 'definition' : 'reference',
                scheme: moniker.scheme,
                identifier: moniker.identifier,
                ...p,
            }),
            ctx
        )

        return { locations: locations.map(location => ({ ...location, dumpId: this.dumpId })), count }
    }

    /**
     * Return the package information data with the given identifier.
     *
     * @param path The path of the document.
     * @param packageInformationId The identifier of the package information data.
     * @param ctx The tracing context.
     */
    public packageInformation(
        path: string,
        packageInformationId: sqliteModels.PackageInformationId,
        ctx: TracingContext = {}
    ): Promise<sqliteModels.PackageInformationData | undefined> {
        return this.request(
            'packageInformation',
            new URLSearchParams({ path, packageInformationId: String(packageInformationId) }),
            ctx
        )
    }

    //
    //

    private async request<T>(method: string, searchParams: URLSearchParams, ctx: TracingContext): Promise<T> {
        const url = new URL(`/dbs/${this.dumpId}/${method}`, settings.LSIF_DUMP_MANAGER_URL)
        url.search = searchParams.toString()
        const resp = await got.get(url.href)
        return parseJSON(resp.body)
    }
}
