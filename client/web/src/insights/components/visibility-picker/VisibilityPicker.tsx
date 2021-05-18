import React, { ChangeEvent } from 'react'

import { FormGroup } from '../form/form-group/FormGroup'
import { FormRadioInput } from '../form/form-radio-input/FormRadioInput'

export const getVisibilityValue = (event: VisibilityChangeEvent): string => {
    if (event.type === 'personal') {
        return 'personal'
    }

    if (event.type === 'organization') {
        return event.orgID
    }

    return ''
}

export type VisibilityChangeEvent = { type: 'personal' } | { type: 'organization'; orgID: string }

export interface Organization {
    id: string
    name: string
    displayName?: string | null
}

export interface VisibilityPickerProps {
    /**
     * Current visibility value. Possible value
     * 'personal' or any org id which is just a string.
     * */
    value: string

    /**
     * On change handler.
     * */
    onChange: (event: VisibilityChangeEvent) => void

    /**
     * Organization list - to display org radio buttons right after
     * personal radio
     * */
    organizations: Organization[]
}

/**
 * Shared component for visibility field for creation UI pages.
 * */
export const VisibilityPicker: React.FunctionComponent<VisibilityPickerProps> = props => {
    const { value, organizations, onChange } = props

    const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const value = event.target.value

        if (value === 'personal') {
            return onChange({ type: 'personal' })
        }

        const org = organizations.find(org => org.id === value)

        if (org) {
            onChange({ type: 'organization', orgID: org.id })
        }
    }

    return (
        <FormGroup
            name="visibility"
            title="Visibility"
            description="This insight will be visible only on your personal dashboard. It will not be shown to other
                            users in your organization."
            className="mb-0 mt-4"
            contentClassName="d-flex flex-wrap mb-n2"
        >
            <FormRadioInput
                name="visibility"
                value="personal"
                title="Personal"
                description="only you"
                checked={value === 'personal'}
                className="mr-3"
                onChange={handleChange}
            />

            {organizations.map(org => (
                <FormRadioInput
                    key={org.id}
                    name="visibility"
                    value={org.id}
                    title={org.displayName ?? org.name}
                    description={`all users in ${org.displayName ?? org.name} organization`}
                    checked={value === org.id}
                    onChange={handleChange}
                    className="mr-3"
                />
            ))}

            {organizations.length === 0 && (
                <FormRadioInput
                    name="visibility"
                    value="organization"
                    disabled={true}
                    title="Organization"
                    description="all users in your organization"
                    data-tooltip="Enable regular expression"
                    className="mr-3"
                    labelTooltipText="Create or join the Organization to share code insights with others!"
                />
            )}
        </FormGroup>
    )
}
