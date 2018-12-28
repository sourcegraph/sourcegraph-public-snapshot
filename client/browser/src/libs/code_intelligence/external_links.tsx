import * as React from 'react'
import { render } from 'react-dom'
import { Button } from '../../shared/components/Button'
import { CodeHost, CodeHostContext } from './code_intelligence'

interface ViewOnSourcegraphButtonProps {
    context: CodeHostContext
    sourcegraphUrl: string
    className?: string
}

class ViewOnSourcegraphButton extends React.Component<ViewOnSourcegraphButtonProps> {
    public render(): React.ReactNode {
        return (
            <Button
                url={this.getURL()}
                label="View Repository"
                ariaLabel="View repository on Sourcegraph"
                className={`open-on-sourcegraph ${this.props.className || ''}`}
            />
        )
    }

    private getURL(): string {
        const rev = this.props.context.rev ? `@${this.props.context.rev}` : ''

        return `${this.props.sourcegraphUrl}/${this.props.context.repoName}${rev}`
    }
}

export function injectViewContextOnSourcegraph(
    sourcegraphUrl: string,
    {
        getContext,
        getViewContextOnSourcegraphMount,
        contextButtonClassName,
    }: Pick<CodeHost, 'getContext' | 'getViewContextOnSourcegraphMount' | 'contextButtonClassName'>
): void {
    if (!getContext || !getViewContextOnSourcegraphMount) {
        return
    }

    const mount = getViewContextOnSourcegraphMount()

    render(
        <ViewOnSourcegraphButton
            context={getContext()}
            className={contextButtonClassName}
            sourcegraphUrl={sourcegraphUrl}
        />,
        mount
    )
}
