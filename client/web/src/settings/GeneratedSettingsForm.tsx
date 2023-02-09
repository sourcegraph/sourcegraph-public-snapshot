import React from 'react'

import { Checkbox, H3, Input, Label, Text } from '@sourcegraph/wildcard'

import { SettingsSchema } from './SettingsFile'

import styles from './GeneratedSettingsForm.module.scss'

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

const groupLevelClasses = [styles.groupLevel0, styles.groupLevel1, styles.groupLevel2, styles.groupLevel3]

export const GeneratedSettingsForm: React.FunctionComponent<GeneratedSettingsFormProps> = ({ jsonSchema }) => (
    <>{convertPropertiesToComponents(jsonSchema as unknown as settingsNode, [])}</>
)

function convertPropertiesToComponents(node: settingsNode, parentNames: string[]): JSX.Element[] {
    return Object.entries(node.properties).map(([name, subNode]) => {
        const id = parentNames.concat(name).join('.')
        switch (subNode.type) {
            case 'boolean':
                return (
                    <BooleanSettingItem
                        key={id}
                        id={id}
                        name={name}
                        title={subNode.title}
                        description={subNode.description}
                    />
                )
            case 'string':
                return (
                    <StringSettingItem
                        key={id}
                        id={id}
                        name={name}
                        title={subNode.title}
                        description={subNode.description}
                    />
                )
            case 'object':
                if (subNode.properties) {
                    return (
                        <div key={id} className={groupLevelClasses[parentNames.length]}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            {convertPropertiesToComponents(subNode, [...parentNames, name])}
                        </div>
                    )
                }
                return <div key={id}>Unsupported object setting type</div>
            default:
                return <div key={id}>Unsupported setting type</div>
        }
    })
}

function BooleanSettingItem(props: { id: string; name: string; title: string; description: string }): JSX.Element {
    return (
        <>
            <Checkbox id={props.id} label={props.name} checked={false} onChange={() => {}} />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function StringSettingItem(props: { id: string; name: string; title: string; description: string }): JSX.Element {
    return (
        <>
            <Label htmlFor={props.id} className="sr-only">
                {props.name}
            </Label>
            <Input id={props.id} type="text" />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}
