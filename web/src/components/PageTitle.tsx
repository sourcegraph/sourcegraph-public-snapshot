import * as React from 'react'

interface Props {
    title?: string
}

let titleSet = false

export class PageTitle extends React.Component<Props, {}> {
    public componentDidMount(): void {
        if (titleSet) {
            console.error('more than one PageTitle used at the same time')
        }
        titleSet = true
        this.updateTitle(this.props.title)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.updateTitle(nextProps.title)
    }

    public componentWillUnmount(): void {
        titleSet = false
        document.title = 'Sourcegraph'
    }

    public render(): JSX.Element | null {
        return null
    }

    private updateTitle(title?: string): void {
        document.title = title ? `${title} - Sourcegraph` : 'Sourcegraph'
    }
}
