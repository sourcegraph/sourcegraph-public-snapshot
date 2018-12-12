/**
 * @license
 * Copyright 2018 Palantir Technologies, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
import * as Lint from 'tslint'
import { findImports, ImportKind } from 'tsutils'
import * as ts from 'typescript'

interface IOption {
    pattern: RegExp
    message?: string
}

export class Rule extends Lint.Rules.AbstractRule {
    public static metadata: Lint.IRuleMetadata = {
        ruleName: 'ban-imports',
        description: Lint.Utils.dedent`
            Bans specific modules from being imported.`,
        options: {
            type: 'list',
            listType: {
                type: 'array',
                items: { type: 'string' },
                minLength: 1,
                maxLength: 2,
            },
        },
        optionsDescription: Lint.Utils.dedent`
            A list of \`["regex", "optional explanation here"]\`, which bans
            imports that match \`regex\``,
        optionExamples: [[true, ['react-router-dom', 'Use {} instead.'], ['String']]],
        type: 'typescript',
        typescriptOnly: false,
    }
    /* tslint:enable:object-literal-sort-keys */

    public static FAILURE_STRING_FACTORY(pattern: string, messageAddition?: string): string {
        return `Import of module matching pattern '${pattern}' is banned.${
            messageAddition !== undefined ? ` ${messageAddition}` : ''
        }`
    }

    public apply(sourceFile: ts.SourceFile): Lint.RuleFailure[] {
        return this.applyWithFunction(sourceFile, walk, [parseOption(this.ruleArguments[0], this.ruleArguments[1])])
    }
}

function parseOption(pattern: string, message: string | undefined): IOption {
    return { message, pattern: new RegExp(`${pattern}`) }
}

function walk(ctx: Lint.WalkContext<IOption[]>): void {
    for (const name of findImports(ctx.sourceFile, ImportKind.All)) {
        const ban = ctx.options.find(({ pattern }) => pattern.test(name.text))
        if (ban) {
            ctx.addFailure(
                name.getStart(ctx.sourceFile) + 1,
                name.end - 1,
                Rule.FAILURE_STRING_FACTORY(ban.pattern.toString(), ban.message)
            )
        }
    }
}
