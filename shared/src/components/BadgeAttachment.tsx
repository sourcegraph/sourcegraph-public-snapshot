import * as React from 'react'
import isAbsoluteUrl from 'is-absolute-url'
import { BadgeAttachmentRenderOptions } from 'sourcegraph'
import { badgeAttachmentStyleForTheme } from '../api/client/services/decoration'
import { LinkOrSpan } from './LinkOrSpan'
import { isEncodedImage } from '../util/icon'

export const BadgeAttachment: React.FunctionComponent<{
    attachment: BadgeAttachmentRenderOptions
    isLightTheme: boolean
}> = ({ attachment, isLightTheme }) => {
    const style = badgeAttachmentStyleForTheme(attachment, isLightTheme)

    return (
        <LinkOrSpan
            className="badge-decoration-attachment"
            to={attachment.linkURL}
            data-tooltip={attachment.hoverMessage}
            // Use target to open external URLs
            target={attachment.linkURL && isAbsoluteUrl(attachment.linkURL) ? '_blank' : undefined}
            // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
            rel="noreferrer noopener"
        >
            {style.icon && isEncodedImage(style.icon) && (
                <img
                    className="badge-decoration-attachment__contents"
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{
                        color: style.color,
                        backgroundColor: style.backgroundColor,
                    }}
                    src={style.icon}
                />
            )}
        </LinkOrSpan>
    )
}
