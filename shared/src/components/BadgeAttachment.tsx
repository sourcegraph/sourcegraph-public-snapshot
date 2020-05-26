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

const iconComponents: Record<BadgeAttachmentRenderOptions['kind'], MdiReactIconComponentType> = {
    info: InformationIcon,
    warning: WarningIcon,
    error: ErrorIcon,
}

const renderIcon = (badge: BadgeAttachmentRenderOptions, isLightTheme: boolean): JSX.Element | null => {
    if ('kind' in badge) {
        // means that we are using predefined icons
        const Icon = iconComponents[badge.kind]
        return <Icon className="icon-inline badge-decoration-attachment__icon-svg" />
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
    <LinkOrSpan
        className="badge-decoration-attachment"
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
