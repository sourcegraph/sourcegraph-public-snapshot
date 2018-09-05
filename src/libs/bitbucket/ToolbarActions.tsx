import * as React from 'react'
import { OpenOnSourcegraph } from '../../shared/components/OpenOnSourcegraph'
import { OpenInSourcegraphProps } from '../../shared/repo'
import { BitbucketState, getRevisionState } from './utils/util'

interface Props {
    filePath: string
    bitbucketState: BitbucketState
}

export class ToolbarActions extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        const { repository } = this.props.bitbucketState
        const revState = getRevisionState(this.props.bitbucketState)
        if (!revState || !repository) {
            return null
        }
        const repoPath = `${window.location.hostname}/${repository.project.key}/${repository.slug}`
        const props: OpenInSourcegraphProps = {
            repoPath,
            filePath: this.props.filePath,
            rev: revState.headRev,
        }
        let margin = '0 10px'
        if (revState.baseRev) {
            margin = ''
            props.commit = {
                baseRev: revState.baseRev,
                headRev: revState.headRev,
            }
        }
        return (
            <div style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}>
                <OpenOnSourcegraph
                    style={{ margin }}
                    className="aui-button"
                    label={revState.baseRev ? 'View Diff' : 'View File'}
                    ariaLabel={revState.baseRev ? 'View diff on Sourcegraph' : 'View file on Sourcegraph'}
                    openProps={props}
                />
            </div>
        )
    }
}
