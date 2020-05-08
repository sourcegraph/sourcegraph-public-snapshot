import * as React from 'react'
import isAbsoluteUrl from 'is-absolute-url'
import InformationIcon from 'mdi-react/InfoCircleOutlineIcon'
import WarningIcon from 'mdi-react/AlertCircleOutlineIcon'
import ErrorIcon from 'mdi-react/AlertDecagramOutlineIcon'
import { BadgeAttachmentRenderOptions } from 'sourcegraph'
import { badgeAttachmentStyleForTheme } from '../api/client/services/decoration'
import { LinkOrSpan } from './LinkOrSpan'
import { isEncodedImage } from '../util/icon'
import { MdiReactIconComponentType } from 'mdi-react'
import classNames from 'classnames'

const chooseIconComponent = (icon: BadgeAttachmentRenderOptions['kind']): MdiReactIconComponentType => {
    switch (icon) {
        case 'info':
            return InformationIcon
        case 'warning':
            return WarningIcon
        case 'error':
            return ErrorIcon
    }
}

const isPredefinedIcon = (badge: BadgeAttachmentRenderOptions): boolean => 'kind' in badge

const renderIcon = (badge: BadgeAttachmentRenderOptions, isLightTheme: boolean): JSX.Element | null => {
    if (isPredefinedIcon(badge)) {
        const Icon = chooseIconComponent(badge.kind)
        return <Icon className={classNames('icon-inline', 'badge-decoration-attachment__icon-svg')} />
    }

    const style = badgeAttachmentStyleForTheme(badge, isLightTheme)

    if (!style.icon || !isEncodedImage(style.icon)) {
        return null
    }

    return (
        <img
            className="badge-decoration-attachment__contents"
            // eslint-disable-next-line react/forbid-dom-props
            style={{
                color: style.color,
                backgroundColor: style.backgroundColor,
            }}
            src={style.icon}
        />
    )
}

export const BadgeAttachment: React.FunctionComponent<{
    attachment: BadgeAttachmentRenderOptions
    isLightTheme: boolean
}> = ({ attachment, isLightTheme }) => (
    // 'badge-decoration-attachment__contents'
    <LinkOrSpan
        className={classNames('badge-decoration-attachment', isPredefinedIcon(attachment) && 'btn-icon')}
        to={attachment.linkURL}
        data-tooltip={attachment.hoverMessage}
        data-placement="left"
        // Use target to open external URLs
        target={attachment.linkURL && isAbsoluteUrl(attachment.linkURL) ? '_blank' : undefined}
        // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
        rel="noreferrer noopener"
    >
        {renderIcon(attachment, isLightTheme)}
    </LinkOrSpan>
)
