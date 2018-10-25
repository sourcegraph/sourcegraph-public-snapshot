import * as React from 'react'

interface Avatarable {
    avatarURL: string | null
}

interface Props {
    onClick?: () => void
    size?: number
    user: Avatarable
    className?: string
    tooltip?: string
}

/**
 * UserAvatar displays the avatar of an Avatarable object
 */
export const UserAvatar: React.SFC<Props> = props => {
    let avatar: JSX.Element | null = null
    if (props.user && props.user.avatarURL) {
        try {
            const url = new URL(props.user.avatarURL)
            if (props.size) {
                url.searchParams.set('s', props.size + '')
            }
            avatar = <img className="avatar-icon" src={url.href} data-tooltip={props.tooltip} />
        } catch (e) {
            // noop
        }
    }
    if (!avatar) {
        return null
    }

    return (
        <div onClick={props.onClick} className={`avatar${props.className ? ' ' + props.className : ''}`}>
            {avatar}
        </div>
    )
}
