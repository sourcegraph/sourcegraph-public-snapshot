import * as sourcegraph from 'sourcegraph'
import _eslintConfigStandard from 'eslint-config-standard'
import { TSESLint } from '@typescript-eslint/experimental-utils'
import * as tseslintParser from '@typescript-eslint/parser'
import { Linter, Rule, CLIEngine } from 'eslint'
import { isDefined } from '../../../../shared/src/util/types'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { flatten, sortedUniq, sortBy } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from, combineLatest } from 'rxjs'
import { map, switchMap, startWith, first, toArray } from 'rxjs/operators'
import { queryGraphQL, settingsObservable, memoizedFindTextInFiles } from './util'
import * as GQL from '../../../../shared/src/graphql/schema'
import { OTHER_CODE_ACTIONS, MAX_RESULTS, REPO_INCLUDE } from './misc'

/**
 * From https://sourcegraph.com/github.com/eslint/eslint@d8f26886f19a17f2e1cdcb91e2db84fc7ba3fdfb/-/blob/lib/shared/types.js#L125-129.
 */
type GlobalConf = boolean | 'off' | 'readable' | 'readonly' | 'writable' | 'writeable'
interface ConfigData {
    env: Record<string, boolean>
    extends: string | string[]
    globals: Record<string, GlobalConf>
    parser: string
    parserOptions: Linter.ParserOptions
    plugins: string[]
    processor: string
    root: boolean
    rules: Record<string, Linter.RuleLevel | Linter.RuleLevelAndOptions>
    settings: Object
}
interface Environment {
    globals: Record<string, GlobalConf>
    parserOptions: Linter.ParserOptions
}
interface Processor {
    preprocess?: (text: string, filename: string) => Array<string | { text: string; filename: string }>
    postprocess?: (messagesList: Linter.LintMessage[][], filename: string) => Linter.LintMessage[]
    supportsAutofix?: boolean
}
interface Plugin {
    configs: Record<string, ConfigData>
    environments: Record<string, Environment>
    processors: Record<string, Processor>
    rules: Record<string, Rule.RuleModule>
}

export function registerESLintRules(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.status.registerStatusProvider('eslint', createStatusProvider(diagnostics)))
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    return subscriptions
}

interface Settings {
    ['eslint.rules']?: { [ruleId: string]: RulePolicy }
}

enum RulePolicy {
    Ignore = 'ignore',
    Default = 'default',
}

const TAG_ESLINT = 'eslint'

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()

    const diagnosticsCollection = sourcegraph.languages.createDiagnosticCollection('eslint')
    subscriptions.add(diagnosticsCollection)
    subscriptions.add(
        diagnostics.subscribe(entries => {
            diagnosticsCollection.set(entries)
        })
    )

    return diagnosticsCollection
}

