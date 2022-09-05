import { ButtonLink, H1, H2, Icon } from '@sourcegraph/wildcard'
import React from 'react'
import { Link } from 'react-router-dom'
import { InstanceData } from '../../model'
import { InstanceDetails } from '../InstanceDetails'
import { InstanceHostname } from '../InstanceHostname'
import { mdiArrowLeftThick } from '@mdi/js'

export const InstanceManagePage: React.FunctionComponent<{ instance: InstanceData }> = ({ instance }) => (
    <div>
        <ButtonLink as={Link} variant="secondary" to="/instances" className="mb-3">
            <Icon aria-hidden={true} svgPath={mdiArrowLeftThick} className="icon-inline" />
            View all instances
        </ButtonLink>
        <H1>Manage Sourcegraph Cloud instance</H1>
        <H2>
            <InstanceHostname url={instance.url} />
        </H2>
        <InstanceDetails instance={instance} className="mt-1" />
    </div>
)
