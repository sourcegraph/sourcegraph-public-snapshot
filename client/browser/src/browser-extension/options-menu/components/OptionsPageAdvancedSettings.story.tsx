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

export default {
    title: 'browser/Options/OptionsPageAdvancedSettings',

    decorators: [
        story => (
            <>
                <style>{brandedStyles}</style>
                <div className="options-page options-page--full">{story()}</div>
            </>
        ),
    ],
}

export const Default = () => <StoryWrapper />
export const WithEmptyEnabledBlocklist = () => <StoryWrapper blocklist={{ enabled: true, content: '' }} />

WithEmptyEnabledBlocklist.story = {
    name: 'With empty enabled "blocklist"',
}

export const WithNonEmptyEnabledBlocklist = () => (
    <StoryWrapper blocklist={{ enabled: true, content: 'https://github.com/my-repo/*' }} />
)

WithNonEmptyEnabledBlocklist.story = {
    name: 'With non-empty enabled "blocklist"',
}
