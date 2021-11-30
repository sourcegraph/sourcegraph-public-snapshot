import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'
import { EntityList } from '../entity-list/EntityList'

interface Props extends TelemetryProps, CatalogEntityFiltersProps {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<Props> = ({ filters, onFiltersChange }) => (
    <EntityList filters={filters} onFiltersChange={onFiltersChange} size="lg" />
)
