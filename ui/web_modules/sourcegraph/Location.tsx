export interface Location {
	pathname: string;
	search: string;
	hash: string;
	query: {[key: string]: string | undefined};
	state: any | null;
	action: string;
	key: string;
};
