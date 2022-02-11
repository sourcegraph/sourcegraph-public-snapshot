export interface Point {
    x: number
    y: number
}

export const createPoint = (xCoord: number, yCoord: number): Point => ({ x: xCoord, y: yCoord })
