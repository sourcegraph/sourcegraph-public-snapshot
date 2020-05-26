import * as React from 'react'
import { storiesOf } from '@storybook/react'
import { BadgeAttachment } from './BadgeAttachment'
import badgeStyles from './BadgeAttachment.scss'
import webStyles from '../../../web/src/SourcegraphWebApp.scss'
import { BadgeAttachmentRenderOptions } from 'sourcegraph'

import { radios } from '@storybook/addon-knobs'

const label = 'Theme'
const options = {
    Light: 'light',
    Dark: 'dark',
}
const defaultValue = 'light'
const groupId = 'GROUP-ID1'

const isLightTheme = () => radios(label, options, defaultValue, groupId) === 'light'

const { add } = storiesOf('BadgeAttachment', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <style>{badgeStyles}</style>
        <div style={{ color: 'var(--body-color)' }} className={isLightTheme() ? 'theme-light' : 'theme-dark'}>
            <div>{story()}</div>
        </div>
    </>
))

add('info', () => (
    <BadgeAttachment
        attachment={{ kind: 'info', hoverMessage: ' this is hover tooltip(info)' }}
        isLightTheme={isLightTheme()}
    />
))

add('warning', () => (
    <BadgeAttachment
        attachment={{ kind: 'warning', hoverMessage: 'this is hover tooltip(warning)' }}
        isLightTheme={isLightTheme()}
    />
))

add('error', () => (
    <BadgeAttachment
        attachment={{ kind: 'error', hoverMessage: 'this is hover tooltip(error)' }}
        isLightTheme={isLightTheme()}
    />
))

const oldFormatBadge: Omit<BadgeAttachmentRenderOptions, 'kind'> = {
    icon: makeInfoIcon('#ffffff'),
    light: { icon: makeInfoIcon('#000000') },
}

add('old format icon', () => (
    <BadgeAttachment attachment={oldFormatBadge as BadgeAttachmentRenderOptions} isLightTheme={isLightTheme()} />
))

function makeIcon(svg: string): string {
    return `data:image/svg+xml;base64,${btoa(
        svg
            .split('\n')
            .map(r => r.trimStart())
            .join(' ')
    )}`
}

function makeInfoIcon(color: string): string {
    return makeIcon(`
        <svg xmlns='http://www.w3.org/2000/svg' style="width:24px;height:24px" viewBox="0 0 24 24" fill="${color}">
            <path d="
                M11,
                9H13V7H11M12,
                20C7.59,
                20 4,
                16.41 4,
                12C4,
                7.59 7.59,
                4 12,
                4C16.41,
                4 20,
                7.59 20,
                12C20,
                16.41 16.41,
                20 12,
                20M12,
                2A10,
                10 0 0,
                0 2,
                12A10,
                10 0 0,
                0 12,
                22A10,
                10 0 0,
                0 22,
                12A10,
                10 0 0,
                0 12,
                2M11,
                17H13V11H11V17Z"
            />
        </svg>
    `)
}
