import { DependencyManagementProvider, DependencyQuery } from '.'
import { flatten } from 'lodash'
import * as sourcegraph from 'sourcegraph'
import { Subscription, Unsubscribable, Observable, of, combineLatest } from 'rxjs'
import {
    provideDependencyManagementDiagnostics,
    parseDependencyManagementDiagnostic,
    DependencyManagementDiagnostic,
} from './diagnostics'
import { filter, map } from 'rxjs/operators'
import { LOADING, DependencyManagementCampaignContextCommon } from './common'
import { DependencySpecificationWithType, WithoutType } from './combinedProvider'
import { isDefined } from '../../../../../shared/src/util/types'

export function registerDependencyManagementProviders<
    Q extends DependencyQuery,
    S extends DependencySpecificationWithType<Q>
>(
    id: string,
    provider: WithoutType<DependencyManagementProvider<Q, S>>,
    parseQuery: (context: sourcegraph.ContextValues) => Q
): Unsubscribable {
    const COMMAND_ID = `dependencyManagement.${id}.action`
    const DEPENDENCY_ID = `dependencyManagement.${id}`

    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider(DEPENDENCY_ID, {
            provideDiagnostics: (_scope, context) =>
                provideDependencyManagementDiagnostics(
                    provider,
                    DEPENDENCY_ID,
                    parseQuery(context),
                    (context as unknown) as DependencyManagementCampaignContextCommon
                ),
        })
    )
    subscriptions.add(
        sourcegraph.languages.registerCodeActionProvider(['*'], {
            provideCodeActions: (_doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> =>
                combineLatest(
                    context.diagnostics
                        .map(diagnostic => parseDependencyManagementDiagnostic(diagnostic, DEPENDENCY_ID))
                        .filter(isDefined)
                        .map(diagnostic =>
                            editForDependencyAction(provider, diagnostic).pipe(
                                map(edit => {
                                    const action: sourcegraph.Action = {
                                        title: 'Upgrade dependency',
                                        // edit,
                                        computeEdit: { title: 'Upgrade dependency', command: COMMAND_ID },
                                        diagnostics: [diagnostic],
                                    }
                                    return [action]
                                })
                            )
                        )
                ).pipe(map(allActions => flatten(allActions))),
        })
    )
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(COMMAND_ID, diagnostic => {
            if (!diagnostic) {
                return Promise.resolve(new sourcegraph.WorkspaceEdit())
            }
            const parsed = parseDependencyManagementDiagnostic(diagnostic, DEPENDENCY_ID)
            // TODO!(sqs): dont create changesets unless requested in campaign context
            if (!parsed) {
                return Promise.resolve(new sourcegraph.WorkspaceEdit())
            }
            return editForDependencyAction(provider, parsed).toPromise()
        })
    )
    return subscriptions
}

function editForDependencyAction<Q extends DependencyQuery>(
    provider: WithoutType<DependencyManagementProvider<Q>>,
    { parsedData }: DependencyManagementDiagnostic<Q>
): Observable<sourcegraph.WorkspaceEdit> {
    if (!parsedData.createChangesets) {
        return of(new sourcegraph.WorkspaceEdit())
    }
    if (parsedData.action === 'ban') {
        return provider.resolveDependencyBanAction
            ? provider.resolveDependencyBanAction(parsedData)
            : of(new sourcegraph.WorkspaceEdit())
    }
    return provider.resolveDependencyUpgradeAction
        ? provider.resolveDependencyUpgradeAction(parsedData, parsedData.action.requireVersion)
        : of(new sourcegraph.WorkspaceEdit())
}
