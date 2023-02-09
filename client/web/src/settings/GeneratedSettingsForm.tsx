import React from 'react'

import { Checkbox, H3, Input, Label, Select, Text } from '@sourcegraph/wildcard'

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
    minimum?: number
    maximum?: number
}

const groupLevelClasses = [styles.groupLevel1, styles.groupLevel2, styles.groupLevel3]

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
            case 'integer':
            case 'number':
                return (
                    <Input
                        label={<Text>{name}</Text>}
                        type="number"
                        min={subNode?.minimum ?? 0}
                        max={subNode?.maximum ?? ''}
                    />
                )
            case 'string':
                return subNode.enum?.length ? (
                    <>
                        <Select id={id} name={name} label={name} value={''} className="mb-0">
                            {subNode.enum.map(enumValue => (
                                <option key={enumValue} value={enumValue} label={enumValue} />
                            ))}
                        </Select>
                        <Text className="text-muted">{subNode.description}</Text>
                    </>
                ) : (
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
                        <div key={id}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            <div className={groupLevelClasses[parentNames.length]}>
                                {convertPropertiesToComponents(subNode, [...parentNames, name])}
                            </div>
                        </div>
                    )
                }
                return <div key={id}>Unsupported object setting type</div>
            // TODO: Handle integers, numbers, and "null" type
            // TODO: Handle array types
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
            <Label htmlFor={props.id}>{props.name}</Label>
            <Input id={props.id} type="text" />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}
