import type { FC } from 'react'

import classNames from 'classnames'

import styles from './BrandLogo.module.scss'

interface BrandLogoProps extends Exclude<React.ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    /**
     * The site configuration `branding` property. If not set, the global value from
     * `window.context.branding` is used.
     */
    branding?: typeof window.context.branding

    /**
     * The assets root path. If not set, the global value from `window.context.assetsRoot` is used.
     */
    assetsRoot?: typeof window.context.assetsRoot

    /** Whether to show the full logo (with text) or only the symbol icon. */
    variant: 'logo' | 'symbol'

    isLightTheme: boolean

    /** Whether not to add styles for the spinning effect on hover. */
    disableSymbolSpin?: boolean
}

/**
 * The Sourcegraph logo image. If a custom logo specified in the `branding` site configuration
 * property, it is used instead.
 */
export const BrandLogo: FC<BrandLogoProps> = props => {
    const {
        branding = window.context?.branding,
        assetsRoot = window.context?.assetsRoot || '',
        variant,
        className,
        isLightTheme,
        disableSymbolSpin,
        ...attrs
    } = props

    const themeProperty = isLightTheme ? 'light' : 'dark'

    const sourcegraphLogoUrl =
        variant === 'symbol'
            ? // When changed, update cmd/frontend/internal/app/ui/handlers.go for proper preloading
              `${assetsRoot}/img/sourcegraph-mark.svg?v2` // Add query parameter for cache busting.
            : `${assetsRoot}/img/sourcegraph-logo-${themeProperty}.svg`

    const customBrandingLogoUrl = branding?.[themeProperty]?.[variant]

    return (
        <img
            {...attrs}
            className={classNames(className, {
                [styles.brandLogoSpin]: variant === 'symbol' && !branding?.disableSymbolSpin && !disableSymbolSpin,
            })}
            src={customBrandingLogoUrl || sourcegraphLogoUrl}
            alt={customBrandingLogoUrl ? 'Logo' : 'Sourcegraph logo'}
        />
    )
}
