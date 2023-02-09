import React from 'react'

import { Checkbox, H3, Input, Label, Text } from '@sourcegraph/wildcard'

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

export const GeneratedSettingsForm: React.FunctionComponent<GeneratedSettingsFormProps> = ({ jsonSchema }) => (
    <>{convertPropertiesToComponents(jsonSchema as unknown as settingsNode)}</>
)

function convertPropertiesToComponents(node: settingsNode): JSX.Element[] {
    return Object.entries(node.properties).map(([name, subNode]) => {
        const key = name + '.' + subNode.title
        switch (subNode.type) {
            case 'boolean':
                return <BooleanSettingItem key={key} name={name} title={subNode.title} />
            case 'string':
                return <StringSettingItem key={key} name={name} title={subNode.title} />
            case 'object':
                if (subNode.properties) {
                    return (
                        <div key={key}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            {convertPropertiesToComponents(subNode)}
                        </div>
                    )
                }
                return <div key={key}>Unsupported object setting type</div>
            default:
                return <div key={key}>Unsupported setting type</div>
        }
    })
}

function BooleanSettingItem(props: { name: string; title: string }): JSX.Element {
    // TODO: Include the parent group(s) in the id to make it unique
    return <Checkbox id={props.name} label={props.title} checked={false} onChange={() => {}} />
}

function StringSettingItem(props: { name: string; title: string }): JSX.Element {
    // TODO: Include the parent group(s) in the id to make it unique
    return (
        <>
            <Label className="sr-only" htmlFor={props.name}>
                {props.title}
            </Label>
            <Input id={props.name} name={props.name} type="text" />
        </>
    )
}
