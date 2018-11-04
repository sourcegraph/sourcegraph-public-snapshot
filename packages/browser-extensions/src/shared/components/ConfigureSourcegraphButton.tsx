import * as React from 'react'
import * as runtime from '../../browser/runtime'
import * as github from '../../libs/github/util'
import { isSourcegraphDotCom } from '../util/context'
import { Button } from './Button'

export class ConfigureSourcegraphButton extends React.Component<{}, {}> {
    private handleOpenOptionsPage = (): void => {
        runtime.sendMessage({ type: 'openOptionsPage' })
    }

    public render(): JSX.Element | null {
        const { repoName } = github.parseURL()
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

        return (
            <Button
                iconStyle={iconStyle}
                onClick={this.handleOpenOptionsPage}
                style={style}
                className={className}
                ariaLabel={ariaLabel}
                label={label}
                target="_blank"
            />
        )
    }
}
