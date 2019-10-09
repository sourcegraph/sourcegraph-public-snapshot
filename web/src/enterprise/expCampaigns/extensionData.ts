import { combineLatest, Observable, of } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap, throttleTime } from 'rxjs/operators'
import { CodeActionError, isCodeActionError } from '../../../../shared/src/api/client/services/codeActions'
import { DiagnosticWithType } from '../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../shared/src/api/types/action'
import { fromDiagnostic } from '../../../../shared/src/api/types/diagnostic'
import { WorkspaceEdit } from '../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { propertyIsDefined } from '../../../../shared/src/util/types'
import { diagnosticQueryMatcher, getCodeActions } from '../diagnostics/backend'
import { RuleDefinition } from '../rules/types'
import { computeDiff, computeDiffFromEdits, FileDiff } from './backend/computeDiff'

interface DiagnosticsAndFileDiffs {
    diagnostics: DiagnosticWithType[]
    fileDiffs: Pick<FileDiff, 'patchWithFullURIs'>[]
    status: ExtensionDataStatus
}

export interface ExtensionDataStatus {
    message: string
}

const LOADING = 'loading' as const

const getDiagnosticsAndFileDiffs = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rule: RuleDefinition
): Observable<DiagnosticsAndFileDiffs> => {
    interface DiagnosticWithAction {
        diagnostic: DiagnosticWithType
        errors: CodeActionError[]
        action: Action | typeof LOADING | undefined
    }
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
                                              errors: actions.filter(isCodeActionError),
                                              action:
                                                  rule.action !== undefined
                                                      ? actions
                                                            .filter((a): a is Action => !isCodeActionError(a))
                                                            .filter(propertyIsDefined('computeEdit'))
                                                            .find(
                                                                a =>
                                                                    a.computeEdit &&
                                                                    a.computeEdit.command === rule.action
                                                            )
                                                      : undefined,
                                          })),
                                          startWith<DiagnosticWithAction>({
                                              diagnostic: d,
                                              errors: [],
                                              action: LOADING,
                                          })
                                      )
                                  )
                              )
                            : of([])
                    ),
                    throttleTime(2500, undefined, { leading: true, trailing: true }),
                    switchMap(async diagnosticsAndActions => {
                        const actionInvocations = diagnosticsAndActions
                            .filter(propertyIsDefined('action'))
                            .map(d => ({
                                actionEditCommand: d.action !== LOADING ? d.action.computeEdit : undefined,
                                diagnostic: fromDiagnostic(d.diagnostic),
                            }))
                            .filter(propertyIsDefined('actionEditCommand'))
                        const fileDiffs = await computeDiff({ extensionsController, actionInvocations })

                        const withExtraDetail = (
                            diagnostic: DiagnosticWithType,
                            detail: string
                        ): DiagnosticWithType => ({
                            ...diagnostic,
                            detail: `${diagnostic.detail ? `${diagnostic.detail} - ` : ''}${detail}`,
                        })
                        return {
                            diagnostics: diagnosticsAndActions
                                // .filter(({ action }) => action)
                                .map(({ diagnostic, action, errors }) =>
                                    errors.length === 0
                                        ? action === LOADING
                                            ? withExtraDetail(diagnostic, 'Loading fix...')
                                            : diagnostic
                                        : withExtraDetail(diagnostic, errors.map(e => e.message).join(', '))
                                ),
                            fileDiffs,
                            status: { message: diagnosticsAndActions.flatMap(da => da.errors).join(', ') },
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
                    return { fileDiffs, diagnostics: [], status: { message: 'OK' } }
                }),
                catchError(err => [{ fileDiffs: [], diagnostics: [], status: { message: err.message } }]),
                startWith({ fileDiffs: [], diagnostics: [], status: { message: `Running ${rule.action}...` } })
            )
        }

        default:
            return of({ diagnostics: [], fileDiffs: [], status: { message: 'Waiting...' } })
    }
}

export const getCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rules: RuleDefinition[]
): Observable<[GQL.IExpCreateCampaignInput['extensionData'], ExtensionDataStatus]> =>
    rules.length > 0
        ? combineLatest(rules.map(rule => getDiagnosticsAndFileDiffs(extensionsController, rule))).pipe(
              map(results => {
                  const combined: DiagnosticsAndFileDiffs = {
                      diagnostics: results.flatMap(r => r.diagnostics),
                      fileDiffs: results.flatMap(r => r.fileDiffs),
                      status: { message: results.map(r => r.status.message).join(', ') },
                  }
                  return combined
              }),
              map(({ diagnostics, fileDiffs, status }) => [
                  {
                      rawDiagnostics: diagnostics.map(d =>
                          // tslint:disable-next-line: no-object-literal-type-assertion
                          JSON.stringify({
                              __typename: 'Diagnostic',
                              type: d.type,
                              data: d,
                          } as GQL.IDiagnostic)
                      ),
                      rawFileDiffs: fileDiffs.map(({ patchWithFullURIs }) => patchWithFullURIs),
                  },
                  status,
              ])
          )
        : of([
              { rawDiagnostics: [], rawFileDiffs: [] },
              { message: 'No campaign extensions are active. Enable extensions to create a campaign.' },
          ])
