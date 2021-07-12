import * as React from 'react'

import { MagnifyingGlassIcon } from '@sourcegraph/web/src/components/MagnifyingGlassIcon'

import { CtaBanner } from '../../components/CtaBanner'

interface Props {
    className?: string
}

export const SignUpCta: React.FunctionComponent<Props> = ({ className }) => (
    <CtaBanner
        className={className}
        icon={<MagnifyingGlassIcon />}
        title="Improve your workflow"
        bodyText="Sign up to add your code, monitor searches for changes, and access additional search features."
        linkText="Sign up"
        href="/sign-up"
        googleAnalytics={true}
    />
)
