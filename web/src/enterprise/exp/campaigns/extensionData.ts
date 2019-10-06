import { combineLatest, Observable, of } from 'rxjs'
import { debounceTime, map, switchMap, distinctUntilChanged } from 'rxjs/operators'
import { DiagnosticWithType } from '../../../../../shared/src/api/client/services/diagnosticService'
import { fromDiagnostic } from '../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { propertyIsDefined } from '../../../../../shared/src/util/types'
import { RuleDefinition } from '../../rules/types'
import { diagnosticQueryMatcher, getCodeActions } from '../../diagnostics/backend'
import { computeDiff, FileDiff, computeDiffFromEdits } from './backend/computeDiff'
import { WorkspaceEdit } from '../../../../../shared/src/api/types/workspaceEdit'

const getDiagnosticsAndFileDiffs = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rule: RuleDefinition
): Observable<DiagnosticsAndFileDiffs> => {
    switch (rule.type) {
        case 'DiagnosticRule': {
            // TODO!(sqs): handle case when there are no extensions registered that emit matching diagnostics
            const matchesQuery = diagnosticQueryMatcher(rule.query)
            return extensionsController.services.diagnostics
                .observeDiagnostics({}, rule.context || {}, rule.query.type)
                .pipe(
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
                                                            .find(
                                                                a =>
                                                                    a.computeEdit &&
                                                                    a.computeEdit.command === rule.action
                                                            )
                                                      : undefined,
                                          }))
                                      )
                                  )
                              )
                            : of([])
                    ),
                    debounceTime(0),
                    switchMap(async diagnosticsAndActions => {
                        const actionInvocations = diagnosticsAndActions
                            .filter(propertyIsDefined('action'))
                            .map(d => ({
                                actionEditCommand: d.action.computeEdit,
                                diagnostic: fromDiagnostic(d.diagnostic),
                            }))
                            .filter(propertyIsDefined('actionEditCommand'))
                        const fileDiffs = await computeDiff({ extensionsController, actionInvocations })
                        return {
                            diagnostics: diagnosticsAndActions
                                .filter(({ action }) => !action)
                                .map(({ diagnostic }) => diagnostic),
                            fileDiffs,
                        }
                    })
                )
        }

        case 'ActionRule': {
            // Handle the command not being registered initially (e.g., if the extension that registers it is being loaded in the background).
            return extensionsController.services.commands.commands.pipe(
                map(commands => commands.find(c => c.command === rule.action)),
                distinctUntilChanged(),
                switchMap(() =>
                    extensionsController.services.commands.executeCommand({
                        command: rule.action,
                        arguments: [rule.context],
                    })
                ),
                switchMap(async (edit: WorkspaceEdit) => {
                    const fileDiffs = await computeDiffFromEdits(extensionsController, [WorkspaceEdit.fromJSON(edit)])
                    return { fileDiffs, diagnostics: [] }
                })
            )
        }

        default:
            return of({ diagnostics: [], fileDiffs: [] })
    }
}

interface DiagnosticsAndFileDiffs {
    diagnostics: DiagnosticWithType[]
    fileDiffs: Pick<FileDiff, 'patchWithFullURIs'>[]
}

export const getCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rules: RuleDefinition[]
): Observable<GQL.ICreateCampaignInput['extensionData']> =>
    (rules.length > 0
        ? combineLatest(rules.map(rule => getDiagnosticsAndFileDiffs(extensionsController, rule)))
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
