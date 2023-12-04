import { describe, expect, it } from 'vitest'

import { getByteUnitLabel, getByteUnitValue } from '../site-admin/analytics/utils'

describe('getByteUnitLabel', () => {
    it('should determine the correct label from the number of bytes', () => {
        expect(getByteUnitLabel(886)).toEqual('Bytes')
        expect(getByteUnitLabel(1234)).toEqual('KB')
        expect(getByteUnitLabel(12343)).toEqual('KB')
        expect(getByteUnitLabel(123432)).toEqual('KB')
        expect(getByteUnitLabel(8234333)).toEqual('MB')
        expect(getByteUnitLabel(82343253)).toEqual('MB')
        expect(getByteUnitLabel(823433335)).toEqual('MB')
        expect(getByteUnitLabel(6234325223)).toEqual('GB')
        expect(getByteUnitLabel(62343252236)).toEqual('GB')
        expect(getByteUnitLabel(623432522396)).toEqual('GB')
    })
})

describe('getByteUnitValue', () => {
    it('should change unit of measurement based on amount', () => {
        expect(getByteUnitValue(886)).toEqual(886)
        expect(getByteUnitValue(12000)).toEqual(12)
        expect(getByteUnitValue(3400000)).toEqual(3.4)
        expect(getByteUnitValue(5600700200)).toEqual(5.6007002)
        expect(getByteUnitValue(1000)).toEqual(1)
        expect(getByteUnitValue(1000000)).toEqual(1)
        expect(getByteUnitValue(1000000000)).toEqual(1)
    })
})
