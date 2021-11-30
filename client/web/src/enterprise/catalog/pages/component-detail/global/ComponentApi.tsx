import classNames from 'classnames'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { useLocation } from 'react-router'
import { Link } from 'react-router-dom'
import { of } from 'rxjs'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { CatalogComponentAPIFields, SymbolFields } from '../../../../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../../../../repo/backend'

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
    const location = useLocation()

    if (!api) {
        return (
            <div className={className}>
                <div className="alert alert-warning">Unable to determine API</div>
            </div>
        )
    }

    const { symbols, schema } = api
    return (
        <>
            <style>
                {
                    'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container { border: solid 1px var(--border-color) !important; border-left: none !important; border-right: none !important; margin: 0; } .result-container small { display: none; } .result-container__header > .mdi-icon { display: none; } .result-container__header-divider { display: none; } .result-container__header { padding-left: 0.25rem; } .FileMatchChildren-module__file-match-children { border: none !important; } .result-container { border: none !important; }'
                }
            </style>
            {schema && schema.__typename === 'GitBlob' && (
                <FileLocations
                    location={location}
                    locations={of<Location[]>([
                        {
                            uri: makeRepoURI({
                                repoName: schema.repository.name,
                                commitID: schema.commit.oid,
                                filePath: schema.path,
                            }),
                            range: { start: { line: 0, character: 0 }, end: { line: 10, character: 0 } },
                        },
                    ])}
                    icon={SourceRepositoryIcon}
                    fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                    settingsCascade={settingsCascade}
                    className={classNames('mb-3')}
                    parentContainerIsEmpty={false}
                    telemetryService={telemetryService}
                />
            )}
            <ol className={classNames('list-group', className)}>
                {symbols.nodes
                    .filter(symbol => !symbol.fileLocal)
                    // .filter(symbol => !symbol.containerName)
                    .filter(
                        // TODO(sqs): hack
                        symbol =>
                            symbol.language === 'TypeScript' || symbol.language === 'Go' || symbol.language === 'tsx'
                    )
                    .map(symbol => (
                        <APISymbol key={symbol.url} symbol={symbol} className="list-group-item" />
                    ))}
            </ol>
        </>
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
