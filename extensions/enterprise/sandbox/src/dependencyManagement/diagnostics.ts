import * as sourcegraph from 'sourcegraph'
import { DependencySpecificationWithType } from './combinedProvider'
import { DependencyManagementCampaignContextCommon, LOADING } from './common'
import { DependencyQuery, DependencyManagementProvider } from '.'
import { Observable, from, of } from 'rxjs'
import { startWith, map, switchMap } from 'rxjs/operators'
import { isDefined } from '../../../../../shared/src/util/types'
import { toLocation } from '../../../../../shared/src/api/extension/api/types'

export interface DependencyManagementDiagnosticData<Q extends DependencyQuery>
    extends DependencySpecificationWithType<Q>,
        Pick<DependencyManagementCampaignContextCommon, 'action' | 'createChangesets'> {}

export interface DependencyManagementDiagnostic<Q extends DependencyQuery> extends sourcegraph.Diagnostic {
    parsedData: DependencyManagementDiagnosticData<Q>
}

export const parseDependencyManagementDiagnostic = <Q extends DependencyQuery>(
    diagnostic: sourcegraph.Diagnostic,
    tag: string
): DependencyManagementDiagnostic<Q> | null => {
    if (!diagnostic.data) {
        return null
    }
    if (!diagnostic.tags || !diagnostic.tags.includes(tag)) {
        return null
    }

    const parsed: DependencyManagementDiagnosticData<Q> = JSON.parse(diagnostic.data) // URL objects are stringified
    const converted: DependencyManagementDiagnosticData<Q> = {
        ...parsed,
        declarations: parsed.declarations.map(d => ({ ...d, location: toLocation(d.location as any) })),
        resolutions: parsed.resolutions.map(r => ({
            ...r,
            location: r.location ? toLocation(r.location as any) : undefined,
        })),
    }
    return { ...diagnostic, parsedData: converted }
}

export const provideDependencyManagementDiagnostics = <
    Q extends DependencyQuery,
    S extends DependencySpecificationWithType<Q> = DependencySpecificationWithType<Q>
>(
    { provideDependencySpecifications }: Pick<DependencyManagementProvider<Q, S>, 'provideDependencySpecifications'>,
    dependencyTag: string,
    query: Q,
    {
        action,
        createChangesets,
        filters,
    }: Pick<DependencyManagementCampaignContextCommon, 'action' | 'createChangesets' | 'filters'>
): Observable<sourcegraph.Diagnostic[] | typeof LOADING> =>
    from(sourcegraph.workspace.rootChanges).pipe(
        startWith(undefined),
        map(() => sourcegraph.workspace.roots),
        switchMap(roots => {
            if (roots.length > 0) {
                return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
            }

            const specs = provideDependencySpecifications(query, filters)
            return specs.pipe(
                map(specs =>
                    specs
                        .map(spec => {
                            if (spec.error) {
                                console.error(spec.error)
                                return null
                            }
                            const specMain = spec.declarations[0]
                                ? spec.declarations[0]
                                : { ...spec.resolutions[0], direct: false }
                            if (!specMain.location) {
                                return null
                            }
                            const data: DependencyManagementDiagnosticData<Q> = { ...spec, action, createChangesets }
                            const diagnostic: sourcegraph.Diagnostic = {
                                resource: specMain.location.uri,
                                message: `${specMain.direct ? '' : 'Indirect '}npm dependency ${specMain.name}${
                                    query.versionRange === '*' ? '' : `@${query.versionRange}`
                                } ${action === 'ban' ? 'is banned' : `must be upgraded to ${action.requireVersion}`}`,
                                range: specMain.location.range || new sourcegraph.Range(0, 0, 0, 0),
                                severity: sourcegraph.DiagnosticSeverity.Warning,
                                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                data: JSON.stringify(data),
                                tags: [dependencyTag],
                            }
                            return diagnostic
                        })
                        .filter(isDefined)
                )
            )
        }),
        startWith(LOADING)
    )
