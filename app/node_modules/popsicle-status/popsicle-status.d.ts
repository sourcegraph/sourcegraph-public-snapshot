declare function popsicleStatus (): (req: any, next: () => any) => any;
declare function popsicleStatus (statusCode: number): (req: any, next: () => any) => any;
declare function popsicleStatus (lowerStatusCode: number, upperStatusCode: number): (req: any, next: () => any) => any;

export = popsicleStatus;
