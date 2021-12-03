import classNames from 'classnames'
import React from 'react'

import { CatalogEntityStateFields, CatalogEntityStatusState } from '../../../../../../graphql-operations'
import { STATE_TO_COLOR } from '../../../entity-detail/global/OverviewStatusContextItem'

export const CatalogEntityStateIndicator: React.FunctionComponent<{
    entity: CatalogEntityStateFields
    className?: string
}> = ({ entity, className }) => (
    <span className={classNames(`ml-2 text-${STATE_TO_COLOR[entity.status.state]}`, className)}>
        {entity.status.state === CatalogEntityStatusState.SUCCESS
            ? '\u2713'
            : entity.status.state === CatalogEntityStatusState.FAILURE ||
              entity.status.state === CatalogEntityStatusState.ERROR
            ? '\u00D7'
            : entity.status.state.toLowerCase()}
    </span>
)
