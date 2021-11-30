import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container } from '@sourcegraph/wildcard'

import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'
import { EntityList } from '../entity-list/EntityList'

import { OverviewEntityGraph } from './OverviewEntityGraph'

interface Props extends TelemetryProps, CatalogEntityFiltersProps {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<Props> = ({ filters, onFiltersChange }) => (
    <div className="d-flex flex-column">
        <Container className="p-0 mb-2">
            <EntityList filters={filters} onFiltersChange={onFiltersChange} size="sm" />
        </Container>
        <OverviewEntityGraph />
    </div>
)
