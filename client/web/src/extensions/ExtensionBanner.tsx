import classNames from 'classnames'
import React from 'react'

import { CtaBanner } from '../components/CtaBanner'

import styles from './ExtensionBanner.module.scss'
import { extensionBannerIconURL } from './icons'

interface Props {
    className?: string
}

export const ExtensionBanner: React.FunctionComponent<Props> = ({ className }) => (
    <CtaBanner
        className={classNames('border-0 shadow-none', styles.extensionBanner, className)}
        icon={<img className={styles.icon} src={extensionBannerIconURL} alt="" />}
        headingElement="h2"
        title="Create your own extension"
        bodyText="You can improve your workflow by creating custom extensions. See the Sourcegraph Docs for details about writing and publishing."
        bodyTextClassName={classNames('mt-3', styles.bodyText)}
        linkText="Explore extension API"
        href="https://docs.sourcegraph.com/extensions/authoring"
    />
)
