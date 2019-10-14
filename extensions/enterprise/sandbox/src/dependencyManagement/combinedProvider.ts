import { combineLatest } from 'rxjs'
import { map } from 'rxjs/operators'
import { flatten } from 'lodash'
import { DependencyManagementProvider, DependencyQuery, DependencySpecification } from '.'
import { propertyIsDefined } from '../../../../../shared/src/util/types'
import { combineWorkspaceEdits } from '../../../../../shared/src/api/types/workspaceEdit'

export interface DependencySpecificationWithType<Q extends DependencyQuery> extends DependencySpecification<Q> {
    type: string
}

export const combinedProvider = <Q extends DependencyQuery>(
    providers: (DependencyManagementProvider<Q> & { type: string })[]
): Required<DependencyManagementProvider<Q, DependencySpecificationWithType<Q>>> => ({
    provideDependencySpecifications: (query, filters) =>
        combineLatest(
            providers.map(provider =>
                provider
                    .provideDependencySpecifications(query, filters)
                    .pipe(
                        map(specs =>
                            specs.map<DependencySpecificationWithType<Q>>(spec => ({ ...spec, type: provider.type }))
                        )
                    )
            )
        ).pipe(map(allSpecs => flatten(allSpecs))),
    resolveDependencyUpgradeAction: (dep, version) =>
        combineLatest(
            providers
                .filter(provider => provider.type === dep.type)
                .filter(propertyIsDefined('resolveDependencyUpgradeAction'))
                .map(provider => provider.resolveDependencyUpgradeAction(dep, version))
        ).pipe(map(results => combineWorkspaceEdits(results))),
    resolveDependencyBanAction: dep =>
        combineLatest(
            providers
                .filter(provider => provider.type === dep.type)
                .filter(propertyIsDefined('resolveDependencyBanAction'))
                .map(provider => provider.resolveDependencyBanAction(dep))
        ).pipe(map(results => combineWorkspaceEdits(results))),
})
