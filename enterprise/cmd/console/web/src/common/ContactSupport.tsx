import React from 'react'

import { Icon } from '@sourcegraph/wildcard'
import { Link } from 'react-router-dom'
import { mdiHelpCircle } from '@mdi/js'
import classNames from 'classnames'

export const ContactSupport: React.FunctionComponent<{
    className?: string
}> = ({ className }) => (
    <p className={classNames(className, 'text-muted')}>
        <Icon aria-hidden={true} svgPath={mdiHelpCircle} className="mr-1" />
        Need help managing an instance? Contact <Link to="mailto:support@sourcegraph.com">support@sourcegraph.com</Link>
        .
    </p>
)
