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
        const key = name + '.' + subNode.title
        switch (subNode.type) {
            case 'boolean':
                return (
                    <BooleanSettingItem key={key} name={name} title={subNode.title} description={subNode.description} />
                )
            case 'string':
                return <StringSettingItem key={key} name={name} title={subNode.title} />
            case 'object':
                if (subNode.properties) {
                    return (
                        <div key={key} className={groupLevelClasses[parentNames.length]}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            {convertPropertiesToComponents(subNode, [...parentNames, name])}
                        </div>
                    )
                }
                return <div key={key}>Unsupported object setting type</div>
            default:
                return <div key={key}>Unsupported setting type</div>
        }
    })
}

function BooleanSettingItem(props: { name: string; title: string; description: string }): JSX.Element {
    // TODO: Include the parent group(s) in the id to make it unique
    return (
        <>
            <Checkbox id={props.name} label={props.name} checked={false} onChange={() => {}} />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
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
