
export function getRangeWithPadding([min, max]:number[], paddingCoefficient: number): [number, number] {
    const increment = (max - min) * paddingCoefficient / 2;

    return [min - increment, max + increment];
}
