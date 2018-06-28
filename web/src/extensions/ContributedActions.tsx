import * as React from 'react'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import { ContributedActionItem, ContributedActionItemProps } from './ContributedActionItem'
import { ContributableMenu } from './contributions'

interface Props extends ExtensionsProps, ExtensionsChangeProps {
    menu: ContributableMenu
}

export const ContributedActions: React.SFC<Props> = props => {
    const items: ContributedActionItemProps[] = []
    for (const x of props.extensions) {
        if (x.contributions && x.contributions.commands && x.contributions.menus && x.contributions.menus[props.menu]) {
            for (const { command: commandID } of x.contributions.menus[props.menu]) {
                const command = x.contributions.commands.find(c => c.command === commandID)
                if (command) {
                    items.push({
                        extensionID: x.extensionID,
                        contribution: command,
                    })
                }
            }
        }
    }

    return (
        <>
            {items.map((item, i) => (
                <ContributedActionItem
                    key={i}
                    {...item}
                    extensions={props.extensions}
                    onExtensionsChange={props.onExtensionsChange}
                />
            ))}
        </>
    )
}
