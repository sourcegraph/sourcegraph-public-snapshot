import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'

import { Timestamp, TimestampFormat } from './Timestamp'

describe('Timestamp', () => {
    test('mocked current time', () => expect(render(<Timestamp date="2006-01-02" />).asFragment()).toMatchSnapshot())

    test('with time time', () =>
        expect(render(<Timestamp date="2006-01-02T01:02:00Z" />).asFragment()).toMatchSnapshot())

    test('noAbout', () => expect(render(<Timestamp date="2006-01-02" noAbout={true} />).asFragment()).toMatchSnapshot())

    test('noAgo', () => expect(render(<Timestamp date="2006-01-02" noAgo={true} />).asFragment()).toMatchSnapshot())

    test('absolute time with formatting', () =>
        expect(
            render(
                <Timestamp
                    date="2006-01-02T01:02:00Z"
                    timestampFormat={TimestampFormat.FULL_TIME}
                    preferAbsolute={true}
                />
            ).asFragment()
        ).toMatchSnapshot())
})
