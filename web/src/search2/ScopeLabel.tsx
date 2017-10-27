import * as React from 'react'

interface Props {
    /** The query value of the active search scope, or undefined if it's still loading */
    scopeQuery?: string
}

/**
 * The label showing the scope query
 */
export class ScopeLabel extends React.Component<Props> {

    public render(): JSX.Element | null {
        const label = 'Scoped to: '
        let title: string | undefined

        const parts: React.ReactChild[] = []

        if (this.props.scopeQuery === '') {
            parts.push('Searching all code')
        } else {
            parts.push(label)
            if (this.props.scopeQuery === undefined) {
                parts.push('loading...')
            } else {
                title = `The scope query is merged with the user-provided query to perform the search`
                parts.push(<span key='1' className='scope-label2__query'>{this.props.scopeQuery}</span>)
            }
        }

        return (
            <div
                className='scope-label2'
                title={title}
            >
                {parts}
            </div>
        )
    }
}
