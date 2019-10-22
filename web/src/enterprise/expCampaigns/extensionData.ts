import { isEqual } from 'lodash'
import { combineLatest, Observable, of, EMPTY, from } from 'rxjs'
import {
    distinctUntilChanged,
    first,
    map,
    startWith,
    switchMap,
    throttleTime,
    defaultIfEmpty,
    tap,
    catchError,
    last,
    filter,
} from 'rxjs/operators'
import { CodeActionError, isCodeActionError } from '../../../../shared/src/api/client/services/codeActions'
import { DiagnosticWithType } from '../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../shared/src/api/types/action'
import { fromDiagnostic } from '../../../../shared/src/api/types/diagnostic'
import {
    WorkspaceEdit,
    combineWorkspaceEdits,
    SerializedWorkspaceEdit,
} from '../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../shared/src/util/strings'
import { propertyIsDefined, isDefined } from '../../../../shared/src/util/types'
import { getCodeActions, diagnosticID } from '../diagnostics/backend'
import { WorkflowRun, Workflow, Command } from '../../schema/workflow.schema'
import { Changeset } from '../../../../extensions/enterprise/sandbox/src/workflow/behaviors/edits'
import { computeDiffFromEdits } from './backend/computeDiff'
import { Diagnostic } from 'sourcegraph'
import { ErrorLike, asError, isErrorLike } from '../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'

interface DiagnosticsAndFileDiffs {
    diagnostics: Diagnostic[]
    edits: WorkspaceEdit[]
    status: ExtensionDataStatus
}

export interface ExtensionDataStatus {
    isLoading: boolean
    progress?: readonly [number, number]
    errors?: string[]
    messages?: string[]
}

const LOADING = 'loading' as const

const withExtraDetail = (diagnostic: DiagnosticWithType, detail: string): DiagnosticWithType => ({
    ...diagnostic,
    detail: `${diagnostic.detail ? `${diagnostic.detail} - ` : ''}${detail}`,
})

interface ActionWithResult {
    action: Action
    result: SerializedWorkspaceEdit | ErrorLike | typeof LOADING
}

interface DiagnosticWithAction {
    diagnostic: DiagnosticWithType
    errors: CodeActionError[]
    action: ActionWithResult | typeof LOADING | undefined
}

const executeDiagnosticCodeAction = memoizeObservable(
    ({
        extensionsController,
        diagnostic,
        codeActions,
    }: ExtensionsControllerProps & {
        diagnostic: DiagnosticWithType
        codeActions: Command[]
    }): Observable<DiagnosticWithAction> =>
        getCodeActions({ diagnostic, extensionsController }).pipe(
            switchMap(actions => {
                const v: DiagnosticWithAction = {
                    diagnostic,
                    errors: actions.filter(isCodeActionError),
                    action: undefined,
                }

                const action = actions
                    .filter((a): a is Action => !isCodeActionError(a))
                    .filter(propertyIsDefined('computeEdit'))
                    .find(({ computeEdit }) => codeActions.some(a => a.command === computeEdit.command))

                return action
                    ? from(
                          extensionsController.services.commands.executeActionEditCommand(
                              fromDiagnostic(diagnostic),
                              action.computeEdit
                          )
                      ).pipe(
                          catchError(err => of<ErrorLike>(asError(err))),
                          map(
                              result => ({ ...v, action: { action, result } }),
                              startWith({ ...v, action: { action, result: LOADING } })
                          )
                      )
                    : of(v)
            }),
            startWith<DiagnosticWithAction>({
                diagnostic,
                errors: [],
                action: LOADING,
            })
        ),
    ({ diagnostic, codeActions }) => `${diagnosticID(diagnostic)}:${JSON.stringify(codeActions)}`,
    result => result.pipe(last())
)

