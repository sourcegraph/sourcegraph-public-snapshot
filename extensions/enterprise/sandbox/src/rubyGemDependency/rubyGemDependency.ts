import * as sourcegraph from 'sourcegraph'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { flatten } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, toArray, filter } from 'rxjs/operators'
import { settingsObservable, memoizedFindTextInFiles } from '../util'
import { rubyGemfileDependencies, RubyGemfileDependency } from './rubyGemfile'
import { bundlerRemove } from './rubyBundler'

const REMOVE_COMMAND = 'rubyGemDependency.remove'

interface Settings {}

export interface RubyGemDependencyCampaignContext {
    gemName: string
    createChangesets: boolean
    filters?: string
    campaignName?: string
}

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('rubyGemDependency', {
            provideDiagnostics: (_scope, context) =>
                provideDiagnostics((context as any) as RubyGemDependencyCampaignContext).pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(REMOVE_COMMAND, async diagnostic =>
            diagnostic
                ? computeRemoveDependencyEdit(
                      diagnostic,
                      await sourcegraph.workspace.openTextDocument(diagnostic.resource)
                  )
                : new sourcegraph.WorkspaceEdit()
        )
    )
    return subscriptions
}

const DEPENDENCY_TAG = 'type:rubyGemDependency'

const LOADING = 'loading' as const

const provideDiagnostics = (
    context: RubyGemDependencyCampaignContext
): Observable<sourcegraph.Diagnostic[] | typeof LOADING> =>
    context.gemName
        ? from(sourcegraph.workspace.rootChanges).pipe(
              startWith(void 0),
              map(() => sourcegraph.workspace.roots),
              switchMap(async roots => {
                  if (roots.length > 0) {
                      return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
                  }

                  const results = flatten(
                      await from(
                          memoizedFindTextInFiles(
                              {
                                  pattern: '',
                                  type: 'regexp',
                              },
                              {
                                  repositories: {
                                      // includes: ['acts-as-taggable'],
                                      includes: ['openproject|activeadmin'],
                                      excludes: ['sd9'],
                                      type: 'regexp',
                                  },
                                  files: {
                                      // TODO!(sqs): also support non-top-level Gemfiles
                                      // includes: ['(^|/)Gemfile$'],
                                      includes: ['^Gemfile$'],
                                      type: 'regexp',
                                  },
                                  maxResults: 20, // TODO!(sqs): increase
                              }
                          )
                      )
                          .pipe(toArray())
                          .toPromise()
                  )

                  await Promise.all(
                      results.map(async ({ uri }) => sourcegraph.workspace.openTextDocument(new URL(uri)))
                  )

                  return from(settingsObservable<Settings>()).pipe(
                      switchMap(async () =>
                          flatten(
                              await Promise.all(
                                  results.map(async ({ uri: gemfileUriStr }) => {
                                      const p = parseRepoURI(gemfileUriStr)

                                      const gemfile = await sourcegraph.workspace.openTextDocument(
                                          new URL(gemfileUriStr)
                                      )
                                      let gemfileLock: sourcegraph.TextDocument | undefined
                                      try {
                                          gemfileLock = await sourcegraph.workspace.openTextDocument(
                                              new URL(gemfileUriStr + '.lock')
                                          )
                                      } catch (err) {
                                          // TODO!(sqs): check error is not-exists
                                      }

                                      try {
                                          const allDeps = await rubyGemfileDependencies(
                                              {
                                                  Gemfile: gemfile.text!,
                                                  'Gemfile.lock': gemfileLock && gemfileLock.text!,
                                              },
                                              {
                                                  repository: p.repoName,
                                                  commit: p.commitID!,
                                              }
                                          )
                                          const matchingDeps = allDeps.filter(({ name }) => name === context.gemName)
                                          return flatten(
                                              matchingDeps.map<sourcegraph.Diagnostic[]>(dep => {
                                                  const partial: Pick<
                                                      sourcegraph.Diagnostic,
                                                      'resource' | 'message' | 'range'
                                                  >[] =
                                                      'range' in dep
                                                          ? [
                                                                {
                                                                    resource: new URL(gemfileUriStr),
                                                                    message: `Ruby gem '${dep.name}' is banned`,
                                                                    range: dep.range,
                                                                },
                                                            ]
                                                          : dep.directAncestors.map(directAncestor => {
                                                                const directAncestorDep = allDeps.find(
                                                                    d => d.name === directAncestor
                                                                )!
                                                                return {
                                                                    resource: new URL(gemfileUriStr + '.lock'), // TODO!(sqs): might not be in lockfile
                                                                    message: `Ruby gem '${directAncestorDep.name}' transitively depends on banned Ruby gem '${dep.name}'`,
                                                                    range: directAncestorDep.range,
                                                                }
                                                            })
                                                  return partial.map(partial => ({
                                                      ...partial,
                                                      detail: `see campaign [${context.campaignName}](#)`,
                                                      severity: sourcegraph.DiagnosticSeverity.Warning,
                                                      data: JSON.stringify(dep),
                                                      tags: [DEPENDENCY_TAG, dep.name],
                                                  }))
                                              })
                                          )
                                      } catch (err) {
                                          console.error(err)
                                          // if (sourcegraph.app.activeWindow) {
                                          // sourcegraph.app.activeWindow.showNotification(
                                          //     `Error: ${err.message}`,
                                          //         sourcegraph.NotificationType.Error
                                          //     )
                                          // }
                                          return []
                                      }
                                  })
                              )
                          )
                      )
                  )
              }),
              switchMap(results => results),
              startWith(LOADING)
          )
        : of<sourcegraph.Diagnostic[]>([])

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: async (doc, _rangeOrSelection, context): Promise<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isProviderDiagnostic)
            if (!diag) {
                return []
            }
            return [
                {
                    title: `Remove Ruby gem dependency from Gemfile`,
                    edit: await computeRemoveDependencyEdit(diag, doc),
                    computeEdit: { title: 'Remove', command: REMOVE_COMMAND },
                    diagnostics: [diag],
                },
            ]
        },
    }
}

function isProviderDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return !!diag.tags && diag.tags.includes(DEPENDENCY_TAG)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): RubyGemfileDependency {
    return JSON.parse(diag.data!)
}

async function computeRemoveDependencyEdit(
    diag: sourcegraph.Diagnostic,
    gemfile: sourcegraph.TextDocument
): Promise<sourcegraph.WorkspaceEdit> {
    const edit = new sourcegraph.WorkspaceEdit()
    const dep = getDiagnosticData(diag)

    const addBundlerRemoveEdits = async (
        gemfile: sourcegraph.TextDocument,
        edit: sourcegraph.WorkspaceEdit
    ): Promise<void> => {
        const p = parseRepoURI(gemfile.uri)
        const result = await bundlerRemove(dep.name, { repository: p.repoName, commit: p.commitID! })
        for (const path of Object.keys(result.files)) {
            const uri = new URL(gemfile.uri.replace('Gemfile', '') + path)
            const doc = await sourcegraph.workspace.openTextDocument(uri)
            if (doc.text !== result.files[path]) {
                edit.replace(
                    uri,
                    new sourcegraph.Range(new sourcegraph.Position(0, 0), doc.positionAt(doc.text!.length)),
                    result.files[path]
                )
            }
        }
    }
    const addIndirectDependencyEdits = (gemfile: sourcegraph.TextDocument, edit: sourcegraph.WorkspaceEdit): void => {
        // Handle when the dep to remove is an indirect dep.
        edit.insert(
            new URL(gemfile.uri),
            gemfile.positionAt(gemfile.text!.length - 1),
            `\n\nraise ${JSON.stringify(
                `Gem ${dep.name} is banned${
                    dep.directAncestors.length > 0
                        ? `, but the following gems use it: ${dep.directAncestors.join(', ')}`
                        : ''
                }. Manual fixup needed.`
            )}`
        )
    }

    const isLikelyDirectDependency = gemfile.text!.includes(dep.name) // HACK!(sqs) TODO!(sqs)
    if (isLikelyDirectDependency) {
        await addBundlerRemoveEdits(gemfile, edit)
    }
    const countEdits = () => Array.from(edit.textEdits()).length
    if (!isLikelyDirectDependency && countEdits() === 0) {
        await addBundlerRemoveEdits(gemfile, edit)
    }
    if (countEdits() === 0) {
        addIndirectDependencyEdits(gemfile, edit)
    }
    return edit
}
