import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CatalogComponentAPIFields, SymbolFields } from '../../../../../graphql-operations'

interface Props extends SettingsCascadeProps, TelemetryProps {
    catalogComponent: CatalogComponentAPIFields
    className?: string
}

export const ComponentAPI: React.FunctionComponent<Props> = ({
    catalogComponent: { api },
    className,
    settingsCascade,
    telemetryService,
}) => {
    if (!api) {
        return (
            <div className={className}>
                <div className="alert alert-warning">Unable to determine API</div>
            </div>
        )
    }

    const { symbols } = api
    return symbols && symbols.nodes.length > 0 ? (
        <ol className={classNames('list-group', className)}>
            {symbols.nodes
                .filter(symbol => !symbol.fileLocal)
                // .filter(symbol => !symbol.containerName)
                .filter(
                    // TODO(sqs): hack
                    symbol => symbol.language === 'TypeScript' || symbol.language === 'Go' || symbol.language === 'tsx'
                )
                .map(symbol => (
                    <APISymbol key={symbol.url} symbol={symbol} className="list-group-item" />
                ))}
        </ol>
    ) : (
        <p>No uses found</p>
    )
}

const APISymbol: React.FunctionComponent<{
    symbol: SymbolFields
    tag?: 'li'
    className?: string
}> = ({ symbol, tag: Tag = 'li', className }) => (
    <Tag className={className}>
        <Link to={symbol.url} className="d-flex align-items-center">
            <SymbolIcon kind={symbol.kind} className="icon-inline mr-1" />
            <span className={classNames('')}>{symbol.name}</span>
            {symbol.containerName && <small className="text-muted ml-1">{symbol.containerName}</small>}
        </Link>
    </Tag>
)