const executeRun = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    {
        variables: runVariables,
        diagnostics: runDiagnostics,
        codeActions: runCodeActions,
        commands: runCommands,
    }: WorkflowRun
): Observable<DiagnosticsAndFileDiffs> => {
    if (runCommands && runCommands.length > 0) {
        const command = runCommands[0] // TODO!(sqs)
        return extensionsController.services.commands.commands.pipe(
            filter(commands => !!commands.find(c => c.command === command.command)),
            switchMap(() =>
                from(
                    extensionsController.services.commands.executeActionEditCommand(null, {
                        ...command,
                        arguments: [runVariables, ...(command.arguments || [])],
                    })
                )
            ),
            map(result => {
                const v: DiagnosticsAndFileDiffs = {
                    diagnostics: [],
                    edits: [WorkspaceEdit.fromJSON(result)],
                    status: { isLoading: false },
                }
                return v
            })
        )
    }
    const diagnostics = runDiagnostics
        ? extensionsController.services.diagnostics.observeDiagnostics({}, runVariables || {}, runDiagnostics)
        : EMPTY
    const codeActions = runCodeActions
        ? diagnostics.pipe(
              defaultIfEmpty<DiagnosticWithType[]>([]),
              switchMap(diagnostics =>
                  diagnostics.length > 0
                      ? combineLatest(
                            diagnostics.map(diagnostic =>
                                executeDiagnosticCodeAction({
                                    extensionsController,
                                    diagnostic,
                                    codeActions: runCodeActions,
                                })
                            )
                        )
                      : EMPTY
              ),
              throttleTime(1500, undefined, { leading: true, trailing: true })
          )
        : EMPTY
    return codeActions.pipe(
        defaultIfEmpty<DiagnosticWithAction[]>([]),
        map(diagnosticsAndActions => {
            const loadingCount = diagnosticsAndActions.filter(
                d =>
                    (d.action !== undefined && (d.action === LOADING || d.action.result === LOADING)) ||
                    (d.diagnostic.tags && d.diagnostic.tags.includes(LOADING))
            ).length
            const totalCount = diagnosticsAndActions.length
            const allErrors = [
                ...diagnosticsAndActions.flatMap(d => d.errors),
                ...diagnosticsAndActions
                    .map(d =>
                        d.action !== undefined && d.action !== LOADING && isErrorLike(d.action.result)
                            ? d.action.result
                            : undefined
                    )
                    .filter(isDefined),
            ]
            const errorsFoundStr =
                allErrors.length > 0 ? `${allErrors.length} ${pluralize('error', allErrors.length)} occurred` : ''

            return {
                diagnostics: diagnosticsAndActions.map(({ diagnostic, action, errors }) =>
                    errors.length === 0
                        ? action !== undefined && (action === LOADING || action.result === LOADING)
                            ? withExtraDetail(diagnostic, 'Loading fix...')
                            : diagnostic
                        : withExtraDetail(diagnostic, errors.map(e => e.message).join(', '))
                ),
                edits: diagnosticsAndActions
                    .map(({ action }) =>
                        action !== undefined &&
                        action !== LOADING &&
                        action.result !== LOADING &&
                        !isErrorLike(action.result)
                            ? action.result
                            : undefined
                    )
                    .filter(isDefined)
                    .map(edit => WorkspaceEdit.fromJSON(edit)),
                status: {
                    isLoading: loadingCount > 0,
                    progress: [totalCount - loadingCount, totalCount] as const,
                    messages: [
                        loadingCount > 0
                            ? `Generating fixes ${totalCount - loadingCount}/${totalCount}... ${
                                  errorsFoundStr ? `(${errorsFoundStr})` : ''
                              }`
                            : errorsFoundStr,
                    ],
                    errors: allErrors.map(e => e.message),
                },
            }
        })
    )
}

type EditsBehaviorResult = Pick<
    GQL.IExpCreateCampaignInput['extensionData'],
    'rawChangesets' | 'rawSideEffects' | 'rawLogMessages'
>

