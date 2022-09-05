import classNames from 'classnames'
import React from 'react'
import { mdiCloud, mdiClock } from '@mdi/js'

import { InstanceData } from '../../model'
import { Badge, ButtonLink, Icon, LoadingSpinner } from '@sourcegraph/wildcard'
import { Link } from 'react-router-dom'
import styles from './InstanceListItem.module.scss'
import { InstanceHostname } from '../InstanceHostname'
import { InstanceDetails } from '../InstanceDetails'

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
                    <InstanceHostname url={instance.url} />
                </a>
            </h3>
            <InstanceDetails instance={instance} className="mt-1" textClassName="small text-muted" />
        </header>
        <div style={{ flex: '1' }} className="ml-2" />
        {instance.status === 'ready' ? (
            <>
                {instance.viewerIsOwner ? (
                    <ButtonLink variant="secondary" outline={true} as={Link} to={`/instances/${instance.id}`}>
                        Manage
                    </ButtonLink>
                ) : null}
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
