import { Subscription, Unsubscribable } from 'rxjs'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import H from 'history'
import { treeReadme } from './treeReadme'
import { treeCommits } from './treeCommits'
import { repoBranches } from './repoBranches'
import { repoTags } from './repoTags'
import { treeDependencies } from './treeDependencies'

export const registerTreeViews = ({
    extensionsController: { services },
    history,
}: ExtensionsControllerProps & { history: H.History }): Unsubscribable => {
    const subscription = new Subscription()

    subscription.add(
        services.view.register('treeView.branches', ContributableViewContainer.Directory, context =>
            repoBranches(context)
        )
    )

    subscription.add(
        services.view.register('treeView.tags', ContributableViewContainer.Directory, context => repoTags(context))
    )

    subscription.add(
        services.view.register('treeView.commits', ContributableViewContainer.Directory, context =>
            treeCommits(context)
        )
    )

    subscription.add(
        services.view.register('treeView.dependencies', ContributableViewContainer.Directory, context =>
            treeDependencies(context)
        )
    )

    subscription.add(
        services.view.register('treeView.readme', ContributableViewContainer.Directory, context =>
            treeReadme(context, history)
        )
    )

    return subscription
}
