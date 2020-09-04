import React from 'react'
import { BannerPuzzleIcon } from './icons'
import { CtaBanner } from '../components/CtaBanner'

interface Props {
    className?: string
}

export const ExtensionBanner: React.FunctionComponent<Props> = ({ className }) => (
    <CtaBanner
        className={className}
        icon={<BannerPuzzleIcon />}
        title="Create your own extension"
        bodyText="You can improve your workflow by creating custom extensions. See the Sourcegraph Docs for details about writing and publishing."
        linkText="Explore extension API"
        href="https://docs.sourcegraph.com/extensions/authoring"
    />
)
