import { isEqual } from 'lodash'
import { combineLatest, Observable, of, EMPTY, from } from 'rxjs'
import { distinctUntilChanged, first, map, startWith, switchMap, throttleTime, defaultIfEmpty } from 'rxjs/operators'
import { CodeActionError, isCodeActionError } from '../../../../shared/src/api/client/services/codeActions'
import { DiagnosticWithType } from '../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../shared/src/api/types/action'
import { fromDiagnostic } from '../../../../shared/src/api/types/diagnostic'
import { WorkspaceEdit, combineWorkspaceEdits } from '../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../shared/src/util/strings'
import { propertyIsDefined } from '../../../../shared/src/util/types'
import { getCodeActions } from '../diagnostics/backend'
import { WorkflowRun, Workflow, Command } from '../../schema/workflow.schema'
import { Changeset } from '../../../../extensions/enterprise/sandbox/src/workflow/behaviors/edits'
import { computeDiffFromEdits } from './backend/computeDiff'

interface DiagnosticsAndFileDiffs {
    diagnostics: DiagnosticWithType[]
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

interface DiagnosticWithAction {
    diagnostic: DiagnosticWithType
    errors: CodeActionError[]
    action: Action | typeof LOADING | undefined
}

const executeRun = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    { variables: runVariables, diagnostics: runDiagnostics, codeActions: runCodeActions }: WorkflowRun
): Observable<DiagnosticsAndFileDiffs> => {
    const diagnostics = runDiagnostics
        ? extensionsController.services.diagnostics.observeDiagnostics({}, runVariables || {}, runDiagnostics)
        : EMPTY
    const codeActions = runCodeActions
        ? diagnostics.pipe(
              defaultIfEmpty<DiagnosticWithType[]>([]),
              switchMap(diagnostics =>
                  diagnostics.length > 0
                      ? combineLatest(
                            diagnostics.map(d =>
                                getCodeActions({ diagnostic: d, extensionsController }).pipe(
                                    map(actions => ({
                                        diagnostic: d,
                                        errors: actions.filter(isCodeActionError),
                                        action: actions
                                            .filter((a): a is Action => !isCodeActionError(a))
                                            .filter(propertyIsDefined('computeEdit'))
                                            .find(({ computeEdit }) =>
                                                runCodeActions.some(a => a.command === computeEdit.command)
                                            ),
                                    })),
                                    startWith<DiagnosticWithAction>({
                                        diagnostic: d,
                                        errors: [],
                                        action: LOADING,
                                    })
                                )
                            )
                        )
                      : EMPTY
              )
          )
        : EMPTY
    return codeActions.pipe(
        throttleTime(2500, undefined, { leading: true, trailing: true }),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        defaultIfEmpty<DiagnosticWithAction[]>([]),
        switchMap(async diagnosticsAndActions => {
            const actionInvocations = diagnosticsAndActions
                .filter(propertyIsDefined('action'))
                .map(d => ({
                    actionEditCommand: d.action !== LOADING ? d.action.computeEdit : undefined,
                    diagnostic: fromDiagnostic(d.diagnostic),
                }))
                .filter(propertyIsDefined('actionEditCommand'))

            const withExtraDetail = (diagnostic: DiagnosticWithType, detail: string): DiagnosticWithType => ({
                ...diagnostic,
                detail: `${diagnostic.detail ? `${diagnostic.detail} - ` : ''}${detail}`,
            })

            const loadingCount = diagnosticsAndActions.filter(d => d.action === LOADING).length
            const totalCount = diagnosticsAndActions.filter(d => d.action !== undefined).length
            const allErrors = diagnosticsAndActions.flatMap(da => da.errors)
            const errorsFoundStr =
                allErrors.length > 0 ? `${allErrors.length} ${pluralize('error', allErrors.length)} occurred` : ''

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
                edits: await Promise.all(
                    actionInvocations.map(async ({ actionEditCommand, diagnostic }) => {
                        const edit = await extensionsController.services.commands.executeActionEditCommand(
                            diagnostic,
                            actionEditCommand
                        )
                        return edit && WorkspaceEdit.fromJSON(edit)
                    })
                ),
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

type EditsBehaviorResult = Pick<GQL.IExpCreateCampaignInput['extensionData'], 'rawChangesets' | 'rawSideEffects'>

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
                                  return { changesets: [input], sideEffects: changeset.sideEffects || [] }
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
                              }),
                              { rawChangesets: [], rawSideEffects: [] }
                          )
                      )
                  )
                : of<EditsBehaviorResult>({ rawChangesets: [], rawSideEffects: [] })
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
                  const combined: DiagnosticsAndFileDiffs = {
                      diagnostics: results.flatMap(r => r.diagnostics),
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
                          ({ rawChangesets, rawSideEffects }) =>
                              [
                                  {
                                      rawDiagnostics: diagnostics.map(d =>
                                          // tslint:disable-next-line: no-object-literal-type-assertion
                                          JSON.stringify({
                                              __typename: 'Diagnostic',
                                              type: d.type,
                                              data: d,
                                          } as GQL.IDiagnostic)
                                      ),
                                      rawChangesets,
                                      rawSideEffects,
                                  },
                                  status,
                              ] as readonly [GQL.IExpCreateCampaignInput['extensionData'], ExtensionDataStatus]
                      )
                  )
              )
          )
        : of([{ rawDiagnostics: [], rawChangesets: [], rawSideEffects: [] }, { isLoading: false }])
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
