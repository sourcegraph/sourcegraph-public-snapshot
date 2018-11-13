import * as React from 'react'
import { GitHubBlobUrl, GitHubPullUrl, GitHubRepositoryUrl } from '../../libs/github'
import * as github from '../../libs/github/util'
import { OpenInSourcegraphProps } from '../repo'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export class ContextualSourcegraphButton extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        const gitHubState = github.getGitHubState(window.location.href)
        if (!gitHubState) {
            return null
        }

        const { label, openProps, ariaLabel } = this.openOnSourcegraphProps(gitHubState)
        const className = ariaLabel ? 'btn btn-sm tooltipped tooltipped-s' : 'btn btn-sm'
        return <OpenOnSourcegraph openProps={openProps} ariaLabel={ariaLabel} label={label} className={className} />
    }

    private openOnSourcegraphProps(
        state: GitHubBlobUrl | GitHubPullUrl | GitHubRepositoryUrl
    ): { label: string; openProps: OpenInSourcegraphProps; ariaLabel?: string } {
        const props: OpenInSourcegraphProps = {
            repoPath: `${window.location.host}/${state.owner}/${state.repoName}`,
            rev: state.rev || '',
        }
        return {
            label: 'View Repository',
            ariaLabel: 'View repository on Sourcegraph',
            openProps: props,
        }
    }
}
