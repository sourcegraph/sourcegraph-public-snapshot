// tslint:disable: typedef ordered-imports curly

// httpStatusCode returns the HTTP status code that is most appropriate
// for the given Error (or 200 for null errors);
export function httpStatusCode(err: any): number {
	if (!err) return 200;
	if (err.response) return err.response.status;
	return 500;
}
