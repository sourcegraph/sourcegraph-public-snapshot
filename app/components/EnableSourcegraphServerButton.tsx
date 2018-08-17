import * as React from 'react'
import * as github from '../github/util'
import { sourcegraphUrl } from '../util/context'
import { Button } from './Button'

export class EnableSourcegraphServerButton extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        const { repoPath, repoName } = github.parseURL()
        if (!repoName) {
            return null
        }
        const isOnlySourcegraph = sourcegraphUrl === 'https://sourcegraph.com'
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
        return (
            <Button
                iconStyle={iconStyle}
                url={url}
                style={style}
                className={className}
                ariaLabel={ariaLabel}
                label={label}
                target="_blank"
            />
        )
    }
}
