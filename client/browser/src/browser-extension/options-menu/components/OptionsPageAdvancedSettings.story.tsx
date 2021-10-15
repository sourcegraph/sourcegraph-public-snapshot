import { storiesOf } from '@storybook/react'
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

storiesOf('browser/Options/OptionsPageAdvancedSettings', module)
    .addDecorator(story => (
        <>
            <style>{brandedStyles}</style>
            <div className="options-page options-page--full">{story()}</div>
        </>
    ))
    .add('Default', () => <StoryWrapper />)
    .add('With empty enabled "blocklist"', () => <StoryWrapper blocklist={{ enabled: true, content: '' }} />)
    .add('With non-empty enabled "blocklist"', () => (
        <StoryWrapper blocklist={{ enabled: true, content: 'https://github.com/my-repo/*' }} />
    ))
