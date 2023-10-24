import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { Label } from './Label'

describe('Label', () => {
    it('should render correctly', () => {
        expect(
            render(
                <Label
                    size="small"
                    weight="medium"
                    isUppercase={true}
                    isUnderline={true}
                    alignment="center"
                    mode="default"
                >
                    Hello world
                </Label>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(render(<Label as="span">I am a label</Label>).asFragment()).toMatchSnapshot()
    })
})
