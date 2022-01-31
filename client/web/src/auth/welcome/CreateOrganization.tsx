import classNames from 'classnames'
import React, { useState } from 'react'

import { Button, ProductStatusBadge } from '@sourcegraph/wildcard'

import styles from './CreateOrganization.module.scss'
import { useHubSpotForm } from './useHubSpotForm'

interface CreateOrganization {}

export const CreateOrganization: React.FunctionComponent<CreateOrganization> = () => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const form = useHubSpotForm({ portalId: '2762526', formId: 'e0e43746-83e9-4133-97bd-9954a60c7af8' })

    return (
        <div className="mt-2 w-100">
            <h3>
                Create an organization (optional) <ProductStatusBadge status="beta" className="text-uppercase mr-1" />
            </h3>
            <p className="text-muted">
                Teams on Sourcegraph Cloud will be the quickest way to level up your team with powerful code search.
            </p>
            <div className={styles.content}>
                {isExpanded ? (
                    <p>Complete the form below and we’ll reach out to discuss the early beta.</p>
                ) : (
                    <>
                        <p>Would you like to be added to the teams beta?</p>
                        <Button onClick={() => setIsExpanded(true)} variant="success">
                            Apply
                        </Button>
                    </>
                )}

                <div
                    className={classNames({
                        [styles.form]: true,
                        [styles.formExpanded]: isExpanded,
                    })}
                >
                    {form}
                </div>
            </div>
        </div>
    )
}
