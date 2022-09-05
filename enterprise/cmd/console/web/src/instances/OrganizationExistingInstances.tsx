import { Button, ButtonLink } from '@sourcegraph/wildcard'
import ArrowRightThickIcon from 'mdi-react/ArrowRightThickIcon'
import classNames from 'classnames'
import React from 'react'
import { InstanceData } from '../model'

const EXISTING_INSTANCES: InstanceData[] = [
    {
        id: 'c-f8a7d6ebfa8374ace',
        title: 'Acme Corp',
        ownerEmail: 'alice@acme-corp.com',
        ownerName: 'Alice Smith',
        url: 'https://acme.sourcegraph.com',
    },
    {
        id: 'c-c9b7a7e739a6cd7cb',
        title: 'Acme Labs',
        ownerEmail: 'zhao@acme-corp.com',
        ownerName: 'Fangfang Zhao',
        url: 'https://acmelabs.sourcegraph.com',
    },
]

export const OrganizationExistingInstances: React.FunctionComponent = () => (
    <div>
        <ul className="list-group">
            {EXISTING_INSTANCES.map(instance => (
                <ExistingInstance key={instance.id} instance={instance} className="list-group-item" />
            ))}
        </ul>
    </div>
)

const ExistingInstance: React.FunctionComponent<{ instance: InstanceData; className: string }> = ({
    instance,
    className,
}) => (
    <li className={classNames(className, 'd-flex', 'align-items-center', 'justify-content-between', 'p-3')}>
        <header className="flex-1">
            <h4 className="mb-1 font-weight-bold">{instance.title}</h4>
            <span className="text-muted">
                {instance.ownerName} ({instance.ownerEmail})
            </span>
        </header>
        <ButtonLink variant="primary">
            Join <ArrowRightThickIcon className="icon-inline" />
        </ButtonLink>
    </li>
)
