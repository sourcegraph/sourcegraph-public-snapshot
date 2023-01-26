import { DeferredEntity, EntityProvider, EntityProviderConnection } from '@backstage/plugin-catalog-backend';
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth';
export { AuthenticatedUser } from '@sourcegraph/shared/src/auth';
import { Config as Config$1 } from '@backstage/config';

interface Query<T> {
    gql(): string;
    vars(): string;
    Marshal(data: any): T[];
}
interface SearchResult {
    readonly repository: string;
    readonly fileContent: string;
}
declare class SearchQuery implements Query<SearchResult> {
    private readonly query;
    constructor(query: string);
    Marshal(data: any): SearchResult[];
    vars(): any;
    gql(): string;
}
declare class UserQuery implements Query<string> {
    Marshal(data: any): string[];
    vars(): string;
    gql(): string;
}

declare class AuthenticatedUserQuery implements Query<AuthenticatedUser> {
    gql(): string;
    vars(): string;
    Marshal(data: any): AuthenticatedUser[];
}

interface Config {
    endpoint: string;
    token: string;
    sudoUsername?: string;
}
interface UserService {
    CurrentUsername(): Promise<string>;
    GetAuthenticatedUser(): Promise<AuthenticatedUser>;
}
declare const createService: (config: Config) => SourcegraphService;
interface SearchService {
    SearchQuery(query: string): Promise<SearchResult[]>;
}
interface SourcegraphService {
    Users: UserService;
    Search: SearchService;
}

declare const parseCatalog: (src: SearchResult[], providerName: string) => DeferredEntity[];

declare class SourcegraphEntityProvider implements EntityProvider {
    private connection?;
    private readonly sourcegraph;
    static create(config: Config$1): SourcegraphEntityProvider;
    private constructor();
    getProviderName(): string;
    connect(connection: EntityProviderConnection): Promise<void>;
    fullMutation(): Promise<void>;
}

export { AuthenticatedUserQuery, Config, Query, SearchQuery, SearchResult, SearchService, SourcegraphEntityProvider, SourcegraphService, UserQuery, UserService, createService, parseCatalog };
