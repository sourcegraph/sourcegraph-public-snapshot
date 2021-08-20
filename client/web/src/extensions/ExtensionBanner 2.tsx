import classnames from 'classnames'
import React from 'react'

import { CtaBanner } from '../components/CtaBanner'

import { extensionBannerIconURL } from './icons'

interface Props {
    className?: string
}

export const ExtensionBanner: React.FunctionComponent<Props> = ({ className }) => (
    <CtaBanner
        className={classnames(className, 'extension-banner border-0 shadow-none')}
        icon={<img className="extension-banner__icon" src={extensionBannerIconURL} alt="" />}
        headingElement="h2"
        title="Create your own extension"
        bodyText="You can improve your workflow by creating custom extensions. See the Sourcegraph Docs for details about writing and publishing."
        bodyTextClassName="extension-banner__body-text mt-3"
        linkText="Explore extension API"
        href="https://docs.sourcegraph.com/extensions/authoring"
    />
)
