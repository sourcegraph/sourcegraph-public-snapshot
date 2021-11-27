import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentFiltersProps } from '../../../../core/component-filters'
import { ComponentList } from '../component-list/ComponentList'

interface Props extends TelemetryProps, CatalogComponentFiltersProps {
    // TODO(sqs): what scope of catalog (eg repo) or global
}

export const OverviewContent: React.FunctionComponent<Props> = ({ filters, onFiltersChange }) => (
    <ComponentList filters={filters} onFiltersChange={onFiltersChange} size="lg" />
)
