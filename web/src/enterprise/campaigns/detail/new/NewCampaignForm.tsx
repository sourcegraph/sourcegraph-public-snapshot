import React, { useCallback } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignDescriptionField } from '../form/CampaignDescriptionField'
import { CampaignTitleField } from '../form/CampaignTitleField'
import {
    CampaignPlanSpecificationFields,
    CampaignPlanSpecificationFormData,
} from '../form/CampaignPlanSpecificationFields'
import { ThemeProps } from '../../../../../../shared/src/theme'

/**
 * Data represented in {@link NewCampaignForm}.
 */
export interface NewCampaignFormData extends Pick<GQL.ICampaign, 'name' | 'description'> {
    /**
     * An existing campaign plan to use, or the plan specification to create the plan, or undefined
     * if none is specified yet.
     */
    plan: GQL.ID | CampaignPlanSpecificationFormData
}

export const isExistingPlanID = (plan: NewCampaignFormData['plan']): plan is GQL.ID => typeof plan === 'string'

interface Props extends ThemeProps {
    value: NewCampaignFormData
    onChange: (newValue: NewCampaignFormData) => void

    disabled?: boolean
    className?: string
}

export const NewCampaignForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
    className = '',
    isLightTheme,
    children,
}) => {
    const onNameChange = useCallback((name: string): void => onChange({ ...value, name }), [onChange, value])
    const onDescriptionChange = useCallback((description: string): void => onChange({ ...value, description }), [
        onChange,
        value,
    ])
    const onPlanChange = useCallback((plan: CampaignPlanSpecificationFormData): void => onChange({ ...value, plan }), [
        onChange,
        value,
    ])

    return (
        <div className={`card ${className}`}>
            <h3 className="card-header px-3">New campaign</h3>
            <div className="card-body p-3">
                <CampaignTitleField
                    className="e2e-campaign-title mb-2"
                    value={value.name}
                    onChange={onNameChange}
                    disabled={disabled}
                />
                <CampaignDescriptionField
                    value={value.description}
                    onChange={onDescriptionChange}
                    disabled={disabled}
                    className="mb-2"
                />
                <p className="ml-1 mb-0">
                    <small>
                        <a rel="noopener noreferrer" target="_blank" href="/help/user/markdown">
                            Markdown supported
                        </a>
                    </small>
                </p>
            </div>
            {!isExistingPlanID(value.plan) && (
                <>
                    <hr />
                    <div className="card-body p-3">
                        <CampaignPlanSpecificationFields
                            value={value.plan}
                            onChange={onPlanChange}
                            readOnly={false}
                            isLightTheme={isLightTheme}
                        />
                    </div>
                </>
            )}
            {children && (
                <>
                    <hr />
                    {children}
                </>
            )}
        </div>
    )
}
