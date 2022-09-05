import classNames from 'classnames'
import React from 'react'
import { mdiCloud } from '@mdi/js'

import { InstanceData } from '../model'
import { ButtonLink, Icon } from '@sourcegraph/wildcard'
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
            <code className="text-muted small">{instance.id}</code>
        </header>
        <div style={{ flex: '1' }} />
        <ButtonLink variant="secondary" outline={true} as={Link} className="mr-2">
            Manage
        </ButtonLink>
        <ButtonLink variant="primary" outline={true} as={Link}>
            Sign in
        </ButtonLink>
    </Tag>
)
