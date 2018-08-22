import * as React from 'react'
import * as runtime from '../../extension/runtime'
import * as github from '../github/util'
import { isSourcegraphDotCom, sourcegraphUrl } from '../util/context'
import * as featureFlags from '../util/featureFlags'
import { Button } from './Button'

interface State {
    canOpenOptionsPage: boolean
}

export class ConfigureSourcegraphButton extends React.Component<{}, State> {
    public state: State = {
        canOpenOptionsPage: false,
    }

    public componentDidMount(): void {
        featureFlags
            .get('optionsPage')
            .then(canOpenOptionsPage => {
                this.setState(() => ({ canOpenOptionsPage }))
            })
            .catch(err => console.error('could not get feature flag', err))
    }

    private handleOpenOptionsPage = (): void => {
        runtime.sendMessage({ type: 'openOptionsPage' })
    }

    public render(): JSX.Element | null {
        const { repoPath, repoName } = github.parseURL()
        if (!repoName) {
            return null
        }
        const isOnlySourcegraph = isSourcegraphDotCom()
        const label = isOnlySourcegraph ? 'Configure Sourcegraph' : 'View Repository'
        const ariaLabel = isOnlySourcegraph
            ? 'Install Sourcegraph for search and code intelligence on private repositories'
            : 'View Repository on Sourcegraph'
        const className = isOnlySourcegraph
            ? 'btn btn-sm tooltipped tooltipped-s muted'
            : 'btn btn-sm tooltipped tooltipped-s'
        const style = isOnlySourcegraph ? { border: 'none', background: 'none' } : undefined
        const iconStyle = isOnlySourcegraph
            ? { filter: 'grayscale(100%)', marginTop: '-1px', paddingRight: '4px', fontSize: '18px' }
            : undefined
        const url = isOnlySourcegraph ? 'https://about.sourcegraph.com' : `${sourcegraphUrl}/${repoPath}`
        const openOptionsPage = this.state.canOpenOptionsPage && isOnlySourcegraph

        return (
            <Button
                iconStyle={iconStyle}
                url={openOptionsPage ? undefined : url}
                onClick={openOptionsPage ? this.handleOpenOptionsPage : undefined}
                style={style}
                className={className}
                ariaLabel={ariaLabel}
                label={label}
                target="_blank"
            />
        )
    }
}
