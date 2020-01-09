import React from 'react'
import combyJsonSchema from '../../../../../../schema/campaign-types/comby.schema.json'
import credentialsJsonSchema from '../../../../../../schema/campaign-types/credentials.schema.json'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { MonacoSettingsEditor } from '../../../../settings/MonacoSettingsEditor'
import { CampaignType } from '../backend.js'

interface Props extends ThemeProps {
    /** The campaign plan specification type (e.g., "comby"). */
    type: CampaignType | undefined
    onTypeChange: (newType: CampaignType) => void

    /** The campaign plan specification arguments (as JSONC). */
    argumentsJSONC: string | undefined
    onArgumentsJSONCChange: (newArguments: string) => void

    readOnly?: boolean
    className?: string
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const jsonSchemaByType: { [K in CampaignType]: any } = {
    comby: combyJsonSchema,
    credentials: credentialsJsonSchema,
}

const typeLabels: Record<CampaignType | '', string> = {
    '': 'Manual',
    comby: 'Comby search and replace',
    credentials: 'Find leaked credentials',
}

/**
 * Fields for selecting the type and arguments for the campaign plan specification.
 */
export const CampaignPlanSpecificationFields: React.FunctionComponent<Props> = ({
    type,
    onTypeChange,
    argumentsJSONC,
    onArgumentsJSONCChange,
    readOnly,
    className,
    isLightTheme,
}) => (
    <div className={className}>
        <div className="row campaign-details__property-row">
            <h3 className="mr-3 mb-0 campaign-details__property-label">Type</h3>
            <div className="flex-grow-1 form-group mb-0">
                {!readOnly ? (
                    <>
                        <select
                            className="form-control w-auto d-inline-block e2e-campaign-type"
                            placeholder="Select campaign type"
                            onChange={e => onTypeChange(e.currentTarget.value as CampaignType)}
                            value={type}
                        >
                            {(Object.keys(typeLabels) as CampaignType[]).map(typeName => (
                                <option value={typeName || ''} key={typeName}>
                                    {typeLabels[typeName]}
                                </option>
                            ))}
                        </select>
                        {type === 'comby' && (
                            <small className="ml-1">
                                <a rel="noopener noreferrer" target="_blank" href="https://comby.dev/#match-syntax">
                                    Learn about comby syntax
                                </a>
                            </small>
                        )}
                    </>
                ) : (
                    <p className="mb-0">{typeLabels[type || '']}</p>
                )}
            </div>
        </div>
        {type && (
            <div className="row mt-3">
                <h3 className="mr-3 mb-0 flex-grow-0 campaign-details__property-label">Arguments</h3>
                <MonacoSettingsEditor
                    className="flex-grow-1 e2e-campaign-arguments"
                    isLightTheme={isLightTheme}
                    value={argumentsJSONC}
                    jsonSchema={type ? jsonSchemaByType[type] : undefined}
                    height={110}
                    onChange={onArgumentsJSONCChange}
                    readOnly={readOnly}
                />
            </div>
        )}
    </div>
)