const executeEditsBehavior = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    edits: WorkspaceEdit[],
    workflow: Workflow
): Observable<EditsBehaviorResult> => {
    const command: Command =
        workflow && workflow.behaviors && workflow.behaviors.edits
            ? workflow.behaviors.edits
            : { command: 'changesets.byRepositoryAndBaseBranch' }
    const combinedEdit = combineWorkspaceEdits(edits)
    const changesets: Observable<Changeset[]> = from(
        extensionsController.services.commands.executeCommand({
            command: command.command,
            arguments: [(combinedEdit as any).toJSON(), workflow.variables, ...(command.arguments || [])],
        })
    )
    return changesets.pipe(
        switchMap(changesets =>
            changesets.length > 0
                ? combineLatest(
                      changesets.map(changeset =>
                          from(
                              computeDiffFromEdits(extensionsController, [
                                  WorkspaceEdit.fromJSON(changeset.edit as any),
                              ])
                          ).pipe(
                              map(diff => {
                                  const input: GQL.IChangesetInput = {
                                      ...changeset,
                                      patch: diff.map(d => d.patchWithFullURIs).join('\n'),
                                  }
                                  delete (input as any).edit
                                  return {
                                      changesets: [input],
                                      sideEffects: changeset.sideEffects || [],
                                      logMessages: [],
                                  }
                              })
                          )
                      )
                  ).pipe(
                      map(results =>
                          results.reduce<EditsBehaviorResult>(
                              (all, r) => ({
                                  ...all,
                                  rawChangesets: [...all.rawChangesets, ...r.changesets],
                                  rawSideEffects: [...all.rawSideEffects, ...r.sideEffects],
                                  rawLogMessages: [...all.rawLogMessages, ...r.logMessages],
                              }),
                              { rawChangesets: [], rawSideEffects: [], rawLogMessages: [] }
                          )
                      )
                  )
                : of<EditsBehaviorResult>({ rawChangesets: [], rawSideEffects: [], rawLogMessages: [] })
        )
    )
}

const sum = (nums: number[]): number => nums.reduce((sum, n) => sum + n, 0)

export const getCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    workflow: Workflow,
    campaign: Pick<GQL.IExpCreateCampaignInput, 'name' | 'body'>
): Observable<readonly [GQL.IExpCreateCampaignInput['extensionData'], ExtensionDataStatus]> => {
    workflow = {
        ...workflow,
        variables: { ...workflow.variables, title: campaign.name, body: campaign.body },
    }
    return workflow.run && workflow.run.length > 0
        ? combineLatest(
              workflow.run.map(run =>
                  executeRun(extensionsController, { ...run, variables: { ...workflow.variables, ...run.variables } })
              )
          ).pipe(
              map(results => {
                  const editDiagnostics = results.flatMap(r => r.edits.flatMap(e => e.diagnostics)).filter(isDefined)
                  const combined: DiagnosticsAndFileDiffs = {
                      diagnostics: [...results.flatMap(r => r.diagnostics), ...editDiagnostics],
                      edits: results.flatMap(r => r.edits),
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
              switchMap(({ diagnostics, edits, status }) =>
                  executeEditsBehavior(extensionsController, edits, workflow).pipe(
                      map(
                          ({ rawChangesets, rawSideEffects, rawLogMessages }) =>
                              [
                                  {
                                      rawDiagnostics: diagnostics.map(d =>
                                          // tslint:disable-next-line: no-object-literal-type-assertion
                                          JSON.stringify({
                                              __typename: 'Diagnostic',
                                              type: isDiagnosticWithType(d) ? d.type : null,
                                              data: d,
                                          } as GQL.IDiagnostic)
                                      ),
                                      rawChangesets,
                                      rawSideEffects,
                                      rawLogMessages,
                                  },
                                  status,
                              ] as readonly [GQL.IExpCreateCampaignInput['extensionData'], ExtensionDataStatus]
                      )
                  )
              )
          )
        : of([{ rawDiagnostics: [], rawChangesets: [], rawSideEffects: [], rawLogMessages: [] }, { isLoading: false }])
}

function isDiagnosticWithType(d: Diagnostic): d is DiagnosticWithType {
    return typeof (d as any).type === 'string'
}

/** Waits for `getCampaignExtensionData` to finish loading and returns the result. */
export const getCompleteCampaignExtensionData = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    workflow: Workflow,
    campaign: Pick<GQL.IExpCreateCampaignInput, 'name' | 'body'>
): Promise<GQL.IExpCreateCampaignInput['extensionData']> =>
    getCampaignExtensionData(extensionsController, workflow, campaign)
        .pipe(
            first(([, status]) => !status.isLoading),
            map(([extensionData]) => extensionData)
        )
        .toPromise()
