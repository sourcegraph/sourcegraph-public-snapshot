import { render } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { Timestamp, TimestampFormat } from './Timestamp'

describe('Timestamp', () => {
    const getDate = () => new Date('2023-12-11')

    test('mocked current time', () =>
        expect(render(<Timestamp date="2006-01-02" now={getDate} />).asFragment()).toMatchSnapshot())

    test('with time time', () =>
        expect(render(<Timestamp date="2006-01-02T01:02:00Z" now={getDate} />).asFragment()).toMatchSnapshot())

    test('noAbout', () =>
        expect(render(<Timestamp date="2006-01-02" now={getDate} noAbout={true} />).asFragment()).toMatchSnapshot())

    test('noAgo', () =>
        expect(render(<Timestamp date="2006-01-02" now={getDate} noAgo={true} />).asFragment()).toMatchSnapshot())

    test('absolute time with formatting', () =>
        expect(
            render(
                <Timestamp
                    date="2006-01-02T01:02:00Z"
                    now={getDate}
                    timestampFormat={TimestampFormat.FULL_TIME}
                    preferAbsolute={true}
                />
            ).asFragment()
        ).toMatchSnapshot())
})
