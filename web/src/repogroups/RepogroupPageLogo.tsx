import React from 'react'
import { ThemeProps } from '../../../shared/src/theme'

interface Props extends ThemeProps, Exclude<React.ImgHTMLAttributes<HTMLImageElement>, 'src'> {
    icon: string
    text: string
}

/**
 * The Sourcegraph logo image. If a custom logo specified in the `branding` site configuration
 * property, it is used instead.
 */
export const RepogroupPageLogo: React.FunctionComponent<Props> = props => (
    <div className="repogroup-page__logo-container d-flex align-items-center">
        <img {...props} src={props.icon} />
        <span className="h3 font-weight-normal mb-0 ml-1">{props.text}</span>
    </div>
)
