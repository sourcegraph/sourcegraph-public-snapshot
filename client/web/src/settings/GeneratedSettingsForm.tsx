import React from 'react'

import { Checkbox } from '@sourcegraph/wildcard'

import { SettingsSchema } from './SettingsFile'

interface GeneratedSettingsFormProps {
    jsonSchema: SettingsSchema
}

interface settingsNode {
    title: string
    description: string
    type: 'array' | 'boolean' | 'integer' | 'null' | 'number' | 'object' | 'string'
    enum?: string[]
    properties: Record<string, settingsNode>
}

export const GeneratedSettingsForm: React.FunctionComponent<GeneratedSettingsFormProps> = ({ jsonSchema }) => {
    return <>{convertPropertiesToComponents(jsonSchema as unknown as settingsNode)}</>
}

function convertPropertiesToComponents(jsonSchema: settingsNode): (JSX.Element | null)[] {
    return Object.entries(jsonSchema.properties).map(([name, node]) => {
        if (node.type === 'boolean') {
            return <BooleanSettingItem title={node.title} key={name + '.' + node.title} />
        }
        return null
    })
}

function BooleanSettingItem(props: { title: string }): JSX.Element {
    // TODO: Use a more unique ID
    return <Checkbox id={props.title} label={props.title} checked={false} />
}
