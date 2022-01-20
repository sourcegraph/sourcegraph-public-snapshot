import { Meta, Story } from '@storybook/react'
import { startCase } from 'lodash'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { ButtonLink } from '..'
import { BUTTON_VARIANTS } from '../../Button/constants'

import styles from './Story.module.scss'

const Config: Meta = {
    title: 'wildcard/ButtonLink',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: ButtonLink,
    },
}

export default Config

export const Overview: Story = () => (
    <>
        <h1>ButtonLink</h1>
        <h2>Variants</h2>
        <div className={styles.grid}>
            {BUTTON_VARIANTS.map(variant => (
                <ButtonLink
                    key={variant}
                    variant={variant}
                    to="https://sourcegraph.com"
                    target="_blank"
                    onClick={console.log}
                >
                    {startCase(variant)}
                </ButtonLink>
            ))}
        </div>
        <h2>Outline</h2>
        <ButtonLink
            variant="danger"
            outline={true}
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
        >
            Outline
        </ButtonLink>
        <h2>Icons</h2>
        <p>We can use icons with our buttons.</p>{' '}
        <ButtonLink
            variant="secondary"
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
        >
            <SearchIcon className="icon-inline mr-1" />
            Search
        </ButtonLink>
        <h2>Smaller</h2>
        <p>We can make our buttons smaller.</p>
        <ButtonLink
            variant="secondary"
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
            size="sm"
        >
            Smaller
        </ButtonLink>
    </>
)
