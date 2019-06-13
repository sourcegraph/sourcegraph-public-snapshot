import * as sourcegraph from 'sourcegraph'
import { registerImportStar } from './importStar'
import { registerNoInlineProps } from './noInlineProps'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(registerImportStar())
    ctx.subscriptions.add(registerNoInlineProps())
}
