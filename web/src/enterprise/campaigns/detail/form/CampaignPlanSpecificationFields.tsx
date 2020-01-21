import React, { useCallback, useEffect } from 'react'
import combyJsonSchema from '../../../../../../schema/campaign-types/comby.schema.json'
import credentialsJsonSchema from '../../../../../../schema/campaign-types/credentials.schema.json'
import regexSearchReplaceJsonSchema from '../../../../../../schema/campaign-types/regex_search_replace.schema.json'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { MonacoSettingsEditor } from '../../../../settings/MonacoSettingsEditor'
import { CampaignType } from '../backend'
import { MANUAL_CAMPAIGN_TYPE, campaignTypeLabels } from '../presentation'

/**
 * Data represented in {@link CampaignPlanSpecificationFields}.
 */
export interface CampaignPlanSpecificationFormData {
    /** The campaign plan specification type (e.g., "comby"). */
    type: CampaignType | typeof MANUAL_CAMPAIGN_TYPE

    /** The campaign plan specification arguments (as JSONC). */
    arguments: string
}

interface Props extends ThemeProps {
    value: CampaignPlanSpecificationFormData | undefined
    onChange: (newValue: CampaignPlanSpecificationFormData) => void

    readOnly?: boolean
    className?: string
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const jsonSchemaByType: { [K in CampaignType]: any } = {
    comby: combyJsonSchema,
    credentials: credentialsJsonSchema,
    regexSearchReplace: regexSearchReplaceJsonSchema,
}

const defaultInputByType: { [K in CampaignType]: string } = {
    comby: `{
    "scopeQuery": "repo:github.com/foo/bar",
    "matchTemplate": "",
    "rewriteTemplate": ""
}`,
    credentials: `{
    "scopeQuery": "repo:github.com/foo/bar",
    "matchers": [{ "type": "npm" }]
}`,
    regexSearchReplace: `{
    "scopeQuery": "repo:github.com/foo/bar file:.*",
    "regexpMatch": "foo",
    "textReplace": "bar"
}`,
}

/**
 * Fields for selecting the type and arguments for the campaign plan specification.
 */
export const CampaignPlanSpecificationFields: React.FunctionComponent<Props> = ({
    value: rawValue,
    onChange,
    readOnly,
    className,
    isLightTheme,
}) => {
    const value: CampaignPlanSpecificationFormData =
        rawValue !== undefined
            ? rawValue
            : { type: 'regexSearchReplace', arguments: defaultInputByType.regexSearchReplace }
    useEffect(() => {
        if (rawValue === undefined) {
            onChange(value)
        }
    }, [onChange, rawValue, value])

    const onTypeChange = useCallback(
        (type: CampaignType | typeof MANUAL_CAMPAIGN_TYPE): void => {
            onChange({ type, arguments: type === MANUAL_CAMPAIGN_TYPE ? '' : defaultInputByType[type] })
        },
        [onChange]
    )
    const onArgumentsChange = useCallback((arguments_: string): void => onChange({ ...value, arguments: arguments_ }), [
        onChange,
        value,
    ])

    return (
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
                                value={value.type}
                            >
                                {(Object.keys(campaignTypeLabels) as CampaignType[]).map(typeName => (
                                    <option value={typeName || ''} key={typeName}>
                                        {campaignTypeLabels[typeName]}
                                    </option>
                                ))}
                            </select>
                            {value.type === 'comby' && (
                                <small className="ml-1">
                                    <a
                                        rel="noopener noreferrer"
                                        target="_blank"
                                        tabIndex={-1}
                                        href="https://comby.dev/#match-syntax"
                                    >
                                        Learn about comby syntax
                                    </a>
                                </small>
                            )}
                        </>
                    ) : (
                        <p className="mb-0">{campaignTypeLabels[value.type || '']}</p>
                    )}
                </div>
            </div>
            {value.type !== MANUAL_CAMPAIGN_TYPE && (
                <div className="row mt-3">
                    <h3 className="mr-3 mb-0 flex-grow-0 campaign-details__property-label">Arguments</h3>
                    <MonacoSettingsEditor
                        className="flex-grow-1 e2e-campaign-arguments"
                        isLightTheme={isLightTheme}
                        value={value.arguments}
                        jsonSchema={value.type ? jsonSchemaByType[value.type] : undefined}
                        height={110}
                        onChange={onArgumentsChange}
                        readOnly={readOnly}
                    />
                </div>
            )}
        </div>
    )
}
