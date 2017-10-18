import * as React from 'react'

interface Props {
    org: string
}

/**
 * OrgAvatar displays the avatar of an Avatarable object
 */
export class OrgAvatar extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className='org-avatar'>{this.props.org.substr(0, 2).toUpperCase()}</div>
        )
    }
}
