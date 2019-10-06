import * as sourcegraph from 'sourcegraph'
import { register as packageJsonDependencyRegister } from './packageJsonDependency'
import { register as rubyGemDependencyRegister } from './rubyGemDependency/rubyGemDependency'
import { register as findReplaceRegister } from './findReplace'

export function activate(ctx: sourcegraph.ExtensionContext): void {
    ctx.subscriptions.add(rubyGemDependencyRegister())
    ctx.subscriptions.add(packageJsonDependencyRegister())
    ctx.subscriptions.add(findReplaceRegister())
}
