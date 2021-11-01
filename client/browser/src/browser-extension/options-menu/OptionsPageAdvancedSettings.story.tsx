import { DecoratorFn, Meta } from '@storybook/react'
import React, { useState } from 'react'

import brandedStyles from '../../branded.scss'

import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'

const OPTIONS_FLAGS = [
    { key: 'one', label: 'Flag One', value: true },
    { key: 'two', label: 'Flag Two', value: false },
    { key: 'three', label: 'Flag Three', value: false },
]

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

export const Default: React.FunctionComponent = () => {
    const [optionFlagValues, setOptionFlagValues] = useState(OPTIONS_FLAGS)
    const setOptionFlag = (key: string, value: boolean) => {
        setOptionFlagValues(optionFlagValues.map(flag => (flag.key === key ? { ...flag, value } : flag)))
    }

    return <OptionsPageAdvancedSettings optionFlags={optionFlagValues} onChangeOptionFlag={setOptionFlag} />
}
