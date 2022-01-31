import classNames from 'classnames'
import React, { useState } from 'react'

import { Button, ProductStatusBadge } from '@sourcegraph/wildcard'

import styles from './CreateOrganization.module.scss'

interface CreateOrganization {}

export const CreateOrganization: React.FunctionComponent<CreateOrganization> = () => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    return (
        <div className="mt-2 w-100">
            <h3>
                Create an organization (optional) <ProductStatusBadge status="beta" className="text-uppercase mr-1" />
            </h3>
            <p className="text-muted">
                Teams on Sourcegraph Cloud will be the quickest way to level up your team with powerful code search.
            </p>
            <div
                className={classNames({
                    [styles.content]: true,
                    [styles.expanded]: isExpanded,
                })}
            >
                <p> Would you like to be added to the teams beta?</p>
                <Button onClick={() => setIsExpanded(a => !a)} variant="success">
                    Apply
                </Button>
            </div>
        </div>
    )
}
