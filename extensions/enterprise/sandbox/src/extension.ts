import * as sourcegraph from 'sourcegraph'
// import { registerImportStar } from './importStar'
// import { registerNoInlineProps } from './noInlineProps'
// import { registerDependencyRules } from './dependencyRules'
// import { registerCodeOwnership } from './codeOwnership'
// import { registerTravisGo } from './travisGo'
// import { register as eslintRegister } from './eslint'
import { register as packageJsonDependencyRegister } from './packageJsonDependency'
// import { register as codeDuplicationRegister } from './codeDuplication'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    // ctx.subscriptions.add(eslintRegister())
    // ctx.subscriptions.add(codeDuplicationRegister())
    ctx.subscriptions.add(packageJsonDependencyRegister())
    // ctx.subscriptions.add(registerTravisGo())
    // ctx.subscriptions.add(registerImportStar())
    // ctx.subscriptions.add(registerNoInlineProps())
    // ctx.subscriptions.add(registerDependencyRules())
    // ctx.subscriptions.add(registerCodeOwnership())
    // ctx.subscriptions.add(registerSampleStatusProviders())
}
