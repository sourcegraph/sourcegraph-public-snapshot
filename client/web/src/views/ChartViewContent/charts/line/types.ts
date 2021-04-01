
export interface Accessors<Datum, Key extends string | number> {
    x: (d: Datum) => Date | number;
    y: Record<Key, (data: Datum) => any>;
}
