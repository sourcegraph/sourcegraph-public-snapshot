import * as sourcegraph from 'sourcegraph'
import { register as findReplaceRegister } from './findReplace'
import { register as javaDependencyRegister } from './javaDependency'
import { register as npmCredentialsRegister } from './npmCredentials/providers'
import { register as packageJsonDependencyRegister } from './packageJsonDependency'
import { register as rubyGemDependencyRegister } from './rubyGemDependency/rubyGemDependency'
import { register as workflowRegister } from './workflow'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(findReplaceRegister())
    ctx.subscriptions.add(javaDependencyRegister())
    ctx.subscriptions.add(npmCredentialsRegister())
    ctx.subscriptions.add(packageJsonDependencyRegister())
    ctx.subscriptions.add(rubyGemDependencyRegister())
    ctx.subscriptions.add(workflowRegister())
}
