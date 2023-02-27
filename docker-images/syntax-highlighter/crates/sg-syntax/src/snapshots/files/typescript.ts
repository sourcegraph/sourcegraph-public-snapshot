import fs from 'fs'
export class MyClass {
  public static myValue: string;
  constructor(init: string) {
    super();
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
    key4: true,
    key5: 1.5,
    key6: 'a',
    key7: [1].map(n => ({n, a: n + 1}))
}
export var e = Object.keys(d) as MyClass
export const e2: never[] = []
export const e3: undefined = undefined
export const e4: null = null
export function e5(): void = {}
export const e6 = Math.max(Math.min, Math.PI)
const { a, b: c } = { a, b: 42 }
export function f() {}

const negatedFilterToNegatableFilter: { [key: string]: MyClass } = null as any

const scanToken = <T extends Term = Literal>(
    regexp: RegExp,
    output?: T | ((input: string, range: CharacterRange) => T),
    expected?: string
): Parser<T> => {
    const { a, b: c } = { a, b: 42 }
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

export function newFuzzyFSM(filenames: string[], createUrl: createUrlFunction): FuzzyFSM {
    return newFuzzyFSMFromValues(
        filenames.map(text => ({
            text,
            icon: fileIcon(text),
        })),
        createUrl
    )
}

// 1. Advanced types
type Age = number;
type Person = {
    name: string;
    age: Age;
};

const john: Person = {
    name: "John",
    age: 30
};

// 2. Intersection Types
type Admin = {
    role: "admin";
};
type Manager = {
    role: "manager";
};
type User = Person & (Admin | Manager);

const admin: User = {
    name: "admin",
    age: 35,
    role: "admin"
};

// 3. Union Types
type StringOrNumber = string | number;
const unionExample: StringOrNumber = "hello";

// 4. Type Aliases
type AgeRange = 18 | 25 | 30 | 40;
const ageRange: AgeRange = 25;

// 5. Type Guards
function isString(value: StringOrNumber): value is string {
    return typeof value === "string";
}
if (isString(unionExample)) {
    console.log(`Value is a string: ${unionExample}`);
}

// 6. Type inference
const value = "Hello";
const valueLength = value.length;

// 7. Type parameter constraints
class Collection<T extends object> {
    items: T[];
    constructor(items: T[]) {
        this.items = items;
    }
    getFirst(): T {
        return this.items[0];
    }
}

const people = new Collection([{ name: "John" }, { name: "Jane" }]);
const firstPerson = people.getFirst();

// 8. Higher Order Types
type Filter = {
    (array: number[], callback: (item: number) => boolean): number[];
};
const filter: Filter = (array, callback) => {
    const result = [];
    for (const item of array) {
        if (callback(item)) {
            result.push(item);
        }
    }
    return result;
};

const filtered = filter([1, 2, 3, 4], item => item % 2 === 0);
console.log(filtered);

// 9. Index Types
type People = {
    [key: string]: Person;
};
const peopleObject: People = {
    john: { name: "John", age: 30 },
    jane: { name: "Jane", age: 25 }
};

// 10. Readonly properties
interface Car {
    readonly make: string;
    readonly model: string;
    readonly year: number;
}
const car: Car = {
    make: "Tesla",
    model: "Model S",
    year: 2020
};

// 11. Keyof operator
type CarProperties = keyof Car;
const property: CarProperties = "make";

// 12. Mapped Types
type ReadonlyCar = Readonly<Car>;
const readonlyCar: ReadonlyCar = {
    make: "Tesla",
    model: "Model S",
    year: 2020
};

// 13. Conditional Types
type IsNumber<T> = T extends number ? true : false;
type IsNumberType = IsNumber<number>;
const isNumberType: IsNumberType = true;



// 14. Exclude from type
type ExcludePersonAge = Exclude<keyof Person, "age">;
const excludedPersonProperties: ExcludePersonAge = "name";

// 15. Extract from type
type ExtractPersonAge = Extract<keyof Person, "age">;
const extractedPersonProperties2: ExtractPersonAge = "age";

// 16. Non-nullable type
type NonNullableCar = NonNullable<Car>;
const nonNullableCar: NonNullableCar = {
    make: "Tesla",
    model: "Model S",
    year: 2020
};

// 17. Required properties
interface CarDetails {
    make?: string;
    model?: string;
    year?: number;
}
type RequiredCarDetails = Required<CarDetails>;
const requiredCarDetails: RequiredCarDetails = {
    make: "Tesla",
    model: "Model S",
    year: 2020
};


// 18. Tuple types
type PersonTuple = [string, Age];
const personTuple: PersonTuple = ["John", 30];

// 19. Literal types
type Days = "Monday" | "Tuesday" | "Wednesday" | "Thursday" | "Friday";
const days: Days = "Monday";

// 20. Enum
enum Color {
    Red,
    Green,
    Blue
}
const color = Color.Blue;

// 21. Numeric enums
enum Result {
    Success = 100,
    Failure = 200
}
const result = Result.Success;

// 22. String enums
enum Direction {
    Up = "UP",
    Down = "DOWN",
    Left = "LEFT",
    Right = "RIGHT"
}
const direction = Direction.Right;

// 23. Generics
function identity<T>(value: T): T {
    return value;
}
const identityExample = identity<string>("hello");

// 24. Polymorphic this types
class MyArray<T> {
    add(value: T) {
        this[this.length] = value;
        return this;
    }
}
const myArray = new MyArray<string>();
myArray.add("hello").add("world");

// 25. Partial types
type PartialPerson = Partial<Person>;
const partialPerson: PartialPerson = { name: "John" };

// 26. Pick types
type PickPersonAge = Pick<Person, 'age'>;
const pickedPersonProperties: PickPersonAge = { age: 30 };

// 27. Record types
type RecordPerson = Record<"key", Person>;
const recordPerson: RecordPerson = { key: { name: "John", age: 30 } };

// 28. Interface inheritance
interface Shape {
    width: number;
    height: number;
}
interface Square extends Shape {
    sideLength: number;
}
const square: Square = {
    width: 10,
    height: 10,
    sideLength: 10
};
