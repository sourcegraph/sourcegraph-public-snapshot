import * as sourcegraph from 'sourcegraph'
import { registerImportStar } from './importStar'
import { registerNoInlineProps } from './noInlineProps'
import { registerDependencyRules } from './dependencyRules'
import { registerCodeOwnership } from './codeOwnership'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(registerImportStar())
    ctx.subscriptions.add(registerNoInlineProps())
    ctx.subscriptions.add(registerDependencyRules())
    ctx.subscriptions.add(registerCodeOwnership())
}
