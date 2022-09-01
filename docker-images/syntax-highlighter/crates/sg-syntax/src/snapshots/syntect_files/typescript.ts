export class MyClass {
  public static myValue: string;
  constructor(init: string) {
    this.myValue = init;
  }
}
export abstract class MyAbstractClass {}
import fs = require("fs");
declare module MyModule {
  export interface MyInterface extends Other {
    myProperty: any;
    myKeyoff: keyof MyClass;
  }
}
declare magicNumber number;
myArray.forEach(() => { }); // fat arrow syntax
const oneOf = (a:number): number => a + 1
export enum Day {
    Weekday = 1,
    Weekend = 2
}
export type MyNumber = number
export const a = 42
const aa = 42
export let b = 42
export var c = 42
export var d1 = {e2:41}
export var d = {
    key1: 1,
    key2: null,
    key3: `abc${d1.e2}`,
}
export var e = Object.keys(d) as MyClass
export function f() {}

const negatedFilterToNegatableFilter: { [key: string]: MyClass } = null as any

const scanToken = <T extends Term = Literal>(
    regexp: RegExp,
    output?: T | ((input: string, range: CharacterRange) => T),
    expected?: string
): Parser<T> => {
    if (!regexp.source.startsWith('^')) {
        regexp = new RegExp(`^${regexp.source}`, regexp.flags)
    }
}

export const URI: typeof URL

export class SiteAdminUsageStatisticsPage extends React.Component<
    SiteAdminUsageStatisticsPageProps,
    SiteAdminUsageStatisticsPageState
> {
    private loadLatestChartFromStorage(): keyof ChartOptions {
        const latest = localStorage.getItem(CHART_ID_KEY)
        return latest && latest in chartGeneratorOptions ? (latest as keyof ChartOptions) : 'daus'
    }

}