function createStatusProvider(diagnostics: Observable<[URL, sourcegraph.Diagnostic[]][]>): sourcegraph.StatusProvider {
    const LOADING: 'loading' = 'loading'
    return {
        provideStatus: (scope): sourcegraph.Subscribable<sourcegraph.Status | null> => {
            // TODO!(sqs): dont ignore scope
            return combineLatest([diagnostics.pipe(startWith(LOADING)), settingsObservable<Settings>()]).pipe(
                map(([diagnostics, settings]) => {
                    const deps = new Map<string, RulePolicy>()
                    if (diagnostics !== LOADING) {
                        for (const [, diags] of diagnostics) {
                            for (const diag of diags) {
                                const rule = getLintMessageFromDiagnosticData(diag)
                                const rulePolicy = getRulePolicyFromSettings(settings, rule.ruleId)
                                deps.set(rule.ruleId, rulePolicy)
                            }
                        }
                    }

                    const rulePolicy: sourcegraph.Status = {
                        title: 'ESLint',
                        description: {
                            kind: sourcegraph.MarkupKind.Markdown,
                            value: 'Checks code using ESLint, an open-source JavaScript linting utility.',
                        },
                        state:
                            diagnostics === LOADING
                                ? {
                                      completion: sourcegraph.StatusCompletion.InProgress,
                                      message: 'Running ESLint...',
                                  }
                                : {
                                      completion: sourcegraph.StatusCompletion.Completed,
                                      result:
                                          diagnostics.length > 0
                                              ? sourcegraph.StatusResult.Failure
                                              : sourcegraph.StatusResult.Success,
                                      message:
                                          diagnostics.length > 0
                                              ? 'ESLint problems found'
                                              : 'Code is compliant with ESLint',
                                  },
                        sections: {
                            settings: {
                                kind: sourcegraph.MarkupKind.Markdown,
                                value: `
- Use \`eslint@6.0.1\`
- Check for new, recommended ESLint rules
- Ignore projects with only JavaScript files`,
                            },
                            notifications: {
                                kind: sourcegraph.MarkupKind.Markdown,
                                value: `
- Fail changesets that add code not checked by ESLint
- Notify **@felixfbecker** of new ESLint rules`,
                            },
                        },
                        notifications: sortBy(Array.from(deps.entries()), 0)
                            .filter(([, rulePolicy]) => rulePolicy !== RulePolicy.Ignore)
                            .map<sourcegraph.Notification>(([ruleId, rulePolicy]) => ({
                                title: ruleId,
                                type:
                                    rulePolicy === RulePolicy.Default
                                        ? sourcegraph.NotificationType.Warning
                                        : sourcegraph.NotificationType.Error,
                            })),
                    }
                    return rulePolicy
                })
            )
        },
    }
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.CodeAction[]> => {
            const diag = context.diagnostics.find(isESLintDiagnostic)
            if (!diag) {
                return of([])
            }
            const lintMessage = getLintMessageFromDiagnosticData(diag)
            return from(settingsObservable<Settings>()).pipe(
                map(settings => {
                    const rulePolicy = getRulePolicyFromSettings(settings, lintMessage.ruleId)
                    if (rulePolicy === RulePolicy.Ignore) {
                        return []
                    }
                    return [
                        ...(lintMessage.fix
                            ? [
                                  {
                                      title: `Fix`,
                                      edit: createWorkspaceEditFromESLintFix(doc, lintMessage.fix),
                                      diagnostics: [diag],
                                  },
                              ]
                            : []),
                        {
                            title: `Ignore all (${lintMessage.ruleId})`,
                            command: updateRulePoliciesCommand(RulePolicy.Ignore, lintMessage.ruleId),
                            diagnostics: [diag],
                        },
                        {
                            title: `Documentation`,
                            command: {
                                title: '',
                                command: 'open',
                                arguments: [`https://eslint.org/docs/rules/${lintMessage.ruleId}`],
                            },
                        },
                    ].filter(isDefined)
                })
            )
        },
    }
}

