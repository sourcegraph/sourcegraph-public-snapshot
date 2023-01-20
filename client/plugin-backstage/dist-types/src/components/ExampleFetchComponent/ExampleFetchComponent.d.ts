type User = {
    gender: string;
    name: {
        title: string;
        first: string;
        last: string;
    };
    location: object;
    email: string;
    login: object;
    dob: object;
    registered: object;
    phone: string;
    cell: string;
    id: {
        name: string;
        value: string;
    };
    picture: {
        medium: string;
    };
    nat: string;
};
type DenseTableProps = {
    users: User[];
};
export declare const DenseTable: ({ users }: DenseTableProps) => JSX.Element;
export declare const ExampleFetchComponent: () => JSX.Element;
export {};
