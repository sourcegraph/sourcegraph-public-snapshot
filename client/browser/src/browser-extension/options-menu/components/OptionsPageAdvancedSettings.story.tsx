import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useState } from 'react'

import brandedStyles from '../../../branded.scss'
import { IOptionsPageContext, OptionsPageContext } from '../OptionsPage.context'

import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'

const StoryWrapper: React.FC<{ blocklist?: IOptionsPageContext['blocklist'] }> = ({ blocklist }) => {
    const [optionFlags, setOptionFlags] = useState([
        { key: 'one', label: 'Flag One', value: true },
        { key: 'two', label: 'Flag Two', value: false },
        { key: 'three', label: 'Flag Three', value: false },
    ])
    const setOptionFlag = (key: string, value: boolean) => {
        setOptionFlags(optionFlags.map(flag => (flag.key === key ? { ...flag, value } : flag)))
    }

    return (
        <OptionsPageContext.Provider
            value={{
                blocklist,
                optionFlags,
                onChangeOptionFlag: setOptionFlag,
                onBlocklistChange: () => {},
            }}
        >
            <OptionsPageAdvancedSettings />
        </OptionsPageContext.Provider>
    )
}

const decorator: DecoratorFn = story => (
    <>
        <style>{brandedStyles}</style>
        <div className="options-page options-page--full">{story()}</div>
    </>
)

const config: Meta = {
    title: 'browser/Options/OptionsPageAdvancedSettings',

    decorators: [decorator],
}
export default config

export const Default: Story = () => <StoryWrapper />
export const WithEmptyEnabledBlocklist: Story = () => <StoryWrapper blocklist={{ enabled: true, content: '' }} />

WithEmptyEnabledBlocklist.storyName = 'With empty enabled "blocklist"'

export const WithNonEmptyEnabledBlocklist: Story = () => (
    <StoryWrapper blocklist={{ enabled: true, content: 'https://github.com/my-repo/*' }} />
)

WithNonEmptyEnabledBlocklist.storyName = 'With non-empty enabled "blocklist"'
