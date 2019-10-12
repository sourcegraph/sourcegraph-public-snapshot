import { isEqual } from 'lodash'
import { combineLatest, Observable, of } from 'rxjs'
import { catchError, distinctUntilChanged, first, map, startWith, switchMap, throttleTime } from 'rxjs/operators'
import { CodeActionError, isCodeActionError } from '../../../../shared/src/api/client/services/codeActions'
import { DiagnosticWithType } from '../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../shared/src/api/types/action'
import { fromDiagnostic } from '../../../../shared/src/api/types/diagnostic'
import { WorkspaceEdit } from '../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../shared/src/util/strings'
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
    isLoading: boolean
    progress?: readonly [number, number]
    errors?: string[]
    messages?: string[]
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
                    distinctUntilChanged((a, b) => isEqual(a, b)),
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

                        const loadingCount = diagnosticsAndActions.filter(d => d.action === LOADING).length
                        const totalCount = diagnosticsAndActions.filter(d => d.action !== undefined).length
                        const allErrors = diagnosticsAndActions.flatMap(da => da.errors)
                        const errorsFoundStr =
                            allErrors.length > 0
                                ? `${allErrors.length} ${pluralize('error', allErrors.length)} occurred`
                                : ''

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
                            status: {
                                isLoading: loadingCount > 0,
                                progress: [totalCount - loadingCount, totalCount] as const,
                                messages: [
                                    loadingCount > 0
                                        ? `Generating fixes... ${errorsFoundStr ? `(${errorsFoundStr})` : ''}`
                                        : errorsFoundStr,
                                ],
                                errors: allErrors.map(e => e.message),
                            },
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
                    const v: DiagnosticsAndFileDiffs = {
                        fileDiffs,
                        diagnostics: [],
                        status: { isLoading: false },
                    }
                    return v
                }),
                catchError(err => [
                    {
                        fileDiffs: [],
                        diagnostics: [],
                        status: { isLoading: false, errors: [err.message || 'Unknown error'] },
                    },
                ]),
                startWith<DiagnosticsAndFileDiffs>({
                    fileDiffs: [],
                    diagnostics: [],
                    status: { isLoading: true, messages: [`Running ${rule.action} to generate fixes...`] },
                })
            )
        }

        default:
            return of<DiagnosticsAndFileDiffs>({
                diagnostics: [],
                fileDiffs: [],
                status: { isLoading: true, messages: ['Waiting...'] },
            })
    }
}

const sum = (nums: number[]): number => nums.reduce((sum, n) => sum + n, 0)

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
                      status: {
                          isLoading: results.some(r => r.status.isLoading),
                          progress: [
                              sum(results.map(r => (r.status.progress ? r.status.progress[0] : 0))),
                              sum(results.map(r => (r.status.progress ? r.status.progress[1] : 0))),
                          ],
                          errors: results.flatMap(r => r.status.errors || []).filter(e => !!e),
                          messages: results.flatMap(r => r.status.messages || []),
                      },
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
              {
                  isLoading: false,
                  errors: ['No campaign extensions are active. Enable extensions to create a campaign.'],
              },
          ])

/** Waits for `getCampaignExtensionData` to finish loading and returns the result. */
export const getCompleteCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rules: RuleDefinition[]
): Promise<GQL.IExpCreateCampaignInput['extensionData']> =>
    getCampaignExtensionData(extensionsController, rules)
        .pipe(
            first(([, status]) => !status.isLoading),
            map(([extensionData]) => extensionData)
        )
        .toPromise()
