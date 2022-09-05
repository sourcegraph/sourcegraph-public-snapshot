import classNames from 'classnames'
import React from 'react'
import { mdiCloud, mdiClock } from '@mdi/js'

import { InstanceData } from '../../model'
import { Badge, ButtonLink, Icon, LoadingSpinner } from '@sourcegraph/wildcard'
import { Link } from 'react-router-dom'

export const InstanceListItem: React.FunctionComponent<{
    instance: InstanceData
    tag?: 'li'
    className?: string
}> = ({ instance, tag: Tag = 'li', className }) => (
    <Tag className={classNames(className, 'd-flex', 'align-items-start')}>
        <Icon aria-hidden={true} svgPath={mdiCloud} size="md" className="mr-3 text-muted" />
        <header>
            <h3 className="mb-0">
                <a href={instance.url} className="text-body">
                    <span>{instance.url.replace('https://', '').replace('.sourcegraph.com', '')}</span>
                    <span className="font-weight-normal">.sourcegraph.com</span>
                </a>
            </h3>
            <ul className="list-unstyled mt-1">
                <li className="text-muted">
                    <small>
                        Owner:{' '}
                        <a href={`mailto:${instance.ownerEmail}`} className="text-muted">
                            {instance.ownerEmail}
                        </a>
                    </small>{' '}
                    {instance.viewerIsOwner ? (
                        <Badge variant="secondary">Owned by you</Badge>
                    ) : instance.viewerIsOrganizationMember ? (
                        <Badge variant="secondary">Owned by your organization</Badge>
                    ) : null}
                </li>
                <li>
                    <code className="text-muted small">{instance.id}</code>
                </li>
            </ul>
        </header>
        <div style={{ flex: '1' }} className="ml-2" />
        {instance.status === 'ready' ? (
            <>
                {instance.viewerIsOwner ? (
                    <ButtonLink variant="secondary" outline={true} as={Link}>
                        Manage
                    </ButtonLink>
                ) : (
                    <a href={`mailto:${instance.ownerEmail}`} className="text-muted py-2 small mr-2">
                        Contact owner to manage
                    </a>
                )}
                <ButtonLink variant="primary" outline={true} as={Link} className="ml-2">
                    Sign in
                </ButtonLink>
            </>
        ) : (
            <div>
                <Badge variant="secondary" className="d-flex align-items-center justify-content-center mb-2 p-2">
                    <LoadingSpinner className="icon-inline mr-1" />
                    Creating...
                </Badge>
                <div className="text-muted d-flex align-items-center small">
                    <Icon aria-hidden={true} svgPath={mdiClock} size="md" className="mr-1" /> 30 minutes remaining
                </div>
            </div>
        )}
    </Tag>
)