const diagnostics: Observable<[URL, sourcegraph.Diagnostic[]][]> = from(sourcegraph.workspace.rootChanges).pipe(
    startWith(void 0),
    map(() => sourcegraph.workspace.roots),
    switchMap(async roots => {
        if (roots.length > 0) {
            return of([]) // TODO!(sqs): dont run in comparison mode
        }

        const results = flatten(
            await from(
                memoizedFindTextInFiles(
                    { pattern: '', type: 'regexp' },
                    {
                        repositories: {
                            includes: [REPO_INCLUDE],
                            type: 'regexp',
                        },
                        files: {
                            includes: ['\\.tsx?$'], // TODO!(sqs): typescript only
                            type: 'regexp',
                        },
                        maxResults: MAX_RESULTS * 5,
                    }
                )
            )
                .pipe(toArray())
                .toPromise()
        )
        const docs = await Promise.all(
            results.map(async ({ uri }) => sourcegraph.workspace.openTextDocument(new URL(uri)))
        )

        const tseslintConfig: TSESLint.ParserOptions = {
            ecmaVersion: 2018,
            range: true,
            sourceType: 'module',
            filePath: 'foo.tsx',
            useJSXTextNode: true,
            ecmaFeatures: { jsx: true },
        }
        const linter = new Linter()
        linter.defineParser('@typescript-eslint/parser', tseslintParser)

        // const rules = eslintConfigStandard
        // for (const ruleId of Object.keys(rules)) {
        //     if (!linter.getRules().has(ruleId)) {
        //         delete rules[ruleId]
        //     }
        // }
        // delete rules.indent
        // delete rules['space-before-function-paren']
        // delete rules['no-undef']
        // delete rules['comma-dangle']
        // delete rules['no-unused-vars']
        const config: Linter.Config = {
            parser: '@typescript-eslint/parser',
            parserOptions: tseslintConfig,
            rules: {
                'no-useless-constructor': 0,
                'spaced-comment': 0,
            },
        }
        const plugins: Record<string, Plugin> = {
            react: require('eslint-plugin-react'),
            'react-hooks': require('eslint-plugin-react-hooks'),
        }
        for (const pluginName of Object.keys(plugins)) {
            const plugin = plugins[pluginName]
            for (const ruleName of Object.keys(plugin.rules)) {
                const rule = plugin.rules[ruleName]
                linter.defineRule(`${pluginName}/${ruleName}`, rule)
            }
            config.rules = {
                ...config.rules,
                ...(plugin.configs && plugin.configs.recommended ? plugin.configs.recommended.rules : {}),
            }
        }
        delete config.rules['react/prop-types']
        config.rules = {
            ...config.rules,
            'react-hooks/rules-of-hooks': 'error',
            'react-hooks/exhaustive-deps': 'error',
        }

        return from(settingsObservable<Settings>()).pipe(
            map(settings =>
                docs
                    .map(doc => {
                        const lintMessages = linter
                            .verify(doc.text, config, {
                                filename: new URL(doc.uri).pathname.slice(1),
                            })
                            .slice(0, 2)
                        const diagnostics: sourcegraph.Diagnostic[] = lintMessages
                            .map(r => {
                                if (r.fatal) {
                                    return null // TODO!(sqs): dont suppress
                                }
                                const rulePolicy = getRulePolicyFromSettings(settings, r.ruleId)
                                if (rulePolicy === RulePolicy.Ignore) {
                                    return null
                                }
                                return {
                                    message: r.message,
                                    source: r.source,
                                    range: rangeForLintMessage(doc, r),
                                    severity: linterSeverityToDiagnosticSeverity(r.severity),
                                    data: JSON.stringify(r),
                                    tags: [TAG_ESLINT],
                                } as sourcegraph.Diagnostic
                            })
                            .filter(isDefined)
                        return diagnostics.length > 0
                            ? ([new URL(doc.uri), diagnostics] as [URL, sourcegraph.Diagnostic[]])
                            : null
                    })
                    .filter(isDefined)
            )
        )
    }),
    switchMap(results => results)
)

function rangeForLintMessage(doc: sourcegraph.TextDocument, m: Linter.LintMessage): sourcegraph.Range {
    if (m.line === undefined && m.column === undefined) {
        return new sourcegraph.Range(0, 0, 1, 0)
    }
    const start = new sourcegraph.Position(m.line - 1, m.column - 1)
    let end: sourcegraph.Position
    if (m.endLine === undefined && m.endColumn === undefined) {
        const wordRange = doc.getWordRangeAtPosition(start)
        end = wordRange ? wordRange.end : start
    } else {
        end = new sourcegraph.Position(m.endLine - 1, m.endColumn - 1)
    }
    return new sourcegraph.Range(start, end)
}

function linterSeverityToDiagnosticSeverity(ruleSeverity: Linter.Severity): sourcegraph.DiagnosticSeverity {
    switch (ruleSeverity) {
        case 0:
            return sourcegraph.DiagnosticSeverity.Information
        case 1:
            return sourcegraph.DiagnosticSeverity.Warning
        case 2:
            return sourcegraph.DiagnosticSeverity.Error
        default:
            return sourcegraph.DiagnosticSeverity.Error
    }
}

function getRulePolicyFromSettings(settings: Settings, ruleId: string): RulePolicy {
    return (settings['eslint.rules'] && settings['eslint.rules'][ruleId]) || RulePolicy.Default
}

function isESLintDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return diag.tags && diag.tags.includes(TAG_ESLINT)
}

function getLintMessageFromDiagnosticData(diag: sourcegraph.Diagnostic): Linter.LintMessage {
    return JSON.parse(diag.data!)
}

function createWorkspaceEditFromESLintFix(
    doc: sourcegraph.TextDocument,
    fix: Rule.Fix,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    edit.replace(
        new URL(doc.uri),
        new sourcegraph.Range(doc.positionAt(fix.range[0]), doc.positionAt(fix.range[1])),
        fix.text
    )
    return edit
}

/**
 * Returns the object describing how to invoke the command to update the rule policy.
 */
function updateRulePoliciesCommand(
    rulePolicy: RulePolicy.Ignore | RulePolicy.Default,
    ruleId: string
): sourcegraph.Command {
    return { title: '', command: 'updateConfiguration', arguments: [['eslint.rules', ruleId], rulePolicy] }
}
