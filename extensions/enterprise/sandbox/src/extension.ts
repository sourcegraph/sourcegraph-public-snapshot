import * as sourcegraph from 'sourcegraph'
import { register as findReplaceRegister } from './findReplace'
import { register as npmCredentialsRegister } from './npmCredentials/providers'
import { register as packageJsonDependencyRegister } from './packageJsonDependency/packageJsonDependency'
import { register as rubyGemDependencyRegister } from './rubyGemDependency/rubyGemDependency'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(findReplaceRegister())
    ctx.subscriptions.add(npmCredentialsRegister())
    ctx.subscriptions.add(packageJsonDependencyRegister())
    ctx.subscriptions.add(rubyGemDependencyRegister())
}
