import { Connection } from 'typeorm'
import { testFilter, createFilter } from './filter'
import * as entities from './entities'
import { connectionCache } from './cache'

export class CorrelationDatabase {
    public constructor(private database: string) {}

    public async getPackage(scheme: string, name: string, version: string): Promise<entities.Package | undefined> {
        return await this.withConnection(connection =>
            connection.getRepository(entities.Package).findOne({
                where: {
                    scheme,
                    name,
                    version,
                },
            })
        )
    }

    public async getReferences(
        scheme: string,
        name: string,
        version: string,
        uri: string
    ): Promise<entities.Reference[]> {
        return await this.withConnection(connection =>
            connection
                .getRepository(entities.Reference)
                .find({
                    where: {
                        scheme,
                        name,
                        version,
                    },
                })
                .then((results: entities.Reference[]) => results.filter(result => testFilter(result.filter, uri)))
        )
    }

    public async addPackage(
        scheme: string,
        name: string,
        version: string,
        repository: string,
        commit: string
    ): Promise<void> {
        return await this.withConnection(async connection => {
            await connection.getRepository(entities.Package).save({ scheme, name, version, repository, commit })
        })
    }

    public async addReference(
        scheme: string,
        name: string,
        version: string,
        repository: string,
        commit: string,
        uris: string[]
    ): Promise<void> {
        return await this.withConnection(async connection => {
            await connection
                .getRepository(entities.Reference)
                .save({ scheme, name, version, repository, commit, filter: createFilter(uris) })
        })
    }

    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await connectionCache.withConnection(this.database, [entities.Package, entities.Reference], callback)
    }
}
