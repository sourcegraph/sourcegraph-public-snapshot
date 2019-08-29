import * as sourcegraph from 'sourcegraph'
import _eslintConfigStandard from 'eslint-config-standard'
import { TSESLint } from '@typescript-eslint/experimental-utils'
import * as tseslintParser from '@typescript-eslint/parser'
import { Linter, Rule, CLIEngine } from 'eslint'
import { isDefined } from '../../../../shared/src/util/types'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { flatten, sortedUniq, sortBy } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from, combineLatest } from 'rxjs'
import { map, switchMap, startWith, first, toArray, filter } from 'rxjs/operators'
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

const FIX_EDIT_COMMAND = 'eslint.fix'
const DISABLE_RULE_ON_LINE_COMMAND = 'eslint.disableRuleOnLine'

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(sourcegraph.commands.registerActionEditCommand(FIX_EDIT_COMMAND, fixEditCommandCallback))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(DISABLE_RULE_ON_LINE_COMMAND, disableRuleOnLineCommandCallback)
    )
    return subscriptions
}

interface Settings {
    ['eslint.rules']?: { [ruleId: string]: RulePolicy }
}

enum RulePolicy {
    Ignore = 'ignore',
    Default = 'default',
}

const TAG = 'eslint'

const LOADING: 'loading' = 'loading'

const diagnostics: Observable<sourcegraph.Diagnostic[] | typeof LOADING> = from(sourcegraph.workspace.rootChanges).pipe(
    startWith(void 0),
    map(() => sourcegraph.workspace.roots),
    switchMap(async roots => {
        if (roots.length > 0) {
            return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
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
                            includes: ['\\.[jt]sx?$'], // TODO!(sqs): typescript only
                            type: 'regexp',
                        },
                        maxResults: 19, //MAX_RESULTS,
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

        const stdRules = _eslintConfigStandard.rules
        for (const ruleId of Object.keys(stdRules)) {
            if (!linter.getRules().has(ruleId)) {
                delete stdRules[ruleId]
            }
        }
        // delete stdRules.indent
        // delete stdRules['space-before-function-paren']
        // delete stdRules['no-undef']
        // delete stdRules['comma-dangle']
        // delete stdRules['no-unused-vars']

        const config: Linter.Config = {
            parser: '@typescript-eslint/parser',
            parserOptions: tseslintConfig,
            rules: {
                ...stdRules,
                // 'no-useless-constructor': 0,
                // 'spaced-comment': 0,
            },
            settings: {
                react: { version: '16.3' },
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
                flatten(
                    docs.map(doc => {
                        const lintMessages = linter.verify(doc.text, config, {
                            filename: new URL(doc.uri).pathname.slice(1),
                        })
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
                                    resource: new URL(doc.uri),
                                    range: rangeForLintMessage(doc, r),
                                    message: `${r.message} (${r.ruleId})`,
                                    source: r.source,
                                    severity: linterSeverityToDiagnosticSeverity(r.severity),
                                    data: JSON.stringify(r),
                                    tags: [r.ruleId, TAG],
                                } as sourcegraph.Diagnostic
                            })
                            .filter(isDefined)
                        return diagnostics
                    })
                )
            )
        )
    }),
    switchMap(results => results),
    startWith(LOADING)
)

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('eslint', {
            provideDiagnostics: _scope =>
                diagnostics.pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    return subscriptions
}

const diagnosticsRuleIds = diagnostics.pipe(
    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING),
    map(diagnostics => {
        const ruleIdsSet = new Set<string>()
        for (const diag of diagnostics) {
            const rule = getLintMessageFromDiagnosticData(diag)
            ruleIdsSet.add(rule.ruleId)
        }

        const ruleIds = Array.from(ruleIdsSet.values()).sort()
        return ruleIds
    })
)

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
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
                                      computeEdit: { title: 'Fix', command: FIX_EDIT_COMMAND },
                                      diagnostics: [diag],
                                  },
                              ]
                            : []),
                        {
                            title: `Disable rule '${lintMessage.ruleId}'`,
                            edit: createWorkspaceEditForDisablingRule(doc, lintMessage),
                            computeEdit: { title: 'Disable rule on line', command: DISABLE_RULE_ON_LINE_COMMAND },
                            diagnostics: [diag],
                        },
                        // {
                        //     title: `Ignore all (${lintMessage.ruleId})`,
                        //     command: updateRulePoliciesCommand(RulePolicy.Ignore, lintMessage.ruleId),
                        //     diagnostics: [diag],
                        // },
                        // {
                        //     title: `Documentation`,
                        //     command: {
                        //         title: '',
                        //         command: 'open',
                        //         arguments: [`https://eslint.org/docs/rules/${lintMessage.ruleId}`],
                        //     },
                        // },
                    ].filter(isDefined)
                })
            )
        },
    }
}

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
    return diag.tags && diag.tags.includes(TAG)
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

function createWorkspaceEditForDisablingRule(
    doc: sourcegraph.TextDocument,
    lintMessage: Linter.LintMessage,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    const range = rangeForLintMessage(doc, lintMessage)
    // TODO!(sqs): get indent of previous line - in vscode this is inserted on the client automatically, check out how they do it because that seems neat (https://sourcegraph.com/github.com/microsoft/vscode-tslint@30d1a7ae25b0331466f1a54b4f7d23d60fa2da30/-/blob/tslint-server/src/tslintServer.ts#L618)
    const indent = doc.text.slice(doc.offsetAt(range.start.with(undefined, 0))).match(/[ \t]*/)
    edit.insert(
        new URL(doc.uri),
        range.start.with(undefined, 0),
        `${indent ? indent[0] : ''}// eslint-disable-next-line ${lintMessage.ruleId}\n`
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

async function fixEditCommandCallback(diagnostic: sourcegraph.Diagnostic): Promise<sourcegraph.WorkspaceEdit> {
    const r: Linter.LintMessage = JSON.parse(diagnostic.data)
    const doc = await sourcegraph.workspace.openTextDocument(diagnostic.resource)
    // TODO!(sqs): when DiagnosticQuery supports >1 tag, then remove this `new WorkspaceEdit()` branch and just ensure fixEditCommandCallback only gets called for diagnostics with a new tab `auto-fixable`
    return r.fix ? createWorkspaceEditFromESLintFix(doc, r.fix) : new sourcegraph.WorkspaceEdit()
}

async function disableRuleOnLineCommandCallback(
    diagnostic: sourcegraph.Diagnostic
): Promise<sourcegraph.WorkspaceEdit> {
    const r: Linter.LintMessage = JSON.parse(diagnostic.data)
    const doc = await sourcegraph.workspace.openTextDocument(diagnostic.resource)
    return createWorkspaceEditForDisablingRule(doc, r)
}
