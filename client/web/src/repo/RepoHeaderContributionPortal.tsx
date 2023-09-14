import * as React from 'react'

import type { RepoHeaderContribution, RepoHeaderContributionsLifecycleProps } from './RepoHeader'

interface Props extends RepoHeaderContribution, RepoHeaderContributionsLifecycleProps {}

/**
 * Contributes an item to the RepoHeader. The contribution is added and remains present for this component's entire
 * lifecycle.
 *
 * A React component that needs to contribute an item to the RepoHeader should include a
 * <RepoHeaderContributionPortal> element that specifies the item.
 *
 * It is called a "portal" because it is similar to the React concept of a "portal": it effectively renders an
 * element in a different DOM hierarchy (i.e., in RepoHeader).
 */
export class RepoHeaderContributionPortal extends React.Component<Props> {
    public componentDidMount(): void {
        this.addOrUpdateContribution()
    }

    public componentDidUpdate(): void {
        this.addOrUpdateContribution()
    }

    public shouldComponentUpdate(nextProps: Props): boolean {
        return (
            this.props.repoHeaderContributionsLifecycleProps !== nextProps.repoHeaderContributionsLifecycleProps ||
            this.props.position !== nextProps.position ||
            this.props.priority !== nextProps.priority ||
            this.props.children !== nextProps.children
        )
    }

    public componentWillUnmount(): void {
        const id = this.props.id // enforced in RepoHeaderContributionStore
        if (this.props.repoHeaderContributionsLifecycleProps) {
            // Don't need to worry about being unable to remove this because once
            // this.props.repoHeaderContributionsLifecycleProps is set (from RepoHeader), it is never unset.
            this.props.repoHeaderContributionsLifecycleProps.onRepoHeaderContributionRemove(id)
        }
    }

    public render(): JSX.Element | null {
        return null
    }

    private addOrUpdateContribution(): void {
        if (this.props.repoHeaderContributionsLifecycleProps) {
            this.props.repoHeaderContributionsLifecycleProps.onRepoHeaderContributionAdd(this.props)
        }
    }
}
