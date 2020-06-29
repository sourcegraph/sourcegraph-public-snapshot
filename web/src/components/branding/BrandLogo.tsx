import React from 'react'
import { ThemeProps } from '../../../../shared/src/theme'

interface Props extends ThemeProps, Exclude<React.ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    /**
     * The site configuration `branding` property. If not set, the global value from
     * `window.context.branding` is used.
     */
    branding?: typeof window.context.branding

    /**
     * The assets root path. If not set, the global value from `window.context.assetsRoot` is used.
     */
    assetsRoot?: typeof window.context.assetsRoot

    /**
     * A url for a custom logo. This is passed in from parent components for changing the logo for individual pages.
     * For changing the instance-wide default logo use the `branding` prop.
     */
    customLogoUrl?: string
}

/**
 * The Sourcegraph logo image. If a custom logo specified in the `branding` site configuration
 * property, it is used instead.
 */
export const BrandLogo: React.FunctionComponent<Props> = ({
    isLightTheme,
    branding = window.context?.branding,
    assetsRoot = window.context?.assetsRoot || '',
    ...props
}) => {
    const sourcegraphLogoUrl = `${assetsRoot}/img/sourcegraph${isLightTheme ? '-light' : ''}-head-logo.svg`
    const customBrandingLogoUrl = branding && branding[isLightTheme ? 'light' : 'dark']?.logo
    return <img {...props} src={props.customLogoUrl || customBrandingLogoUrl || sourcegraphLogoUrl} />
}
