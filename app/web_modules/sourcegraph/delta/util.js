// isDevNull returns if path is "/dev/null". That is the sentinel value in git diffs
// that indicates the file does not exist (i.e., it is a newly added file, or a removed file,
// if the old or new path is "/dev/null", respectively).
export function isDevNull(path) {
	return path === "/dev/null";
}
