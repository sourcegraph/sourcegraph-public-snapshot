import classNames from 'classnames'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

import { CatalogEntityStateFields, CatalogEntityStatusState } from '../../../../../../graphql-operations'
import { STATE_TO_COLOR } from '../../../entity-detail/global/OverviewStatusContextItem'

export const CatalogEntityStateIndicator: React.FunctionComponent<{
    entity: CatalogEntityStateFields
    className?: string
}> = ({ entity, className }) => (
    <span className={classNames(`text-${STATE_TO_COLOR[entity.status.state]}`, className)}>
        {entity.status.state === CatalogEntityStatusState.SUCCESS ? (
            <CheckBoldIcon className="icon-inline" />
        ) : entity.status.state === CatalogEntityStatusState.FAILURE ||
          entity.status.state === CatalogEntityStatusState.ERROR ? (
            <CloseIcon className="icon-inline" />
        ) : (
            entity.status.state.toLowerCase()
        )}
    </span>
)
