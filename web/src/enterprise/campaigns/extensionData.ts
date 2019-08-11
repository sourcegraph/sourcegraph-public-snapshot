import { combineLatest, Observable, of } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { fromDiagnostic } from '../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { propertyIsDefined } from '../../../../shared/src/util/types'
import { RuleDefinition } from '../rules/types'
import {
    DiagnosticInfo,
    diagnosticQueryMatcher,
    getCodeActions,
    getDiagnosticInfos,
} from '../threadsOLD/detail/backend'
import { computeDiff, FileDiff } from '../threadsOLD/detail/changes/computeDiff'

const getDiagnosticsAndFileDiffs = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rule: RuleDefinition
): Observable<DiagnosticsAndFileDiffs> => {
    if (rule.type !== 'DiagnosticRule') {
        return of({ diagnostics: [], fileDiffs: [] })
    }
    const matchesQuery = diagnosticQueryMatcher(rule.query)
    return getDiagnosticInfos(extensionsController, rule.query.type).pipe(
        map(diagnostics => diagnostics.filter(matchesQuery)),
        switchMap(diagnostics =>
            diagnostics.length > 0
                ? combineLatest(
                      diagnostics.map(d =>
                          getCodeActions({
                              diagnostic: d,
                              extensionsController,
                          }).pipe(
                              map(actions => ({
                                  diagnostic: d,
                                  action:
                                      rule.action !== undefined
                                          ? actions
                                                .filter(propertyIsDefined('computeEdit'))
                                                .find(a => a.computeEdit && a.computeEdit.command === rule.action)
                                          : undefined,
                              }))
                          )
                      )
                  )
                : of([])
        ),
        switchMap(async diagnosticsAndActions => {
            const fileDiffs = await computeDiff(
                extensionsController,
                diagnosticsAndActions
                    .filter(propertyIsDefined('action'))
                    .map(d => ({
                        actionEditCommand: d.action.computeEdit,
                        diagnostic: fromDiagnostic(d.diagnostic),
                    }))
                    .filter(propertyIsDefined('actionEditCommand'))
            )
            return {
                diagnostics: diagnosticsAndActions.filter(({ action }) => !action).map(({ diagnostic }) => diagnostic),
                fileDiffs,
            }
        })
    )
}

interface DiagnosticsAndFileDiffs {
    diagnostics: DiagnosticInfo[]
    fileDiffs: Pick<FileDiff, 'patchWithFullURIs'>[]
}

export const getCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    input: Pick<GQL.ICreateCampaignInput, 'rules'>
): Observable<GQL.ICreateCampaignInput['extensionData']> =>
    (input.rules && input.rules.length > 0
        ? combineLatest(
              (input.rules || []).map(rule => {
                  const def: RuleDefinition = JSON.parse(rule.definition)
                  return getDiagnosticsAndFileDiffs(extensionsController, def)
              })
          )
        : of<DiagnosticsAndFileDiffs[]>([{ diagnostics: [], fileDiffs: [] }])
    ).pipe(
        map(results => {
            const combined: DiagnosticsAndFileDiffs = {
                diagnostics: results.flatMap(r => r.diagnostics),
                fileDiffs: results.flatMap(r => r.fileDiffs),
            }
            return combined
        }),
        map(({ diagnostics, fileDiffs }) => ({
            rawDiagnostics: diagnostics.map(d =>
                // tslint:disable-next-line: no-object-literal-type-assertion
                JSON.stringify({
                    __typename: 'Diagnostic',
                    type: d.type,
                    data: d,
                } as GQL.IDiagnostic)
            ),
            rawFileDiffs: fileDiffs.map(({ patchWithFullURIs }) => patchWithFullURIs),
        }))
    )
