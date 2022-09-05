import React from 'react'

import { InstanceData } from '../model'
import { Badge } from '@sourcegraph/wildcard'
import classNames from 'classnames'

export const InstanceDetails: React.FunctionComponent<{
    instance: InstanceData
    className?: string
    textClassName?: string
}> = ({ instance, className, textClassName }) => (
    <ul className={classNames(className, 'list-unstyled')}>
        <li>
            <span className={classNames(textClassName)}>
                Owner:{' '}
                <a href={`mailto:${instance.ownerEmail}`} className={classNames(textClassName)}>
                    {instance.ownerEmail}
                </a>
            </span>{' '}
            {instance.viewerIsOwner ? (
                <Badge variant="secondary">Owned by you</Badge>
            ) : instance.viewerIsOrganizationMember ? (
                <Badge variant="secondary">Owned by your organization</Badge>
            ) : null}
        </li>
        <li>
            <code className={classNames(textClassName)}>{instance.id}</code>
        </li>
    </ul>
)
