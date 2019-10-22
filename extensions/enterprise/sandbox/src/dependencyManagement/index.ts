import { Observable } from 'rxjs'
import { Location, WorkspaceEdit, Diagnostic } from 'sourcegraph'
import { ErrorLike } from '../../../../../shared/src/util/errors'

export interface DependencyQuery {
    readonly name: string
    readonly versionRange?: string
}

export interface DependencyDeclaration {
    /** The name of the dependency. */
    readonly name: string

    /** The requested version of the dependency. */
    readonly requestedVersion?: string

    /** Whether the dependency was declared directly. */
    readonly direct: boolean

    /** The location where the dependency is declared. */
    readonly location: Location
}

export interface DependencyResolution {
    /** The name of the dependency. */
    readonly name: string

    /** The resolved version of the dependency. */
    readonly resolvedVersion: string

    /** The location where the dependency resolution is recorded. */
    readonly location?: Location
}

/**
 * The resolved location(s) where a dependency is specified.
 */
export interface DependencySpecification<Q extends DependencyQuery> {
    /** The original query that was resolved. */
    readonly query: Q

    /** The locations where the dependency is declared (directly or indirectly). */
    readonly declarations: readonly DependencyDeclaration[]

    /** The locations where the dependency is resolved. */
    readonly resolutions: readonly DependencyResolution[]

    /** Diagnostics with information, warnings, or errors related to resolving this dependency. */
    readonly diagnostics?: Diagnostic[]
}

export interface DependencyManagementProvider<
    Q extends DependencyQuery,
    S extends DependencySpecification<Q> = DependencySpecification<Q>
> {
    type: string
    provideDependencySpecifications(dep: Q, filters?: string): Observable<readonly S[]>
    resolveDependencyUpgradeAction?(dep: S, version: string): Observable<WorkspaceEdit>
    resolveDependencyBanAction?(dep: S): Observable<WorkspaceEdit>
}
