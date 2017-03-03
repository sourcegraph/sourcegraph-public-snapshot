import * as React from "react";
import { Symbol } from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: React.CSSProperties;
	color?: any;
}

export function Activity(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M13 12h38a5 5 0 0 1 5 5v30a5 5 0 0 1-5 5H13a5 5 0 0 1-5-5V17a5 5 0 0 1 5-5zm3.164 36.808a2.002 2.002 0 0 0 2.655-.978l4.798-10.392 3.973 3.97a2 2 0 0 0 3.23-.576l4.596-9.953 3.989 5.316a2.001 2.001 0 0 0 3.498-.567l6-17.995a2 2 0 0 0-3.795-1.264l-4.793 14.374-3.71-4.944a2 2 0 0 0-3.416.362L28.39 36.553l-3.973-3.97a2 2 0 0 0-3.23.576l-6 12.996a2 2 0 0 0 .978 2.653z" /></Symbol>;
};

export function Add(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 34H34v16a2 2 0 1 1-4 0V34H14a2 2 0 0 1 0-4h16V14a2 2 0 1 1 4 0v16h16a2 2 0 0 1 0 4z" /></Symbol>;
};

export function Alarm(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M22.808 9.304a.81.81 0 0 1-.089 1.147L10.31 20.836a.822.822 0 0 1-1.153-.108 8.92 8.92 0 0 1 .98-12.615 9.048 9.048 0 0 1 12.672 1.191zM28 12.384V12a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v.384a20.977 20.977 0 0 1 12.484 33.599l3.47 6.933a.75.75 0 0 1-.67 1.084h-3.99a2 2 0 0 1-1.414-.585l-2.678-2.674a20.877 20.877 0 0 1-22.404 0l-2.678 2.674a2 2 0 0 1-1.414.585h-3.99a.75.75 0 0 1-.67-1.084l3.47-6.933A20.978 20.978 0 0 1 28 12.383zm-3.933 27.504l9.56-4.472a2.386 2.386 0 0 0 1.382-2.411L33.496 18.31a1.509 1.509 0 0 0-3 0l-1.354 13.168-6.638 5.996a1.455 1.455 0 0 0 1.563 2.414zM41.191 9.275a9.045 9.045 0 0 1 12.673-1.192 8.93 8.93 0 0 1 .98 12.623.822.822 0 0 1-1.153.108L41.28 10.423a.811.811 0 0 1-.089-1.148z" /></Symbol>;
};

export function Alert(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 54a25.564 25.564 0 0 1-4.69-.498 2.586 2.586 0 0 0-2.055.532 21.088 21.088 0 0 1-10.057 3.953 1.09 1.09 0 0 1-1.016-1.698 15.53 15.53 0 0 0 2.96-5.967 1.613 1.613 0 0 0-.603-1.57A21.16 21.16 0 0 1 8 32c0-12.135 10.745-22 24-22s24 9.865 24 22-10.745 22-24 22zm-.007-36a2.936 2.936 0 0 0-3 2.997l.555 13.755a2.454 2.454 0 0 0 4.891 0l.555-13.755a2.936 2.936 0 0 0-3-2.997zm0 22a3 3 0 1 0 3 3 3 3 0 0 0-2.999-3z" /></Symbol>;
};

export function Archive(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 20H12a2 2 0 0 1-2-2v-4a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2zm-2 2a1 1 0 0 1 1 1v26a3 3 0 0 1-3 3H16a3 3 0 0 1-3-3V23a1 1 0 0 1 1-1zM26 32h12a2 2 0 0 0 0-4H26a2 2 0 0 0 0 4z" /></Symbol>;
};

export function ArrowDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.398 53.409a1.999 1.999 0 0 1-2.83 0L15.565 38.397a2 2 0 1 1 2.829-2.83L30 47.179V12a2 2 0 1 1 4 0v35.145l11.573-11.578a2.001 2.001 0 1 1 2.83 2.83z" /></Symbol>;
};

export function ArrowLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 34H16.828l11.587 11.573a2.002 2.002 0 0 1-2.832 2.829L10.56 33.397a1.998 1.998 0 0 1 0-2.828l15.023-15.005a2.002 2.002 0 1 1 2.832 2.83L16.795 30H52a2 2 0 1 1 0 4z" /></Symbol>;
};

export function ArrowNE(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46 41a2 2 0 0 1-2-2V22.845l-24.6 24.6a1.99 1.99 0 1 1-2.802-2.825L41.218 20H25a2 2 0 1 1 0-4h21a2 2 0 0 1 2 2v21a2 2 0 0 1-2 2z" /></Symbol>;
};

export function ArrowRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M38.417 48.436a2.002 2.002 0 1 1-2.832-2.829L47.205 34H12a2 2 0 0 1 0-4h35.172L35.585 18.427a2.002 2.002 0 1 1 2.832-2.829L53.44 30.602a1.998 1.998 0 0 1 0 2.83z" /></Symbol>;
};

export function ArrowNW(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47.408 47.439a1.98 1.98 0 0 1-2.808.005L20 22.844V39a2 2 0 0 1-4 0V18a2 2 0 0 1 2-2h21a2 2 0 1 1 0 4H22.782l24.62 24.62a1.999 1.999 0 0 1 .006 2.819z" /></Symbol>;
};

export function ArrowSE(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46 48H25a2 2 0 1 1 0-4h16.218l-24.62-24.62a1.99 1.99 0 0 1 2.802-2.825l24.6 24.6V25a2 2 0 0 1 4 0v21a2 2 0 0 1-2 2z" /></Symbol>;
};

export function ArrowSW(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47.402 19.38L22.782 44H39a2 2 0 0 1 0 4H18a2 2 0 0 1-2-2V25a2 2 0 1 1 4 0v16.155l24.6-24.6a1.99 1.99 0 1 1 2.802 2.825z" /></Symbol>;
};

export function ArrowUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45.607 28.433L34 16.821V52a2 2 0 1 1-4 0V16.855L18.427 28.433a2 2 0 0 1-2.829-2.83l15.004-15.012a2 2 0 0 1 2.83 0l15.004 15.012a2 2 0 0 1-2.829 2.83z" /></Symbol>;
};

export function Backward(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 39H32.011v8.011a2.003 2.003 0 0 1-3.285 1.538L10.699 33.537a2 2 0 0 1 0-3.076L28.726 15.45a2.003 2.003 0 0 1 3.285 1.538V25H52a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2z" /></Symbol>;
};

export function Book(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M29.14 50.833a25.048 25.048 0 0 0-11.79-2.845 27.517 27.517 0 0 0-6.885.942A1.12 1.12 0 0 1 9 47.859V15.243a1.125 1.125 0 0 1 .732-1.055 26.963 26.963 0 0 1 7.618-1.198 24.417 24.417 0 0 1 13.098 3.626 1.24 1.24 0 0 1 .552 1.04v32.107a1.243 1.243 0 0 1-1.86 1.07zM55 15.243v32.616a1.12 1.12 0 0 1-1.466 1.07 27.517 27.517 0 0 0-6.884-.941 25.049 25.049 0 0 0-11.79 2.845 1.243 1.243 0 0 1-1.86-1.07V17.656a1.24 1.24 0 0 1 .552-1.04A24.417 24.417 0 0 1 46.65 12.99a26.966 26.966 0 0 1 7.618 1.198A1.125 1.125 0 0 1 55 15.243z" /></Symbol>;
};

export function Bookmark(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M17 53.989v-42.98a3 3 0 0 1 3-3h24a3 3 0 0 1 3 3v42.98a2 2 0 0 1-2.997 1.734l-11.504-6.616a1 1 0 0 0-.997 0l-11.505 6.616A2 2 0 0 1 17 53.989z" /></Symbol>;
};

export function BookClosed(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 56H20.485a6 6 0 0 1-4.242-1.757l-3.071-3.071A4 4 0 0 1 12 48.343V12a2 2 0 0 1 2-2h30a2 2 0 0 1 2 2v36a2 2 0 0 1-2 2H18.832a.5.5 0 0 0-.353.854l.56.56a2 2 0 0 0 1.414.586H47.5a.5.5 0 0 0 .5-.5v-35a.5.5 0 0 1 .5-.5H50a2 2 0 0 1 2 2v36a2 2 0 0 1-2 2z" /></Symbol>;
};

export function BookmarkEmpty(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M20.488 55.604A2.998 2.998 0 0 1 16 53.001V12a4 4 0 0 1 4-4h24a4 4 0 0 1 4 4v41.001a2.998 2.998 0 0 1-4.488 2.603L32 49.03zM44 51.28V12H20v39.28l10.76-6.146a2.5 2.5 0 0 1 2.48 0z" /></Symbol>;
};

export function Branch(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M37.005 22.01a1.027 1.027 0 0 1-.82-1.622l6.99-9.974a.996.996 0 0 1 1.638 0l6.988 9.974a1.027 1.027 0 0 1-.819 1.623H46v5.676c0 2.67-3.04 7.182-4.93 9.07l-5.313 5.314A5.963 5.963 0 0 0 34 46.314V52a2 2 0 0 1-4 0v-5.686a5.963 5.963 0 0 0-1.757-4.243l-5.314-5.314c-1.888-1.888-4.929-6.4-4.929-9.07V22.01h-4.995a1.027 1.027 0 0 1-.82-1.623l6.99-9.974a.996.996 0 0 1 1.638 0l6.988 9.974a1.027 1.027 0 0 1-.819 1.623H22v5.676a5.958 5.958 0 0 0 1.757 4.242l7.314 7.314A10.133 10.133 0 0 1 32 40.33a10.11 10.11 0 0 1 .929-1.088l7.314-7.314A5.959 5.959 0 0 0 42 27.687V22.01z" /></Symbol>;
};

export function BrowserRefresh(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M29.99 14.102V9.005a1.024 1.024 0 0 1 1.617-.82l9.949 6.99a.997.997 0 0 1 0 1.638l-9.949 6.989a1.024 1.024 0 0 1-1.618-.82v-4.841A15.994 15.994 0 1 0 48 34a2 2 0 1 1 4 0 20 20 0 1 1-22.01-19.898z" /></Symbol>;
};

export function BrowserStop(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 38.018V25.51a8.726 8.726 0 0 1 2.696-6.307l9.184-8.783A8.727 8.727 0 0 1 25.912 8h12.735a8.727 8.727 0 0 1 6.183 2.568l8.625 8.659A8.727 8.727 0 0 1 56 25.386V38.14a8.727 8.727 0 0 1-2.42 6.031l-8.73 9.132A8.727 8.727 0 0 1 38.542 56H26.016a8.728 8.728 0 0 1-6.16-2.545L10.568 44.2A8.727 8.727 0 0 1 8 38.018zm13.585 1.555a2.002 2.002 0 1 0 2.83 2.83l7.574-7.572 7.584 7.584a2.002 2.002 0 0 0 2.83-2.83L34.82 32l7.613-7.613a2.002 2.002 0 0 0-2.83-2.83L31.99 29.17l-7.602-7.602a2.002 2.002 0 0 0-2.83 2.83L29.158 32z" /></Symbol>;
};

export function Business(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 24v2a2 2 0 0 1 2 2v24a2 2 0 0 1-2 2h-8a2 2 0 0 1-2-2V28a2 2 0 0 1 2-2v-2a2 2 0 0 1 2-2l1.426-9.506a.58.58 0 0 1 1.148 0L52 22a2 2 0 0 1 2 2zm-6 25a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm0-8a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm3-11h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1zM8 52V20a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v32a2 2 0 0 1-2 2h-8a2 2 0 0 1-2-2zm4-3a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm0-8a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm0-8a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm0-8a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1zm28 29H24a2 2 0 0 1-2-2V12a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v40a2 2 0 0 1-2 2zM30 15a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm8-32a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Calendar(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M24 10v6a2 2 0 1 1-4 0v-6a2 2 0 1 1 4 0zm30 6v34a4 4 0 0 1-4 4H14a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h3a1 1 0 0 1 1 1v2.828a4.116 4.116 0 0 0 3.607 4.153A4 4 0 0 0 26 16v-3a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v2.828a4.116 4.116 0 0 0 3.607 4.153A4 4 0 0 0 46 16v-3a1 1 0 0 1 1-1h3a4 4 0 0 1 4 4zm-5 12H15a1 1 0 0 0-1 1v20a1 1 0 0 0 1 1h34a1 1 0 0 0 1-1V29a1 1 0 0 0-1-1zM42 8a2 2 0 0 1 2 2v6a2 2 0 1 1-4 0v-6a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Camera(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M10 19a2 2 0 0 1 0-4h8a2 2 0 0 1 0 4zm14-4h28a4 4 0 0 1 4 4v26a4 4 0 0 1-4 4H12a4 4 0 0 1-4-4V23a2 2 0 0 1 2-2h10a2 2 0 0 0 2-2v-2a2 2 0 0 1 2-2zm15 28a11 11 0 1 0-11-11 11 11 0 0 0 11 11zm0-20a9 9 0 1 1-9 9 9 9 0 0 1 9-9z" /></Symbol>;
};

export function CameraRoll(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 54H20a4 4 0 0 1-4-4v-6h-6a4 4 0 0 1-4-4V14a4 4 0 0 1 4-4h34a4 4 0 0 1 4 4v6h6a4 4 0 0 1 4 4v26a4 4 0 0 1-4 4zm0-29a1 1 0 0 0-1-1H21a1 1 0 0 0-1 1v24a1 1 0 0 0 1 1h32a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Categories(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 25v26a3 3 0 0 1-3 3H15a3 3 0 0 1-3-3V25a3 3 0 0 1 3-3h34a3 3 0 0 1 3 3zM19 14a1 1 0 0 1-1-1 3 3 0 0 1 3-3h22a3 3 0 0 1 3 3 1 1 0 0 1-1 1zm30 5a1 1 0 0 1-1 1H16a1 1 0 0 1-1-1 3 3 0 0 1 3-3h28a3 3 0 0 1 3 3z" /></Symbol>;
};

export function Chart(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 52H10a2 2 0 0 1 0-4h44a2 2 0 0 1 0 4zm-2-6h-8a2 2 0 0 1-2-2V22a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v22a2 2 0 0 1-2 2zM36 12a2 2 0 0 1 2 2v30a2 2 0 0 1-2 2h-8a2 2 0 0 1-2-2V14a2 2 0 0 1 2-2zM20 28a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-8a2 2 0 0 1-2-2V30a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Chat(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 54a25.557 25.557 0 0 1-4.69-.498 2.586 2.586 0 0 0-2.054.532 21.089 21.089 0 0 1-10.059 3.953 1.09 1.09 0 0 1-1.015-1.698 15.533 15.533 0 0 0 2.96-5.967 1.613 1.613 0 0 0-.603-1.57A21.159 21.159 0 0 1 8 32c0-12.135 10.745-22 24-22s24 9.865 24 22-10.745 22-24 22z" /></Symbol>;
};

export function Checkmark(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M17.646 30.834l9.79 14.022 20.95-30.01a2.002 2.002 0 0 1 3.05-.278 2.088 2.088 0 0 1 .172 2.671L29.47 48.95a2.482 2.482 0 0 1-4.07 0L14.426 33.228a2.088 2.088 0 0 1 .171-2.672 2.002 2.002 0 0 1 3.05.278z" /></Symbol>;
};

export function ChevronDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.398 43.409a2 2 0 0 1-2.83 0L15.565 28.397a2 2 0 1 1 2.829-2.83l13.59 13.597 13.59-13.597a2.001 2.001 0 1 1 2.829 2.83z" /></Symbol>;
};

export function ChevronLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M38.415 48.402a2.004 2.004 0 0 1-2.832 0L20.56 33.397a1.998 1.998 0 0 1 0-2.829l15.023-15.004a2.002 2.002 0 1 1 2.832 2.829L24.81 31.983l13.606 13.59a1.998 1.998 0 0 1 0 2.829z" /></Symbol>;
};

export function ChevronNE(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M19 22h21a2 2 0 0 1 2 2v21a2 2 0 0 1-4 0V26H19a2 2 0 1 1 0-4z" /></Symbol>;
};

export function ChevronRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M28.417 48.436a2.002 2.002 0 1 1-2.832-2.829l13.606-13.59-13.606-13.59a2.002 2.002 0 0 1 2.832-2.829L43.44 30.602a1.998 1.998 0 0 1 0 2.83z" /></Symbol>;
};

export function ChevronSE(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M19 42h21a2 2 0 0 0 2-2V19a2 2 0 1 0-4 0v19H19a2 2 0 1 0 0 4z" /></Symbol>;
};

export function ChevronNW(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45 22H24a2 2 0 0 0-2 2v21a2 2 0 0 0 4 0V26h19a2 2 0 1 0 0-4z" /></Symbol>;
};

export function ChevronSW(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45 42H24a2 2 0 0 1-2-2V19a2 2 0 1 1 4 0v19h19a2 2 0 1 1 0 4z" /></Symbol>;
};

export function Circle(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20z" /></Symbol>;
};

export function ChevronUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45.607 38.433l-13.59-13.596-13.59 13.596a2 2 0 0 1-2.829-2.83l15.004-15.012a2 2 0 0 1 2.83 0l15.004 15.012a2 2 0 0 1-2.829 2.83z" /></Symbol>;
};

export function Child(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M28.547 24.024a3.98 3.98 0 0 1-2.667-1.026l-8.334-7.609a1.492 1.492 0 0 0-2.057.054l-1.022 1.025a1.502 1.502 0 0 0-.091 2.02l9.652 11.532.745 4.293a16.016 16.016 0 0 1 .197 3.426l-.957 16.688A1.497 1.497 0 0 0 25.506 56h1.981a1.745 1.745 0 0 0 1.71-1.406l2.795-13.546 2.796 13.545A1.747 1.747 0 0 0 36.5 56h1.984a1.498 1.498 0 0 0 1.494-1.575l-.958-16.7a16.027 16.027 0 0 1 .197-3.43L39.963 30l9.661-11.541a1.503 1.503 0 0 0-.091-2.021l-1.023-1.025a1.494 1.494 0 0 0-2.059-.055l-8.343 7.615A3.984 3.984 0 0 1 35.439 24zM31.983 22A7 7 0 1 0 25 15a6.991 6.991 0 0 0 6.983 7z" /></Symbol>;
};

export function CircleAdd(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20zm2 33a2 2 0 1 1-4 0V34H19a2 2 0 0 1 0-4h11V19a2 2 0 1 1 4 0v11h11a2 2 0 0 1 0 4H34z" /></Symbol>;
};

export function CircleArrowLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 19.899 22h-25.07l7.586 7.573a2.002 2.002 0 0 1-2.832 2.829L20.56 33.397a1.998 1.998 0 0 1 0-2.829l11.023-11.004a2.002 2.002 0 0 1 2.832 2.83L26.795 30h25.104A20 20 0 0 0 32 12z" /></Symbol>;
};

export function CircleAddAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm2-11a2 2 0 1 1-4 0V34H19a2 2 0 0 1 0-4h11V19a2 2 0 1 1 4 0v11h11a2 2 0 0 1 0 4H34z" /></Symbol>;
};

export function CircleArrowDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 32a24 24 0 1 1 24 24A24 24 0 0 1 8 32zm44 0a20.001 20.001 0 0 0-18-19.9v25.072l7.573-7.587a2.002 2.002 0 0 1 2.829 2.832L33.397 43.44a1.998 1.998 0 0 1-2.828 0L19.564 32.417a2.002 2.002 0 0 1 2.83-2.832L30 37.205V12.102A20 20 0 1 0 52 32z" /></Symbol>;
};

export function CircleArrowRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20.001 20.001 0 0 0-19.9 18h25.072l-7.587-7.573a2.002 2.002 0 1 1 2.832-2.829L43.44 30.602a1.998 1.998 0 0 1 0 2.83L32.417 44.435a2.002 2.002 0 0 1-2.832-2.83L37.205 34H12.102A20 20 0 1 0 32 12z" /></Symbol>;
};

export function CircleArrowUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 32a24 24 0 1 1 24 24A24 24 0 0 1 8 32zm44 0a20 20 0 1 0-22 19.899v-25.07l-7.573 7.586a2.002 2.002 0 0 1-2.829-2.832L30.603 20.56a1.998 1.998 0 0 1 2.828 0l11.005 11.023a2.002 2.002 0 0 1-2.83 2.832L34 26.795v25.104A20 20 0 0 0 52 32z" /></Symbol>;
};

export function CircleCheckmark(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm13.4-35.419a2.001 2.001 0 0 0-3.049.278L29.507 39.25l-5.874-8.411a2.001 2.001 0 0 0-3.05-.278 2.085 2.085 0 0 0-.17 2.67l6.624 9.485a3.014 3.014 0 0 0 4.94 0l13.594-19.464a2.085 2.085 0 0 0-.171-2.67z" /></Symbol>;
};

export function CircleChevronDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 32a24 24 0 1 1 24 24A24 24 0 0 1 8 32zm25.397 9.44l11.005-11.023a2.002 2.002 0 0 0-2.829-2.832l-9.59 9.606-9.59-9.606a2.002 2.002 0 0 0-2.829 2.832L30.568 41.44a1.998 1.998 0 0 0 2.83 0z" /></Symbol>;
};

export function CircleChevronLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 8A24 24 0 1 1 8 32 24 24 0 0 1 32 8zm-9.44 25.398l11.023 11.004a2.002 2.002 0 0 0 2.832-2.829l-9.606-9.59 9.606-9.59a2.002 2.002 0 0 0-2.832-2.83L22.56 30.57a1.998 1.998 0 0 0 0 2.829z" /></Symbol>;
};

export function CircleChevronRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm9.44-25.398L30.417 19.599a2.002 2.002 0 0 0-2.832 2.83l9.606 9.59-9.606 9.59a2.002 2.002 0 0 0 2.832 2.828L41.44 33.432a1.998 1.998 0 0 0 0-2.83z" /></Symbol>;
};

export function CircleChevronUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56 32A24 24 0 1 1 32 8a24 24 0 0 1 24 24zm-25.398-9.44L19.599 33.583a2.002 2.002 0 0 0 2.83 2.832l9.59-9.606 9.59 9.606a2.002 2.002 0 0 0 2.828-2.832L33.432 22.56a1.998 1.998 0 0 0-2.83 0z" /></Symbol>;
};

export function CircleRemove(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20.001 20A20 20 0 0 0 32 12zM17 32a2 2 0 0 1 2-2h26a2 2 0 0 1 0 4H19a2 2 0 0 1-2-2z" /></Symbol>;
};

export function CircleClose(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20.001 20A20.001 20.001 0 0 0 32 12zm2.82 20l7.584 7.584a2.002 2.002 0 1 1-2.83 2.83l-7.585-7.583-7.573 7.573a2.002 2.002 0 0 1-2.83-2.83L29.16 32l-7.603-7.602a2.001 2.001 0 0 1 2.83-2.83l7.603 7.602 7.613-7.613a2.002 2.002 0 0 1 2.83 2.83z" /></Symbol>;
};

export function CircleRemoveAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zM17 32a2 2 0 0 1 2-2h26a2 2 0 1 1 0 4H19a2 2 0 0 1-2-2z" /></Symbol>;
};

export function Clear(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 8A24.001 24.001 0 1 1 8 32 24 24 0 0 1 32 8zM21.585 39.572a2.002 2.002 0 0 0 2.831 2.83l7.573-7.573 7.584 7.584a2.002 2.002 0 1 0 2.83-2.83L34.82 32l7.613-7.613a2.002 2.002 0 0 0-2.83-2.83L31.99 29.17l-7.602-7.602a2.002 2.002 0 1 0-2.83 2.83l7.6 7.602z" /></Symbol>;
};

export function Clipping(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M15 52a3 3 0 0 1-3-3V15a3 3 0 0 1 3-3h34a3 3 0 0 1 3 3v18a3 3 0 0 1-3 3H39a3 3 0 0 0-3 3v10a3 3 0 0 1-3 3zm24-14h10.586a1 1 0 0 1 .707 1.707L39.707 50.293A1 1 0 0 1 38 49.586V39a1 1 0 0 1 1-1z" /></Symbol>;
};

export function Clock(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm9.496-19.55l-6.638-5.98-1.354-13.133a1.51 1.51 0 0 0-3 0l-1.512 14.656a2.378 2.378 0 0 0 1.382 2.404l9.559 4.46a1.452 1.452 0 0 0 1.563-2.407z" /></Symbol>;
};

export function Close(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47.404 47.415a2.001 2.001 0 0 1-2.83 0l-12.6-12.6-12.59 12.589a2.002 2.002 0 1 1-2.83-2.83l12.589-12.59-12.587-12.586a2.002 2.002 0 0 1 2.83-2.83l12.587 12.586 12.598-12.598a2.002 2.002 0 1 1 2.83 2.831L34.804 31.984l12.6 12.6a2.001 2.001 0 0 1 0 2.83z" /></Symbol>;
};

export function Column(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51 51H13a4 4 0 0 1-4-4V17a4 4 0 0 1 4-4h38a4 4 0 0 1 4 4v30a4 4 0 0 1-4 4zM23 18a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v28a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm14 0a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v28a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm14 0a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v28a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Compose(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52.46 16.762a.752.752 0 0 1-1.061 0l-4.191-4.176a.746.746 0 0 1 0-1.057l2.895-2.888a2.255 2.255 0 0 1 3.182 0l2.069 2.062a2.238 2.238 0 0 1 0 3.172zM29.747 38.47L24.7 39.984a.562.562 0 0 1-.7-.7l1.515-5.054a6.004 6.004 0 0 1 1.504-2.519l17.503-17.516a.75.75 0 0 1 1.06 0l4.187 4.19a.75.75 0 0 1 0 1.062L32.268 36.962a6.001 6.001 0 0 1-2.52 1.507zM52 29v17a6.008 6.008 0 0 1-6 6.002H18A6.008 6.008 0 0 1 12 46V17.994a6.008 6.008 0 0 1 6-6.002h17a2 2 0 0 1 0 4.001H18a2.002 2.002 0 0 0-2 2V46a2.002 2.002 0 0 0 2 2h28a2.003 2.003 0 0 0 2-2V29a2 2 0 1 1 4 0z" /></Symbol>;
};

export function Computer(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 15a5 5 0 0 1 5-5h38a5 5 0 0 1 5 5v25a5 5 0 0 1-5 5H13a5 5 0 0 1-5-5zm4 22a1 1 0 0 0 1 1h38a1 1 0 0 0 1-1V15a1 1 0 0 0-1-1H13a1 1 0 0 0-1 1zm11 17a1 1 0 0 1-1-1 1.071 1.071 0 0 1 1-1c1.228-.209 2.49-2.063 3-4 .14-.534.448-1 1-1h10c.552 0 .86.466 1 1 .51 1.937 1.772 3.791 3 4a1.071 1.071 0 0 1 1 1 1 1 0 0 1-1 1z" /></Symbol>;
};

export function Copy(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M41 56H15a3 3 0 0 1-3-3V19a3 3 0 0 1 3-3h7v29a5 5 0 0 0 5 5h17v3a3 3 0 0 1-3 3zm2-34a1 1 0 0 1-1-1V10.414a1 1 0 0 1 1.707-.707l10.586 10.586A1 1 0 0 1 53.586 22zm13 5v18a3 3 0 0 1-3 3H27a3 3 0 0 1-3-3V11a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 1 3 3z" /></Symbol>;
};

export function CreditCardBack(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 49H12a4 4 0 0 1-4-4V19a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v26a4 4 0 0 1-4 4zm0-25a1 1 0 0 0-1-1H13a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h38a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Dashboard(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zM19 44a2 2 0 0 0 2 2h22a2 2 0 0 0 2-2v-2a2 2 0 0 0-2-2H21a2 2 0 0 0-2 2zm13-32a20 20 0 0 0-20 20v.02A2.021 2.021 0 0 0 14.043 34h13.589L25.04 18.834a1.55 1.55 0 0 1 .986-1.752 1.492 1.492 0 0 1 1.814.78L35.446 34h14.511A2.022 2.022 0 0 0 52 32.02V32a20 20 0 0 0-20-20z" /></Symbol>;
};

export function Database(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 54c-11.046 0-20-4.03-20-9v-1.527a1.454 1.454 0 0 1 2.365-1.141C17.802 45.114 24.445 47 32 47s14.198-1.886 17.635-4.668A1.455 1.455 0 0 1 52 43.473V45c0 4.97-8.954 9-20 9zm0-11c-11.046 0-20-4.03-20-9v-1.527a1.454 1.454 0 0 1 2.365-1.141C17.802 34.114 24.445 36 32 36s14.198-1.886 17.635-4.668A1.455 1.455 0 0 1 52 32.473V34c0 4.97-8.954 9-20 9zm0-11c-11.046 0-20-4.03-20-9v-4c0-4.97 8.954-9 20-9s20 4.03 20 9v4c0 4.97-8.954 9-20 9z" /></Symbol>;
};

export function Delete(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M49 20H15a2 2 0 0 1-2-2v-2a2 2 0 0 1 2-2h6c0-2.931.615-4.643 1.985-6.012s3.076-1.986 6.006-1.986h6.018c2.93 0 4.637.616 6.006 1.986S43 11.068 43 14h6a2 2 0 0 1 2 2v2a2 2 0 0 1-2 2zm-10.992-9.006a3.84 3.84 0 0 0-3.003-.993h-6.01a3.84 3.84 0 0 0-3.003.993A3.848 3.848 0 0 0 25 14h14a3.848 3.848 0 0 0-.992-3.006zM46.99 22a1.001 1.001 0 0 1 1.002 1.062l-1.764 28.125A3.007 3.007 0 0 1 43.222 54H20.775a3.007 3.007 0 0 1-3.005-2.813l-1.765-28.125A1.002 1.002 0 0 1 17.007 22z" /></Symbol>;
};

export function Downward(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M25.022 12v19.989h-8.004a2.003 2.003 0 0 0-1.536 3.285L30.48 53.301a1.998 1.998 0 0 0 3.073 0L48.55 35.274a2.003 2.003 0 0 0-1.536-3.285h-8.006V12a2 2 0 0 0-1.998-2H27.02a2 2 0 0 0-1.998 2z" /></Symbol>;
};

export function Document(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M37 22a1 1 0 0 1-1-1V10.414a1 1 0 0 1 1.707-.707l10.586 10.586A1 1 0 0 1 47.586 22zm13 5v26a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3V11a3 3 0 0 1 3-3h14a3 3 0 0 1 3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 1 3 3z" /></Symbol>;
};

export function Download(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46 44h-6a2 2 0 0 1 0-4h5.905A8.058 8.058 0 0 0 54 32.073c.038-5.378-4.133-7.227-6.59-8.064a1.992 1.992 0 0 1-1.34-1.704l-.103-1.081A10.106 10.106 0 0 0 36.137 12c-4.668-.054-7.12 2.677-8.707 4.683a1.993 1.993 0 0 1-1.993.696 9.815 9.815 0 0 0-2.635-.377 6.005 6.005 0 0 0-5.791 5.871l-.031 1.304a1.987 1.987 0 0 1-1.255 1.807c-2.593 1.016-5.887 2.66-5.718 7.357A6.97 6.97 0 0 0 17 40h7a2 2 0 0 1 0 4h-7a10.995 10.995 0 0 1-3.988-21.243A9.983 9.983 0 0 1 25.1 13.225a13.991 13.991 0 0 1 24.833 7.448A11.994 11.994 0 0 1 46 44zm-20.982 3.99H30V28a2 2 0 1 1 4 0v19.99h4.995a1.027 1.027 0 0 1 .82 1.622l-6.99 9.974a.996.996 0 0 1-1.638 0l-6.989-9.974a1.027 1.027 0 0 1 .82-1.623z" /></Symbol>;
};

export function Drafts(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M59.354 21.875l-2.894 2.888a.752.752 0 0 1-1.061 0l-4.191-4.177a.746.746 0 0 1 0-1.057l2.895-2.888a2.255 2.255 0 0 1 3.182 0l2.069 2.062a2.238 2.238 0 0 1 0 3.172zm-5.583 5.572L36.268 44.962a6.002 6.002 0 0 1-2.52 1.507l-5.046 1.515a.563.563 0 0 1-.7-.7l1.514-5.054a6.003 6.003 0 0 1 1.504-2.519l17.503-17.516a.75.75 0 0 1 1.06 0l4.187 4.19a.75.75 0 0 1 0 1.062zM36 22H10a2 2 0 1 1 0-4h26a2 2 0 0 1 0 4zm-26 8h18a2 2 0 0 1 0 4H10a2 2 0 1 1 0-4zm0 12h10a2 2 0 1 1 0 4H10a2 2 0 1 1 0-4z" /></Symbol>;
};

export function Eject(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 44a2 2 0 0 1-2 2H12a2 2 0 0 1-2-2v-4a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2zm-.146-11.234A2 2 0 0 1 52.005 34H11.993a2 2 0 0 1-1.415-3.414l20.007-20a2 2 0 0 1 2.829 0l20.006 20a2 2 0 0 1 .434 2.18z" /></Symbol>;
};

export function Error(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M19.88 10.42A8.728 8.728 0 0 1 25.912 8h12.735a8.727 8.727 0 0 1 6.183 2.568l8.625 8.659A8.727 8.727 0 0 1 56 25.386v12.755a8.727 8.727 0 0 1-2.42 6.031l-8.73 9.132A8.728 8.728 0 0 1 38.542 56H26.016a8.728 8.728 0 0 1-6.16-2.545L10.568 44.2A8.727 8.727 0 0 1 8 38.018V25.51a8.727 8.727 0 0 1 2.696-6.307zM31.994 46a3 3 0 1 0-3-3 3 3 0 0 0 3 3zm-2.446-11.248a2.454 2.454 0 0 0 4.891 0l.555-13.755a3 3 0 0 0-6 0z" /></Symbol>;
};

export function Expand(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M12 54h20a2 2 0 1 0 0-4H12a2 2 0 1 0 0 4zm0-8h6a2 2 0 1 0 0-4h-6a2 2 0 1 0 0 4zm-2-10a2 2 0 0 0 2 2h12a2 2 0 0 0 0-4H12a2 2 0 0 0-2 2zm36-26H18a2 2 0 0 0 0 4h28a2 2 0 0 0 0-4zm-21.814 9.612l6.989 9.974a.996.996 0 0 0 1.638 0l6.988-9.974a1.027 1.027 0 0 0-.819-1.623H25.005a1.027 1.027 0 0 0-.82 1.623zM52 34H32a2 2 0 0 0 0 4h20a2 2 0 1 0 0-4zm0 8h-6a2 2 0 1 0 0 4h6a2 2 0 1 0 0-4zm2 10a2 2 0 0 0-2-2H40a2 2 0 1 0 0 4h12a2 2 0 0 0 2-2zm-14-8a2 2 0 0 0-2-2H26a2 2 0 1 0 0 4h12a2 2 0 0 0 2-2z" /></Symbol>;
};

export function Eyedropper(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.167 26.644l4.19 4.189a2.272 2.272 0 0 1 0 3.218L21.546 49.836c-3.62 3.62-6.879 1.942-8.601 3.664a1.969 1.969 0 0 1-2.51.065 1.97 1.97 0 0 1 .066-2.509c1.722-1.722.045-4.98 3.664-8.6L29.95 26.643a2.272 2.272 0 0 1 3.217 0zm19.31-7.683l-7.425 7.418.963.962a3.317 3.317 0 0 1-4.693 4.689l-9.385-9.377a3.317 3.317 0 1 1 4.692-4.69l.993.993 7.425-7.418a5.251 5.251 0 1 1 7.43 7.423z" /></Symbol>;
};

export function Factory(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 52H10a2 2 0 0 1-2-2V27.994a2 2 0 0 1 2-2h1L12.91 6.9a1 1 0 0 1 .995-.9h3.19a1 1 0 0 1 .995.9l1.83 18.295 10.526-7.014a1 1 0 0 1 1.554.831v6.13l10.445-6.961a1 1 0 0 1 1.555.832v6.128l10.445-6.96a1 1 0 0 1 1.555.831V50a2 2 0 0 1-2 2zM28 33a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm0 10a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm12-10a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm0 10a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm12-10a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm0 10a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Filter(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20zm12 16H20a2 2 0 1 1 0-4h24a2 2 0 1 1 0 4zm-20 4h16a2 2 0 1 1 0 4H24a2 2 0 1 1 0-4zm4 8h8a2 2 0 1 1 0 4h-8a2 2 0 1 1 0-4z" /></Symbol>;
};

export function Feed(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M49 54H15a5 5 0 0 1-5-5V15a5 5 0 0 1 5-5h34a5 5 0 0 1 5 5v34a5 5 0 0 1-5 5zM22 15a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm27 0H27a1 1 0 1 0 0 2h22a1 1 0 1 0 0-2zm-23 5a1 1 0 0 0 1 1h14a1 1 0 1 0 0-2H27a1 1 0 0 0-1 1zm24 7a1 1 0 0 0-1-1H15a1 1 0 0 0-1 1v22a1 1 0 0 0 1 1h34a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Fire(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M38.256 54.63a15.562 15.562 0 0 0 4.724-7 .516.516 0 0 0-.489-.667 27.936 27.936 0 0 0-6.802.973.517.517 0 0 1-.663-.68A19.953 19.953 0 0 0 36 39.98c0-5.301-3.219-9.405-4.351-10.813a.495.495 0 0 0-.86.15C29.28 34.134 23 41.046 23 47.973c0 4.163 2.765 5.971 2.969 6.937.167.795-.237 1.235-.998.995C19.907 54.314 12 47.03 12 37.98c0-16.44 16-19.72 16-27.982a7.513 7.513 0 0 0-.952-3.002.762.762 0 0 1 .93-.97C33.121 7.411 45 14.592 45 28.986a24.616 24.616 0 0 1-2 8.994s3.362-.584 7.92-1.003a1.052 1.052 0 0 1 1.077 1.172c0 6.94-5.422 14.941-13.137 17.824-.7.261-1.167-.836-.604-1.343z" /></Symbol>;
};

export function FlagSwallowtail(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M20 13.857a2.538 2.538 0 0 1 2-2.42c10.17-1.645 20.34 2.136 30.509.997a1.355 1.355 0 0 1 1.302 2.097 330.644 330.644 0 0 1-6.575 12.021.511.511 0 0 0 0 .503c2.191 3.936 4.383 7.726 6.575 11.267a1.707 1.707 0 0 1-1.302 2.445C42.339 41.907 32.17 38.125 22 39.77a1.621 1.621 0 0 1-2-1.627 2.538 2.538 0 0 1 2-2.42V15.485a1.621 1.621 0 0 1-2-1.628zM18 12v44a2 2 0 1 1-4 0V12a2 2 0 1 1 4 0z" /></Symbol>;
};

export function Flag(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 40.563c-10.667 1.726-21.333-2.518-32-.793a1.621 1.621 0 0 1-2-1.627 2.538 2.538 0 0 1 2-2.42V15.485a1.621 1.621 0 0 1-2-1.628 2.538 2.538 0 0 1 2-2.42c10.667-1.725 21.333 2.519 32 .793a1.621 1.621 0 0 1 2 1.627v24.286a2.538 2.538 0 0 1-2 2.42zM14 56V12a2 2 0 1 1 4 0v44a2 2 0 1 1-4 0z" /></Symbol>;
};

export function FlagPennant(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M14 56V12a2 2 0 0 1 4 0v44a2 2 0 0 1-4 0zm42-30a2.6 2.6 0 0 1-1.298 2.17c-10.667 6.04-21.333 5.85-32 11.367a2.363 2.363 0 0 1-.702.233 1.879 1.879 0 0 1-1.14-.153 1.582 1.582 0 0 1-.86-1.474 3.005 3.005 0 0 1 1.51-2.52q.245-.136.49-.268V15.853q-.245-.053-.49-.102A1.841 1.841 0 0 1 20 13.858a2.577 2.577 0 0 1 .86-1.853 2.254 2.254 0 0 1 1.14-.568 1.776 1.776 0 0 1 .702.024c10.667 2.578 21.333 10.864 32 12.918A1.573 1.573 0 0 1 56 26z" /></Symbol>;
};

export function Flash(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47.795 28.176a1.999 1.999 0 0 1-.32 2.145l-21.984 25a1.999 1.999 0 0 1-3.376-2.012L28.12 37H17.995a2 2 0 0 1-1.5-3.321l21.985-25a1.999 1.999 0 0 1 3.375 2.013L35.85 27h10.125a1.997 1.997 0 0 1 1.82 1.176z" /></Symbol>;
};

export function Folder(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M55.99 28.153L54.368 49.23A3.002 3.002 0 0 1 51.372 52H12.628a3.002 3.002 0 0 1-2.995-2.77L8.009 28.153A2.001 2.001 0 0 1 10.006 26h43.988a2.001 2.001 0 0 1 1.997 2.153zM54 20v3a1 1 0 0 1-1 1H11a1 1 0 0 1-1-1v-9a2 2 0 0 1 2-2h11.343a4 4 0 0 1 2.829 1.172l3.656 3.657A4 4 0 0 0 32.657 18H52a2 2 0 0 1 2 2z" /></Symbol>;
};

export function Following(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M57.635 25.149l-10.491 15.01a1.997 1.997 0 0 1-3.21.09l-5.495-7.006a2 2 0 1 1 3.143-2.473l3.832 4.884 8.946-12.8a2 2 0 1 1 3.275 2.295zM34.145 43.56c4.53 1.681 6.59 2.84 6.853 7.37A1.01 1.01 0 0 1 39.995 52H7.005a1.011 1.011 0 0 1-1.003-1.07c.263-4.53 2.323-5.688 6.852-7.369 4.35-1.614 5.839-3.21 6.093-7.163a9.474 9.474 0 0 1-1.993-1.716 9.546 9.546 0 0 1-1.905-4.713 1.176 1.176 0 0 1-.143.019c-1.085 0-1.906-2.282-1.906-3.886s.557-2.093 1.089-2.093c.114 0 .209.014.31.022A13.571 13.571 0 0 1 14 20.938C14 15.247 16.303 12 23.5 12c3.428 0 3.66.848 4.318 2.031a2.064 2.064 0 0 1 1.295-.406c1.83 0 3.887 2.21 3.887 7.313a13.59 13.59 0 0 1-.399 3.094c.102-.01.196-.023.31-.023.532 0 1.09.488 1.09 2.093s-.822 3.886-1.906 3.886a1.176 1.176 0 0 1-.144-.02 9.546 9.546 0 0 1-1.905 4.714 9.475 9.475 0 0 1-1.994 1.716c.255 3.953 1.745 5.55 6.093 7.163z" /></Symbol>;
};

export function Follow(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56 34h-6v6a2 2 0 0 1-4 0v-6h-6a2 2 0 0 1 0-4h6v-6a2 2 0 0 1 4 0v6h6a2 2 0 0 1 0 4zm-21.854 9.561c4.529 1.681 6.588 2.84 6.852 7.37A1.01 1.01 0 0 1 39.995 52H7.005a1.011 1.011 0 0 1-1.003-1.07c.263-4.53 2.323-5.688 6.852-7.369 4.35-1.614 5.838-3.21 6.093-7.163a9.471 9.471 0 0 1-1.993-1.716 9.546 9.546 0 0 1-1.905-4.713 1.176 1.176 0 0 1-.144.019c-1.084 0-1.905-2.282-1.905-3.886s.557-2.093 1.089-2.093c.114 0 .209.014.31.022A13.571 13.571 0 0 1 14 20.938C14 15.247 16.303 12 23.5 12c3.428 0 3.66.848 4.318 2.031a2.063 2.063 0 0 1 1.295-.406c1.83 0 3.887 2.21 3.887 7.313a13.565 13.565 0 0 1-.4 3.094c.103-.01.197-.023.31-.023.532 0 1.09.488 1.09 2.093s-.821 3.886-1.906 3.886a1.178 1.178 0 0 1-.144-.02 9.546 9.546 0 0 1-1.905 4.714 9.486 9.486 0 0 1-1.993 1.716c.255 3.953 1.744 5.55 6.094 7.163z" /></Symbol>;
};

export function Forward(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M35.274 48.518a2.003 2.003 0 0 1-3.285-1.536V39H12a2 2 0 0 1-2-2V27a2 2 0 0 1 2-2h19.989v-8.015a2.003 2.003 0 0 1 3.285-1.536l18.027 14.998a1.998 1.998 0 0 1 0 3.073z" /></Symbol>;
};

export function Friends(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M62.996 52l-25.016-.003a.995.995 0 0 1-.997-.94 9.393 9.393 0 0 0-1.705-5.507 1.25 1.25 0 0 1 .517-1.96l.078-.03c4.34-1.611 5.834-3.198 6.092-7.14a9.384 9.384 0 0 1-2.01-1.738 9.476 9.476 0 0 1-1.908-4.68 1.407 1.407 0 0 1-.146.01c-1.086 0-1.908-2.29-1.908-3.9s.558-2.1 1.09-2.1c.111 0 .215.005.314.013A13.567 13.567 0 0 1 37 20.938c0-3.904 1.095-6.606 4.44-7.087C43.29 12.567 46.088 12 49.662 12c2.287 0 .382 1.564 2.53 2.02 1.33.284 3.807 1.226 3.807 6.918a13.568 13.568 0 0 1-.397 3.088c.1-.008.203-.014.314-.014.532 0 1.09.49 1.09 2.1s-.821 3.9-1.907 3.9a1.408 1.408 0 0 1-.147-.01 9.473 9.473 0 0 1-1.907 4.68 9.35 9.35 0 0 1-1.988 1.724c.255 3.953 1.747 5.542 6.092 7.156 4.527 1.68 6.585 2.838 6.848 7.369A1.01 1.01 0 0 1 62.996 52zM9.998 36.004c-1.235 0-5-3.067-5-10.402 0-8.595 4.935-13.603 10.835-13.603 2.824 0 4.153.13 5.667 1.501 4.322 0 8.503 3.206 8.503 12.102 0 7.278-3.728 10.402-5.001 10.402-.444 0-.78-.388-1.277-.894a10.062 10.062 0 0 1-1.668 1.338c.263 3.922 1.761 5.505 6.095 7.113 4.53 1.681 6.59 2.839 6.854 7.37A1.01 1.01 0 0 1 34.002 52H1.005a1.011 1.011 0 0 1-1.003-1.07c.264-4.53 2.324-5.688 6.854-7.369 4.332-1.607 5.83-3.19 6.095-7.107a10.06 10.06 0 0 1-1.684-1.353c-.502.51-.836.903-1.269.903z" /></Symbol>;
};

export function FriendsMen(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M62.996 52l-25.016-.003a.995.995 0 0 1-.997-.94 9.39 9.39 0 0 0-1.705-5.507 1.251 1.251 0 0 1 .517-1.96l.078-.03c4.34-1.611 5.835-3.198 6.092-7.14a9.388 9.388 0 0 1-2.01-1.738 9.473 9.473 0 0 1-1.908-4.68 1.415 1.415 0 0 1-.146.01c-1.086 0-1.908-2.29-1.908-3.9s.558-2.1 1.09-2.1c.111 0 .215.005.314.013A13.561 13.561 0 0 1 37 20.938c0-3.904 1.095-6.606 4.44-7.087C43.29 12.567 46.088 12 49.662 12c2.287 0 .382 1.564 2.53 2.02 1.33.284 3.807 1.226 3.807 6.917a13.567 13.567 0 0 1-.397 3.089c.099-.008.202-.014.314-.014.532 0 1.09.49 1.09 2.1s-.822 3.9-1.908 3.9a1.413 1.413 0 0 1-.147-.01 9.476 9.476 0 0 1-1.906 4.68 9.362 9.362 0 0 1-1.989 1.724c.255 3.953 1.747 5.541 6.093 7.155 4.526 1.681 6.584 2.839 6.848 7.37A1.01 1.01 0 0 1 62.995 52zM9.719 16.062c1.144-2.992 4.942-4.063 8.646-4.063 2.37 0 3.115.433 3.458 1.625a3.18 3.18 0 0 1 3.458 2.438c.893.312 1.73 1.403 1.73 4.875a13.536 13.536 0 0 1-.398 3.088c.097-.008.197-.013.304-.013.532 0 1.09.49 1.09 2.1s-.822 3.9-1.908 3.9c-.048 0-.092-.004-.137-.009a9.47 9.47 0 0 1-1.91 4.68 9.367 9.367 0 0 1-1.998 1.73c.257 3.947 1.75 5.535 6.098 7.148 4.53 1.681 6.59 2.839 6.854 7.37A1.011 1.011 0 0 1 34.003 52H1.005a1.011 1.011 0 0 1-1.003-1.07c.264-4.53 2.324-5.688 6.854-7.369 4.345-1.612 5.84-3.2 6.097-7.144a9.378 9.378 0 0 1-2.006-1.734 9.473 9.473 0 0 1-1.91-4.68c-.044.004-.088.01-.136.01-1.086 0-1.908-2.29-1.908-3.9s.558-2.1 1.09-2.1c.107 0 .207.004.304.012a13.536 13.536 0 0 1-.398-3.088c0-2.723.558-4.62 1.73-4.875z" /></Symbol>;
};

export function FullScreenExit(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M57 48h-4v4a2 2 0 0 1-4 0v-6a2 2 0 0 1 2-2h6a2 2 0 0 1 0 4zm0-28h-6a2 2 0 0 1-2-2v-6a2 2 0 0 1 4 0v4h4a2 2 0 0 1 0 4zm-46-8a2 2 0 0 1 4 0v6a2 2 0 0 1-2 2H7a2 2 0 1 1 0-4h4zM7 44h6a2 2 0 0 1 2 2v6a2 2 0 0 1-4 0v-4H7a2 2 0 1 1 0-4zm11.5-2a1.5 1.5 0 0 1-1.5-1.5v-17a1.5 1.5 0 0 1 1.5-1.5h27a1.5 1.5 0 0 1 1.5 1.5v17a1.5 1.5 0 0 1-1.5 1.5z" /></Symbol>;
};

export function FullScreen(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M58 52h-6a2 2 0 0 1 0-4h4v-4a2 2 0 0 1 4 0v6a2 2 0 0 1-2 2zm0-30a2 2 0 0 1-2-2v-4h-4a2 2 0 0 1 0-4h6a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2zm-44-8a2 2 0 0 1-2 2H8v4a2 2 0 0 1-4 0v-6a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2zm-3 29.5v-23a1.5 1.5 0 0 1 1.5-1.5h39a1.5 1.5 0 0 1 1.5 1.5v23a1.5 1.5 0 0 1-1.5 1.5h-39a1.5 1.5 0 0 1-1.5-1.5zM6 42a2 2 0 0 1 2 2v4h4a2 2 0 0 1 0 4H6a2 2 0 0 1-2-2v-6a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Gear(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M57.274 34.208l-3.988 1.139a1.001 1.001 0 0 0-.703.773 20.72 20.72 0 0 1-.7 2.599 1 1 0 0 0 .22 1.022l2.889 2.984a1.001 1.001 0 0 1 .147 1.197l-1.245 2.156a1 1 0 0 1-1.11.47l-4.05-1.014a1 1 0 0 0-.994.317 21.207 21.207 0 0 1-1.887 1.892 1 1 0 0 0-.317.994l1.013 4.047a1 1 0 0 1-.47 1.11l-2.156 1.245a1 1 0 0 1-1.197-.148l-3-2.901a1 1 0 0 0-1.022-.222 20.774 20.774 0 0 1-2.58.702 1.002 1.002 0 0 0-.773.702l-1.143 4.002a1.001 1.001 0 0 1-.963.726h-2.49a1.001 1.001 0 0 1-.963-.726l-1.143-4.001a1.002 1.002 0 0 0-.773-.703 20.765 20.765 0 0 1-2.587-.696 1 1 0 0 0-1.022.222l-2.992 2.896a1 1 0 0 1-1.197.147l-2.156-1.245a1.001 1.001 0 0 1-.47-1.11l1.014-4.052a1.001 1.001 0 0 0-.318-.996 21.142 21.142 0 0 1-1.894-1.883 1 1 0 0 0-.993-.317l-4.045 1.013a1.001 1.001 0 0 1-1.11-.47l-1.245-2.157a1.001 1.001 0 0 1 .147-1.197l2.894-2.99a1 1 0 0 0 .222-1.023 20.774 20.774 0 0 1-.707-2.592 1.001 1.001 0 0 0-.702-.772l-3.989-1.14A1 1 0 0 1 6 33.245v-2.49a1 1 0 0 1 .726-.963l3.988-1.14a1.001 1.001 0 0 0 .702-.773 20.777 20.777 0 0 1 .701-2.599 1 1 0 0 0-.221-1.022l-2.888-2.983a1.001 1.001 0 0 1-.147-1.197l1.245-2.156a1.001 1.001 0 0 1 1.11-.47l4.05 1.014a1 1 0 0 0 .995-.317 21.19 21.19 0 0 1 1.886-1.892 1 1 0 0 0 .317-.994l-1.013-4.047a1 1 0 0 1 .47-1.11l2.156-1.245a1.001 1.001 0 0 1 1.197.148l3 2.901a1 1 0 0 0 1.022.222 20.821 20.821 0 0 1 2.58-.702 1.001 1.001 0 0 0 .773-.702l1.143-4.002A1.001 1.001 0 0 1 30.756 6h2.49a1.001 1.001 0 0 1 .963.726l1.143 4.001a1.001 1.001 0 0 0 .773.703 20.747 20.747 0 0 1 2.587.696 1 1 0 0 0 1.022-.222l2.992-2.895a1 1 0 0 1 1.197-.148l2.156 1.245a1 1 0 0 1 .47 1.11l-1.014 4.052a1 1 0 0 0 .318.996 21.155 21.155 0 0 1 1.894 1.884 1 1 0 0 0 .993.316l4.045-1.013a1 1 0 0 1 1.11.47l1.245 2.157a1 1 0 0 1-.148 1.196l-2.893 2.99a1 1 0 0 0-.222 1.024 20.775 20.775 0 0 1 .707 2.592 1.001 1.001 0 0 0 .701.772l3.99 1.14a1 1 0 0 1 .726.963v2.49a1 1 0 0 1-.726.963zM22.632 19.774a.506.506 0 0 0-.753-.144 15.953 15.953 0 0 0 0 24.74.506.506 0 0 0 .753-.145l5.76-9.975a4.5 4.5 0 0 0 0-4.5zM32 16a15.959 15.959 0 0 0-5.666 1.033.503.503 0 0 0-.25.721l5.772 9.996A4.5 4.5 0 0 0 35.753 30h11.525a.506.506 0 0 0 .501-.58A15.993 15.993 0 0 0 32 16zm15.278 18H35.753a4.5 4.5 0 0 0-3.897 2.25l-5.772 9.996a.503.503 0 0 0 .25.721A15.998 15.998 0 0 0 47.779 34.58a.506.506 0 0 0-.501-.58z" /></Symbol>;
};

export function Globe(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32.002 56h-.004a24 24 0 1 1 .004 0zm-10.467-8.095a27.645 27.645 0 0 1-1.531-3.57.496.496 0 0 0-.467-.335h-2.479a.5.5 0 0 0-.392.812 20.09 20.09 0 0 0 4.155 3.742.498.498 0 0 0 .714-.65zM14.023 40h4.189a.492.492 0 0 0 .485-.587 38.981 38.981 0 0 1-.613-4.943.496.496 0 0 0-.493-.47H12.67a.507.507 0 0 0-.503.565 19.875 19.875 0 0 0 1.388 5.126.507.507 0 0 0 .468.309zM12.67 30h4.922a.496.496 0 0 0 .492-.47 38.98 38.98 0 0 1 .613-4.943.492.492 0 0 0-.485-.587h-4.189a.507.507 0 0 0-.468.308 19.877 19.877 0 0 0-1.388 5.127.507.507 0 0 0 .503.565zm4.388-10h2.48a.496.496 0 0 0 .466-.335 27.608 27.608 0 0 1 1.531-3.57.498.498 0 0 0-.714-.648 20.09 20.09 0 0 0-4.155 3.741.5.5 0 0 0 .392.812zm12.181-7.226c-1.97 1.132-3.685 3.451-4.954 6.536a.504.504 0 0 0 .467.69H29.5a.5.5 0 0 0 .5-.5v-6.291a.506.506 0 0 0-.76-.435zM29.5 24h-6.253a.502.502 0 0 0-.494.398 36.66 36.66 0 0 0-.672 5.066.507.507 0 0 0 .504.536H29.5a.5.5 0 0 0 .5-.5v-5a.5.5 0 0 0-.5-.5zm0 10h-6.915a.507.507 0 0 0-.504.536 36.66 36.66 0 0 0 .672 5.066.502.502 0 0 0 .494.398H29.5a.5.5 0 0 0 .5-.5v-5a.5.5 0 0 0-.5-.5zm0 10h-4.748a.504.504 0 0 0-.467.69c1.269 3.085 2.984 5.404 4.954 6.536A.506.506 0 0 0 30 50.79V44.5a.5.5 0 0 0-.5-.5zm21.83-10h-4.922a.496.496 0 0 0-.492.47 38.944 38.944 0 0 1-.613 4.943.492.492 0 0 0 .485.587h4.189a.507.507 0 0 0 .467-.308 19.862 19.862 0 0 0 1.389-5.127.507.507 0 0 0-.503-.565zm-4.388 10h-2.48a.496.496 0 0 0-.466.335 27.645 27.645 0 0 1-1.53 3.57.498.498 0 0 0 .713.648 20.088 20.088 0 0 0 4.155-3.741.5.5 0 0 0-.392-.812zM34.76 51.226c1.97-1.132 3.685-3.451 4.954-6.536a.504.504 0 0 0-.467-.69H34.5a.5.5 0 0 0-.5.5v6.292a.506.506 0 0 0 .76.434zM34.5 40h6.253a.502.502 0 0 0 .494-.398 36.675 36.675 0 0 0 .672-5.066.507.507 0 0 0-.503-.536H34.5a.5.5 0 0 0-.5.5v5a.5.5 0 0 0 .5.5zM34 13.209V19.5a.5.5 0 0 0 .5.5h4.748a.504.504 0 0 0 .467-.69c-1.268-3.085-2.984-5.404-4.954-6.536a.506.506 0 0 0-.761.435zM34 24.5v5a.5.5 0 0 0 .5.5h6.916a.507.507 0 0 0 .503-.536 36.707 36.707 0 0 0-.672-5.066.502.502 0 0 0-.494-.398H34.5a.5.5 0 0 0-.5.5zm8.465-8.405a27.638 27.638 0 0 1 1.531 3.57.496.496 0 0 0 .467.335h2.479a.5.5 0 0 0 .392-.812 20.09 20.09 0 0 0-4.155-3.741.498.498 0 0 0-.714.648zm2.838 8.492a38.944 38.944 0 0 1 .613 4.944.496.496 0 0 0 .492.469h4.922a.507.507 0 0 0 .503-.565 19.869 19.869 0 0 0-1.388-5.127.507.507 0 0 0-.468-.308h-4.189a.492.492 0 0 0-.485.587z" /></Symbol>;
};

export function Grid(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47 50H37a3 3 0 0 1-3-3V37a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v10a3 3 0 0 1-3 3zm0-20H37a3 3 0 0 1-3-3V17a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v10a3 3 0 0 1-3 3zm-30 4h10a3 3 0 0 1 3 3v10a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3V37a3 3 0 0 1 3-3zm-3-17a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v10a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3z" /></Symbol>;
};

export function Hand(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50.995 29a3.513 3.513 0 0 1 2.858 1.156.706.706 0 0 1 0 .826c-1.012 1.426-5.087 6.64-6.896 12.307C44.005 52.533 37.796 56 29.998 56A14 14 0 0 1 16 42V20a2 2 0 1 1 4 0v12a1 1 0 0 0 1 1 1 1 0 0 0 1-1V13.5a2.5 2.5 0 1 1 5 0V31a1 1 0 0 0 1 1 1 1 0 0 0 1-1V10.5a2.5 2.5 0 1 1 4.999 0V31a1 1 0 0 0 1 1 1 1 0 0 0 1-1V13.5a2.5 2.5 0 1 1 4.999 0v24.054a.659.659 0 0 0 1.225.34C44.602 33.886 47.897 29 50.995 29z" /></Symbol>;
};

export function Headphones(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56 32v12a4 4 0 0 1-4 4h-2v1a3 3 0 0 1-3 3h-2a3 3 0 0 1-3-3V35a3 3 0 0 1 3-3h2a3 3 0 0 1 3 3v1h2v-4a20 20 0 0 0-40 0v4h2v-1a3 3 0 0 1 3-3h2a3 3 0 0 1 3 3v14a3 3 0 0 1-3 3h-2a3 3 0 0 1-3-3v-1h-2a4 4 0 0 1-4-4V32a24 24 0 1 1 48 0z" /></Symbol>;
};

export function Help(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20zm2.48 23.014A2.37 2.37 0 0 1 32 37.013a1.956 1.956 0 0 1-2-2.001c0-4.966 5.5-5.451 5.5-9.006a3.18 3.18 0 0 0-3.5-3.002c-2.787 0-3.656 1.595-4 3.002-.332 1.356-1.03 2.001-2 2.001a1.96 1.96 0 0 1-2-2.001A7.683 7.683 0 0 1 32 18a7.675 7.675 0 0 1 8 8.005c0 4.878-4.595 5.07-5.52 9.008zM31.993 40a3 3 0 1 1-3 3 3 3 0 0 1 3-3z" /></Symbol>;
};

export function HeartEmpty(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56C26.774 56 7 43.4 7 26.385A14.223 14.223 0 0 1 21.023 12 13.904 13.904 0 0 1 32 17.445 13.904 13.904 0 0 1 42.977 12 14.223 14.223 0 0 1 57 26.385C57 43.4 37.225 56 32 56zm10.977-40c-6.482 0-8.681 6-10.977 6s-4.495-6-10.977-6A10.22 10.22 0 0 0 11 26.385C11 41.485 29.383 51.969 32 52c2.617-.031 21-10.514 21-25.615A10.22 10.22 0 0 0 42.977 16z" /></Symbol>;
};

export function Heart(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M20.545 13A12.547 12.547 0 0 1 32 20.349a12.569 12.569 0 0 1 24 5.036C56 41.3 36.027 55 32 55S8 41.3 8 25.384A12.466 12.466 0 0 1 20.545 13z" /></Symbol>;
};

export function History(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M39.933 38.888l-9.56-4.46a2.378 2.378 0 0 1-1.382-2.404c.02-.175 1.494-14.48 1.513-14.656a1.51 1.51 0 0 1 3 0L34.858 30.5l6.638 5.981a1.451 1.451 0 0 1-1.563 2.407zM5.005 31.99h4.994A22 22 0 0 1 32 10a22 22 0 0 1 0 44 21.912 21.912 0 0 1-13.953-4.989 2 2 0 0 1 2.589-3.051A18 18 0 1 0 32 14 18 18 0 0 0 14 31.99h4.983a1.027 1.027 0 0 1 .819 1.622l-6.989 9.974a.996.996 0 0 1-1.638 0l-6.99-9.974a1.027 1.027 0 0 1 .82-1.623z" /></Symbol>;
};

export function Home(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48 10a2 2 0 0 1 2 2v8.525a.75.75 0 0 1-1.28.53l-6.134-6.136A2 2 0 0 1 42 13.505V12a2 2 0 0 1 2-2zm0 44h-9a2 2 0 0 1-2-2V40a2 2 0 0 0-2-2h-6a2 2 0 0 0-2 2v12a2 2 0 0 1-2 2h-9a2 2 0 0 1-2-2V33.987h-3.99a2 2 0 0 1-1.414-3.415L30.59 8.565a1.999 1.999 0 0 1 2.828 0l21.994 22.007a2 2 0 0 1-1.414 3.415H50V52a2 2 0 0 1-2 2z" /></Symbol>;
};

export function Inbox(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M49 54H15a5 5 0 0 1-5-5v-7.672a20.007 20.007 0 0 1 .175-2.644l3.246-24.345A5 5 0 0 1 18.378 10h27.244a5 5 0 0 1 4.956 4.34l3.246 24.344A19.956 19.956 0 0 1 54 41.328V49a5 5 0 0 1-5 5zm.814-15.132l-3.2-24a1 1 0 0 0-.991-.868H18.377a1 1 0 0 0-.99.868l-3.2 24a1 1 0 0 0 .99 1.132h7.102a1 1 0 0 1 .949.684l1.544 4.632a1 1 0 0 0 .949.684h12.558a1 1 0 0 0 .949-.684l1.544-4.632a1 1 0 0 1 .949-.684h7.101a1 1 0 0 0 .992-1.132z" /></Symbol>;
};

export function Info(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20zm1.5 12a3.28 3.28 0 0 1-3.5-3 3.28 3.28 0 0 1 3.5-3 3.279 3.279 0 0 1 3.5 3 3.279 3.279 0 0 1-3.5 3zm1.4 8.5l-1.769 8.843a1 1 0 0 0 .426 1.028l1.863 1.242A1.303 1.303 0 0 1 34.697 46h-3.089a4.597 4.597 0 0 1-4.508-5.499l1.769-8.844a1 1 0 0 0-.426-1.028l-1.863-1.242A1.303 1.303 0 0 1 27.303 27h3.09a4.597 4.597 0 0 1 4.507 5.5z" /></Symbol>;
};

export function Invitation(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M53 54H11a3 3 0 0 1-3-3V33.22a1.205 1.205 0 0 1 1.874-1.002L29.227 45.16a5 5 0 0 0 5.546 0l19.352-12.942A1.205 1.205 0 0 1 56 33.22V51a3 3 0 0 1-3 3zm1.664-24.504l-4.979 3.32A1.084 1.084 0 0 1 48 31.914V25a3 3 0 0 0-3-3H19a3 3 0 0 0-3 3v6.914a1.084 1.084 0 0 1-1.685.901l-4.979-3.319A3 3 0 0 1 8 27v-2a3 3 0 0 1 1.336-2.496l21-14a3 3 0 0 1 3.328 0l21 14A3 3 0 0 1 56 25v2a3 3 0 0 1-1.336 2.496z" /></Symbol>;
};

export function JustifyAll(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 50H14a2 2 0 0 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 0 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 0 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 0 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 0 1 0-4h36a2 2 0 1 1 0 4z" /></Symbol>;
};

export function JustifyCenter(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M18 48a2 2 0 0 1 2-2h24a2 2 0 0 1 0 4H20a2 2 0 0 1-2-2zm32-6H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4z" /></Symbol>;
};

export function JustifyRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 50H26a2 2 0 0 1 0-4h24a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 1 1 0 4z" /></Symbol>;
};

export function Key(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M13.414 53.414l-2.828-2.828A2 2 0 0 1 10 49.172v-2.343a2 2 0 0 1 .586-1.414l17.12-17.12a1 1 0 0 0 .292-.707V23.39a13.252 13.252 0 0 1 13.005-13.386 13 13 0 0 1 12.994 12.993 13.252 13.252 0 0 1-13.386 13.005h-6.198a1 1 0 0 0-.707.293l-1.415 1.414a1 1 0 0 0-.292.707v2.172a1 1 0 0 1-.293.707l-1.414 1.414a1 1 0 0 1-.708.293h-1.171a1 1 0 0 0-.707.293l-1.414 1.414a1 1 0 0 0-.293.707v1.171a1 1 0 0 1-.293.708l-1.414 1.414a1 1 0 0 1-.707.293h-2.171a1 1 0 0 0-.708.293l-1.414 1.414a1 1 0 0 0-.293.707v1.171a1 1 0 0 1-.293.707l-.414.414a1 1 0 0 1-.707.293h-2.757a2 2 0 0 1-1.414-.586zm31.584-31.412a3 3 0 1 0-3-3 3 3 0 0 0 3 3zm-30.285 26.7L30.705 32.71a1 1 0 0 0-1.414-1.414L13.299 47.288a1 1 0 1 0 1.414 1.414z" /></Symbol>;
};

export function JustifyLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 42H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zm0-8H14a2 2 0 1 1 0-4h36a2 2 0 0 1 0 4zM12 48a2 2 0 0 1 2-2h24a2 2 0 0 1 0 4H14a2 2 0 0 1-2-2z" /></Symbol>;
};

export function Keyboard(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M53 48H11a5 5 0 0 1-5-5V21a5 5 0 0 1 5-5h42a5 5 0 0 1 5 5v22a5 5 0 0 1-5 5zm-35-5a1 1 0 0 0 1 1h26a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1H19a1 1 0 0 0-1 1zm-8 0a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1zm8-19a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm1 7h-8a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1zm7-7a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1zm1 7h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1zm7-7a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1zm1 7h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1zm7-7a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1zm1 7h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1v-3a1 1 0 0 0-1-1zm11-7a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm0 8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v3a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Laptop(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M57 51H7a3 3 0 0 1-3-3 1 1 0 0 1 1-1h5V18a5 5 0 0 1 5-5h34a5 5 0 0 1 5 5v29h5a1 1 0 0 1 1 1 3 3 0 0 1-3 3zm-7-33a1 1 0 0 0-1-1H15a1 1 0 0 0-1 1v25a1 1 0 0 0 1 1h34a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Layers(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M11.326 20.284l18.844-9.903a4.343 4.343 0 0 1 3.597 0l18.844 9.903c1.795.944 1.795 2.488 0 3.432l-18.844 9.903a4.343 4.343 0 0 1-3.597 0l-18.844-9.903c-1.795-.944-1.795-2.488 0-3.432zm0 20l3.164-1.662a2 2 0 0 1 1.86 0l12.906 6.775a6.332 6.332 0 0 0 5.46 0l12.888-6.767a2 2 0 0 1 1.86 0l3.147 1.654c1.795.944 1.795 2.488 0 3.432l-18.844 9.903a4.343 4.343 0 0 1-3.597 0l-18.844-9.903c-1.795-.944-1.795-2.488 0-3.431zm0-10l3.164-1.662a2 2 0 0 1 1.86 0l12.906 6.775a6.332 6.332 0 0 0 5.46 0l12.888-6.767a2 2 0 0 1 1.86 0l3.147 1.654c1.795.943 1.795 2.488 0 3.431L33.767 43.62a4.343 4.343 0 0 1-3.597 0l-18.844-9.903c-1.795-.944-1.795-2.488 0-3.431z" /></Symbol>;
};

export function Link(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M23.98 33.99a9.973 9.973 0 0 1-7.072-2.93l-4.002-4.002a10.004 10.004 0 0 1 14.147-14.15l4.001 4.002a10.01 10.01 0 0 1 1.797 11.703l2.507 2.508a10.005 10.005 0 0 1 11.7 1.797l4.002 4.002a10.004 10.004 0 1 1-14.147 14.15l-4-4.002a10.012 10.012 0 0 1-1.798-11.703l-2.507-2.508a10 10 0 0 1-4.627 1.133zm12.47 9.54l4 4.002a5.002 5.002 0 1 0 7.072-7.074l-4-4.002a4.974 4.974 0 0 0-4.228-1.4l3.813 3.816a3 3 0 1 1-4.243 4.244L35.05 39.3a4.976 4.976 0 0 0 1.4 4.229zm-8.933-23.083l-4-4.002a5.002 5.002 0 1 0-7.074 7.075l4.001 4.002a4.747 4.747 0 0 0 4.229 1.399l-3.814-3.815a3 3 0 1 1 4.243-4.244l3.814 3.815a4.978 4.978 0 0 0-1.399-4.23z" /></Symbol>;
};

export function List(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48 49H26a2 2 0 0 1 0-4h22a2 2 0 0 1 0 4zm0-10H26a2 2 0 0 1 0-4h22a2 2 0 0 1 0 4zm0-10H26a2 2 0 0 1-2-2 2 2 0 0 1 2-2h22a2 2 0 0 1 2 2 2 2 0 0 1-2 2zm0-10H26a2 2 0 0 1 0-4h22a2 2 0 0 1 0 4zm-34 7a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2zm2 18h2a2 2 0 0 1 2 2v2a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2v-2a2 2 0 0 1 2-2zm-2-8a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2zm6-18a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2v-2a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2z" /></Symbol>;
};

export function Lock(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M44 28h1a4.177 4.177 0 0 1 4 4.333v17.334A4.177 4.177 0 0 1 45 54H19a4.177 4.177 0 0 1-4-4.333V32.333A4.177 4.177 0 0 1 19 28h1v-7.005c0-4.03.846-6.379 2.729-8.262s4.23-2.73 8.258-2.73h2.026c4.029 0 6.376.847 8.258 2.73S44 16.965 44 20.995zm-4-7.003c0-2.565-.539-4.06-1.737-5.258s-2.691-1.737-5.255-1.737h-2.016c-2.564 0-4.058.539-5.255 1.737S24 18.432 24 20.997V28h16z" /></Symbol>;
};

export function Location(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.586 54.753a2.129 2.129 0 0 1-4.03-.506L26.774 38.94a2.128 2.128 0 0 0-1.713-1.714L9.753 34.443a2.128 2.128 0 0 1-.506-4.03L48.97 12.208a2.129 2.129 0 0 1 2.822 2.822z" /></Symbol>;
};

export function LoopedSquare(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45.5 37a8.5 8.5 0 1 1-8.5 8.5V41H27v4.5a8.5 8.5 0 1 1-8.5-8.5H23V27h-4.5a8.5 8.5 0 1 1 8.5-8.5V23h10v-4.5a8.5 8.5 0 1 1 8.5 8.5H41v10zM23 18.5a4.5 4.5 0 1 0-4.5 4.5H23zM18.5 41a4.5 4.5 0 1 0 4.5 4.5V41zM37 27H27v10h10zm8.5-4a4.5 4.5 0 1 0-4.5-4.5V23zM41 45.5a4.5 4.5 0 1 0 4.5-4.5H41z" /></Symbol>;
};

export function Mail(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.664 37.496a3 3 0 0 1-3.328 0l-21-14A3 3 0 0 1 8 21v-2a3 3 0 0 1 3-3h42a3 3 0 0 1 3 3v2a3 3 0 0 1-1.336 2.496zM56 45a3 3 0 0 1-3 3H11a3 3 0 0 1-3-3V27.22a1.205 1.205 0 0 1 1.874-1.002L29.227 39.16a5 5 0 0 0 5.546 0l19.352-12.942A1.205 1.205 0 0 1 56 27.22z" /></Symbol>;
};

export function Man(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M40.99 38h-1.096a1 1 0 0 1-.995-.9l-1.086-10.865-.62 4.333a16 16 0 0 0-.148 2.877l.904 23.517A1 1 0 0 1 36.95 58h-2.056a1 1 0 0 1-.995-.9L32 38l-1.9 19.1a1 1 0 0 1-.994.9H27.05a1 1 0 0 1-1-1.038l.905-23.517a16.01 16.01 0 0 0-.149-2.877l-.619-4.333L25.101 37.1a1 1 0 0 1-.995.9h-1.095a1 1 0 0 1-1-1V23a5 5 0 0 1 5-5h9.978a5 5 0 0 1 5 5v14a1 1 0 0 1-1 1zM32 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function MapPinAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M33.59 55.185a1.958 1.958 0 0 1-3.18 0C27.634 51.295 16 34.38 16 24a16 16 0 0 1 32 0c0 10.38-11.634 27.296-14.41 31.185zM32 14a10 10 0 1 0 10 10 10 10 0 0 0-10-10z" /></Symbol>;
};

export function Map(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M55 13.494v31.968a2 2 0 0 1-1.212 1.839L43.09 51.885A1.5 1.5 0 0 1 41 50.507v-31.97a2 2 0 0 1 1.212-1.838l10.697-4.584A1.5 1.5 0 0 1 55 13.494zM21.788 47.3L11.09 51.885A1.5 1.5 0 0 1 9 50.507v-31.97a2 2 0 0 1 1.212-1.838l10.697-4.584A1.5 1.5 0 0 1 23 13.494v31.968a2 2 0 0 1-1.212 1.839zm5.303-35.186l10.697 4.584A2 2 0 0 1 39 18.538v31.969a1.5 1.5 0 0 1-2.09 1.378l-10.698-4.584A2 2 0 0 1 25 45.462V13.494a1.5 1.5 0 0 1 2.09-1.38z" /></Symbol>;
};

export function Men(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54.989 38h-1.095a1 1 0 0 1-.995-.9l-1.086-10.865-.62 4.333a15.991 15.991 0 0 0-.148 2.877l.904 23.517A1 1 0 0 1 50.95 58h-2.056a1 1 0 0 1-.995-.9L46 38l-1.899 19.1a1 1 0 0 1-.995.9H41.05a1 1 0 0 1-1-1.038l.905-23.517a16 16 0 0 0-.149-2.877l-.619-4.333L39.101 37.1a1 1 0 0 1-.995.9h-1.095a1 1 0 0 1-1-1V23a5 5 0 0 1 5-5h9.978a5 5 0 0 1 5 5v14a1 1 0 0 1-1 1zM46 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5zM26.989 38h-1.095a1 1 0 0 1-.995-.9l-1.087-10.865-.618 4.333a15.982 15.982 0 0 0-.15 2.877l.905 23.517A1 1 0 0 1 22.95 58h-2.056a1 1 0 0 1-.995-.9L18 38l-1.899 19.1a1 1 0 0 1-.995.9H13.05a1 1 0 0 1-1-1.038l.905-23.517a16.01 16.01 0 0 0-.149-2.877l-.619-4.333L11.101 37.1a1 1 0 0 1-.995.9H9.011a1 1 0 0 1-1-1V23a5 5 0 0 1 5-5h9.978a5 5 0 0 1 5 5v14a1 1 0 0 1-1 1zM18 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function Mathematics(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48 54H34a1 1 0 0 1-1-1V34a1 1 0 0 1 1-1h19a1 1 0 0 1 1 1v14a6 6 0 0 1-6 6zm0-15H38a1 1 0 1 0 0 2h10a1 1 0 1 0 0-2zm0 6H38a1 1 0 1 0 0 2h10a1 1 0 1 0 0-2zM34 10h14a6 6 0 0 1 6 6v14a1 1 0 0 1-1 1H34a1 1 0 0 1-1-1V11a1 1 0 0 1 1-1zm4 12h10a1 1 0 0 0 0-2H38a1 1 0 0 0 0 2zM10 48V34a1 1 0 0 1 1-1h19a1 1 0 0 1 1 1v19a1 1 0 0 1-1 1H16a6 6 0 0 1-6-6zm6.303-.293a.999.999 0 0 0 1.414 0L21 44.424l3.283 3.283a1 1 0 1 0 1.414-1.413l-3.284-3.284 3.3-3.3a1 1 0 1 0-1.413-1.413l-3.3 3.3-3.3-3.3a1 1 0 1 0-1.414 1.413l3.3 3.3-3.283 3.283a.999.999 0 0 0 0 1.414zM10 16a6 6 0 0 1 6-6h14a1 1 0 0 1 1 1v19a1 1 0 0 1-1 1H11a1 1 0 0 1-1-1zm6 6h4v4a1 1 0 0 0 2 0v-4h4a1 1 0 0 0 0-2h-4v-4a1 1 0 0 0-2 0v4h-4a1 1 0 0 0 0 2z" /></Symbol>;
};

export function Menu(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M15 16h34a2 2 0 1 1 0 4H15a2 2 0 1 1 0-4zm0 14h34a2 2 0 0 1 0 4H15a2 2 0 1 1 0-4zm0 14h34a2 2 0 0 1 0 4H15a2 2 0 1 1 0-4z" /></Symbol>;
};

export function Merge(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M35.757 31.93l5.314 5.313c1.889 1.888 4.929 6.4 4.929 9.07V52a2 2 0 0 1-4 0v-5.686a5.958 5.958 0 0 0-1.757-4.243l-7.314-7.314A10.126 10.126 0 0 1 32 33.67a10.08 10.08 0 0 1-.929 1.088l-7.314 7.314A5.959 5.959 0 0 0 22 46.314V52a2 2 0 0 1-4 0v-5.686c0-2.671 3.04-7.183 4.93-9.07l5.313-5.314A5.962 5.962 0 0 0 30 27.687V22.01h-4.995a1.027 1.027 0 0 1-.82-1.623l6.99-9.974a.996.996 0 0 1 1.638 0l6.988 9.974a1.027 1.027 0 0 1-.819 1.623H34v5.676a5.962 5.962 0 0 0 1.757 4.242z" /></Symbol>;
};

export function More(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M53 32a5 5 0 1 1-5-5 5 5 0 0 1 5 5zm-16 0a5 5 0 1 1-5-5 5 5 0 0 1 5 5zm-21-5a5 5 0 1 1-5 5 5 5 0 0 1 5-5z" /></Symbol>;
};

export function Move(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54.006 23a1 1 0 0 1-1 1H11.004a1 1 0 0 1-1-1v-9a2 2 0 0 1 2-2h11.343a4 4 0 0 1 2.83 1.172l3.656 3.656A4.002 4.002 0 0 0 32.662 18h19.344a2 2 0 0 1 2 2zM10.01 26h43.993a2 2 0 0 1 1.994 2.154L54.374 49.23A3 3 0 0 1 51.382 52H12.628a3 3 0 0 1-2.992-2.77l-.55-7.153A1 1 0 0 1 10.081 41H27.99v4.995a1.024 1.024 0 0 0 1.618.82l9.949-6.99a.997.997 0 0 0 0-1.638l-9.949-6.989a1.024 1.024 0 0 0-1.618.82V37H9.62a1 1 0 0 1-.997-.923l-.61-7.92A2.003 2.003 0 0 1 10.01 26z" /></Symbol>;
};

export function MoreAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 53a5 5 0 1 1 5-5 5 5 0 0 1-5 5zm0-32a5 5 0 1 1 5-5 5 5 0 0 1-5 5zm0 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function MoveDocument(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M37 22a1 1 0 0 1-1-1V10.414a1 1 0 0 1 1.707-.707l10.586 10.586A1 1 0 0 1 47.586 22zm13 5v26a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3V42a1 1 0 0 1 1-1h12.99v4.995a1.024 1.024 0 0 0 1.617.82l9.949-6.99a.997.997 0 0 0 0-1.638l-9.949-6.989a1.024 1.024 0 0 0-1.618.82V37H15a1 1 0 0 1-1-1V11a3 3 0 0 1 3-3h14a3 3 0 0 1 3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 1 3 3z" /></Symbol>;
};

export function Music(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M22.273 12.713l22.62-5.655A2.5 2.5 0 0 1 48 9.483V43.55C48 47.334 44.605 50 40 50a5.881 5.881 0 0 1-6-5.862 5.82 5.82 0 0 1 4.41-5.646l4.075-1.02A2 2 0 0 0 44 35.533V20.968a1 1 0 0 0-1.242-.97l-17.622 4.405A1.5 1.5 0 0 0 24 25.86v24.69C24 54.334 20.605 57 16 57a5.88 5.88 0 0 1-6-5.861 5.82 5.82 0 0 1 4.41-5.648l4.075-1.019A2 2 0 0 0 20 42.532V15.624a3 3 0 0 1 2.273-2.91z" /></Symbol>;
};

export function Navigation(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm10.913-12.797l-9.65-25.347a1.362 1.362 0 0 0-2.531 0l-9.65 25.347a1.34 1.34 0 0 0 2.034 1.542l8.113-5.48a1.378 1.378 0 0 1 1.537 0l8.113 5.48a1.34 1.34 0 0 0 2.034-1.541z" /></Symbol>;
};

export function NewFolder(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 20v3a1 1 0 0 1-1 1H11a1 1 0 0 1-1-1v-9a2 2 0 0 1 2-2h11.343a3.999 3.999 0 0 1 2.828 1.172l3.657 3.656A4 4 0 0 0 32.657 18H52a2 2 0 0 1 2 2zm-43.994 6h43.988a2 2 0 0 1 1.996 2.153L54.367 49.23A3.002 3.002 0 0 1 51.372 52H12.628a3.002 3.002 0 0 1-2.995-2.77L8.01 28.153A2.001 2.001 0 0 1 10.005 26zM23 40a1 1 0 0 0 1 1h6v6a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-6h6a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1h-6v-6a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v6h-6a1 1 0 0 0-1 1z" /></Symbol>;
};

export function NewDocument(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M37 22a1 1 0 0 1-1-1V10.414a1 1 0 0 1 1.707-.707l10.586 10.586A1 1 0 0 1 47.586 22zm13 5v26a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3V11a3 3 0 0 1 3-3h14a3 3 0 0 1 3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 1 3 3zm-9 11a1 1 0 0 0-1-1h-6v-6a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v6h-6a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h6v6a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-6h6a1 1 0 0 0 1-1z" /></Symbol>;
};

export function News(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M16 54a6 6 0 0 1-6-6V11c0-.552.895-1 2-1s2 .448 2 1 .895 1 2 1 2-.448 2-1 .895-1 2-1 2 .448 2 1 .895 1 2 1 2-.448 2-1 .895-1 2-1 2 .448 2 1 .895 1 2 1 2-.448 2-1 .895-1 2-1 2 .448 2 1 .895 1 2 1 2-.448 2-1 .895-1 2-1 2 .448 2 1 .895 1 2 1 2-.448 2-1 .895-1 2-1 2 .448 2 1v37a6 6 0 0 1-6 6zm1-22a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V33a1 1 0 0 0-1-1zm27-11a1 1 0 0 0-1-1H21a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h22a1 1 0 0 0 1-1zm3 11H33a1 1 0 0 0 0 2h14a1 1 0 1 0 0-2zm0 6H33a1 1 0 0 0 0 2h14a1 1 0 1 0 0-2zm0 6H33a1 1 0 0 0 0 2h14a1 1 0 1 0 0-2z" /></Symbol>;
};

export function Network(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 34h-8v6h3a3 3 0 0 1 3 3v8a3 3 0 0 1-3 3H39a3 3 0 0 1-3-3v-8a3 3 0 0 1 3-3h3v-6H22v6h3a3 3 0 0 1 3 3v8a3 3 0 0 1-3 3H15a3 3 0 0 1-3-3v-8a3 3 0 0 1 3-3h3v-6h-8a2 2 0 0 1 0-4h20v-6h-3a3 3 0 0 1-3-3v-8a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v8a3 3 0 0 1-3 3h-3v6h20a2 2 0 0 1 0 4z" /></Symbol>;
};

export function No(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zM16.51 19.352A19.996 19.996 0 0 0 44.648 47.49zM52 32a19.997 19.997 0 0 0-32.66-15.48l28.14 28.14A19.913 19.913 0 0 0 52 32z" /></Symbol>;
};

export function NoEntry(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm16-26a2 2 0 0 0-2-2H18a2 2 0 0 0-2 2v4a2 2 0 0 0 2 2h28a2 2 0 0 0 2-2z" /></Symbol>;
};

export function Octothorpe(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50.998 25.18A2 2 0 0 1 49.006 27h-7.519l-.908 10h7.332a2.043 2.043 0 0 1 2.097 2.18A2.067 2.067 0 0 1 47.91 41h-7.695L39.2 52.181A2 2 0 0 1 37.208 54a2 2 0 0 1-1.991-2.18L36.199 41H26.166L25.15 52.181a2 2 0 1 1-3.983-.362L22.15 41h-7.156a2 2 0 0 1 0-4h7.519l.908-10H16.09a2.043 2.043 0 0 1-2.097-2.181A2.066 2.066 0 0 1 16.09 23h7.696L24.8 11.82a2 2 0 0 1 3.984.36L27.8 23h10.033l1.016-11.18a2 2 0 0 1 3.984.36L41.85 23h7.156a2 2 0 0 1 1.992 2.18zM27.438 27l-.908 10h10.032l.909-10z" /></Symbol>;
};

export function Package(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52.997 43.215l-20 11.496a2 2 0 0 1-1.994 0L11 43.213a1.997 1.997 0 0 1-1-1.732V22.488a1.997 1.997 0 0 1 1-1.731l5.315-3.056a.997.997 0 0 1 .997.002L39.5 30.559a1.004 1.004 0 0 1 .5.869v5.057a.5.5 0 0 0 .755.431l4.756-2.834a1.004 1.004 0 0 0 .489-.863v-5.557a.502.502 0 0 0-.25-.435l-22.16-12.84a.503.503 0 0 1 .002-.87l7.41-4.26a2 2 0 0 1 1.995 0L53 20.758a1.998 1.998 0 0 1 1 1.731v18.99a2.003 2.003 0 0 1-1.003 1.736zM22 37.29a1 1 0 0 0-.502-.867l-6.749-3.88a.5.5 0 0 0-.749.434v6.765a1 1 0 0 0 .501.867l6.75 3.88a.5.5 0 0 0 .749-.433z" /></Symbol>;
};

export function PathAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54 46a2 2 0 0 1-2 2H24a2 2 0 1 1 0-4h28a2 2 0 0 1 2 2zM10 18a2 2 0 0 1 2-2h28a2 2 0 1 1 0 4H12a2 2 0 0 1-2-2zm6 14a2 2 0 0 1 2-2h28a2 2 0 0 1 2 2 2 2 0 0 1-2 2H18a2 2 0 0 1-2-2z" /></Symbol>;
};

export function Paste(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M21 18a1 1 0 0 1-1-1v-5a2 2 0 0 1 2-2h1.5a1 1 0 0 0 .8-.4l1.8-2.4A3 3 0 0 1 28.5 6h7a3 3 0 0 1 2.4 1.2l1.8 2.4a1 1 0 0 0 .8.4H42a2 2 0 0 1 2 2v5a1 1 0 0 1-1 1zm26-4h2a3 3 0 0 1 3 3v36a3 3 0 0 1-3 3H15a3 3 0 0 1-3-3V17a3 3 0 0 1 3-3h2a1 1 0 0 1 1 1v2a3 3 0 0 0 3 3h22a3 3 0 0 0 3-3v-2a1 1 0 0 1 1-1z" /></Symbol>;
};

export function Path(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56 46a2 2 0 0 1-2 2H32a2 2 0 1 1 0-4h22a2 2 0 0 1 2 2zM17 35a3 3 0 1 1 3-3 3 3 0 0 1-3 3zm1-17a2 2 0 0 1 2-2h22a2 2 0 1 1 0 4H20a2 2 0 0 1-2-2zm6 14a2 2 0 0 1 2-2h22a2 2 0 0 1 2 2 2 2 0 0 1-2 2H26a2 2 0 0 1-2-2zm-1 17a3 3 0 1 1 3-3 3 3 0 0 1-3 3zm-9-31a3 3 0 1 1-3-3 3 3 0 0 1 3 3z" /></Symbol>;
};

export function Pencil(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M10.042 53.043l2.018-6.738a8.003 8.003 0 0 1 2.003-3.359L39.71 17.259a.998.998 0 0 1 1.412 0l5.579 5.588a1.001 1.001 0 0 1 0 1.414L21.054 49.948a7.994 7.994 0 0 1-3.357 2.008l-6.722 2.02a.75.75 0 0 1-.933-.933zm40.313-42.189l2.746 2.75a2.993 2.993 0 0 1 0 4.228l-3.346 3.35a.994.994 0 0 1-1.408 0l-5.56-5.568a.998.998 0 0 1 0-1.41l3.346-3.35a2.982 2.982 0 0 1 4.222 0z" /></Symbol>;
};

export function Phone(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M41 56H23a5 5 0 0 1-5-5V13a5 5 0 0 1 5-5h18a5 5 0 0 1 5 5v38a5 5 0 0 1-5 5zm1-40a1 1 0 0 0-1-1H23a1 1 0 0 0-1 1v32a1 1 0 0 0 1 1h18a1 1 0 0 0 1-1z" /></Symbol>;
};

export function PhoneCall(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50.11 51.701a1.878 1.878 0 0 1-1.831 1.301c-3.608-.074-13.628-1.386-24.765-12.523-11.138-11.138-12.448-21.16-12.522-24.767a1.878 1.878 0 0 1 1.3-1.83l8.549-2.79a1.95 1.95 0 0 1 2.383 1.085l3.413 7.905a1.906 1.906 0 0 1-.39 2.111L22.24 26.2a.942.942 0 0 0-.22.998 23.338 23.338 0 0 0 5.769 9.008 23.341 23.341 0 0 0 9.006 5.768.944.944 0 0 0 1-.22l4.004-4.004a1.909 1.909 0 0 1 2.115-.391l7.903 3.412a1.948 1.948 0 0 1 1.084 2.38z" /></Symbol>;
};

export function PhotoLandscape(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56 45a6 6 0 0 1-6 6H14a6 6 0 0 1-6-6V19a6 6 0 0 1 6-6h36a6 6 0 0 1 6 6zm-4-26a2 2 0 0 0-2-2H14a2 2 0 0 0-2 2v24.159l12.578-12.573a2 2 0 0 1 2.829 0l10.588 10.586 4.587-4.586a2 2 0 0 1 2.829 0L52 43.173zM41 31a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function PlaybackFastForward(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M34.988 19.285l17.993 10.92a2.12 2.12 0 0 1 0 3.588l-17.993 10.92a1.997 1.997 0 0 1-3.001-1.793v-8.523L14.988 44.713a1.997 1.997 0 0 1-3.001-1.793V21.079a1.997 1.997 0 0 1 3.001-1.794l16.999 10.317v-8.523a1.997 1.997 0 0 1 3.001-1.794z" /></Symbol>;
};

export function PlaybackNext(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M12.988 19.285l16.999 10.317v-8.523a1.997 1.997 0 0 1 3.001-1.794L50 29.61V21a2 2 0 0 1 4 0v22a2 2 0 0 1-4 0v-8.611L32.988 44.713a1.997 1.997 0 0 1-3.001-1.793v-8.523L12.988 44.713a1.997 1.997 0 0 1-3.001-1.793V21.079a1.997 1.997 0 0 1 3.001-1.794z" /></Symbol>;
};

export function PlaybackPrevious(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51.012 44.715L34.013 34.398v8.524a1.997 1.997 0 0 1-3.001 1.793L14 34.39V43a2 2 0 0 1-4 0V21a2 2 0 0 1 4 0v8.611l17.012-10.324a1.997 1.997 0 0 1 3.001 1.794v8.523l16.999-10.317a1.997 1.997 0 0 1 3.001 1.794v21.84a1.997 1.997 0 0 1-3.001 1.794z" /></Symbol>;
};

export function PlaybackPause(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M25 50h-6a3 3 0 0 1-3-3V17a3 3 0 0 1 3-3h6a3 3 0 0 1 3 3v30a3 3 0 0 1-3 3zm14 0a3 3 0 0 1-3-3V17a3 3 0 0 1 3-3h6a3 3 0 0 1 3 3v30a3 3 0 0 1-3 3z" /></Symbol>;
};

export function PlaybackPlay(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M18.88 12.198l35.997 17.996a2 2 0 0 1 0 3.577L18.881 51.768a2 2 0 0 1-2.894-1.788V13.986a2 2 0 0 1 2.894-1.788z" /></Symbol>;
};

export function PlaybackStop(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M47 14a3 3 0 0 1 3 3v30a3 3 0 0 1-3 3H17a3 3 0 0 1-3-3V17a3 3 0 0 1 3-3z" /></Symbol>;
};

export function PlaybackPlayButton(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm11.866-25.8l-17.984-9.01a2 2 0 0 0-2.892 1.792v18.017a2 2 0 0 0 2.892 1.79l17.984-9.008a2.006 2.006 0 0 0 0-3.582z" /></Symbol>;
};

export function PlaybackRewind(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M29.012 44.715l-17.994-10.92a2.12 2.12 0 0 1 0-3.587l17.994-10.92a1.997 1.997 0 0 1 3.001 1.794v8.523l16.999-10.318a1.997 1.997 0 0 1 3.001 1.794v21.84a1.997 1.997 0 0 1-3.001 1.794L32.013 34.398v8.524a1.997 1.997 0 0 1-3.001 1.793z" /></Symbol>;
};

export function PopOut(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51.998 26.976a2 2 0 0 1-2.002-1.998v-8.142L33.372 33.414a1.99 1.99 0 0 1-2.804-2.823L47.21 13.996h-8.232a1.998 1.998 0 1 1 0-3.996h13.02A2 2 0 0 1 54 11.998v12.98a2 2 0 0 1-2.002 1.998zM28 21.993H16a2.002 2.002 0 0 0-2 2V48a2.002 2.002 0 0 0 2 2h24a2.003 2.003 0 0 0 2-2V36a2 2 0 1 1 4 0v12a6.008 6.008 0 0 1-6 6.002H16A6.008 6.008 0 0 1 10 48V23.994a6.008 6.008 0 0 1 6-6.002h12a2 2 0 0 1 0 4.001z" /></Symbol>;
};

export function Postcard(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path d="M53,49H11a3,3,0,0,1-3-3V18a3,3,0,0,1,3-3H53a3,3,0,0,1,3,3V46A3,3,0,0,1,53,49ZM13,39h7a1,1,0,1,0,0-2H13a1,1,0,1,0,0,2ZM24,25H13a1,1,0,1,0,0,2H24a1,1,0,1,0,0-2Zm2,6H13a1,1,0,1,0,0,2H26a1,1,0,0,0,0-2Zm7-11a1,1,0,0,0-2,0V44a1,1,0,0,0,2,0Zm19,.5A1.5,1.5,0,0,0,50.5,19h-9A1.5,1.5,0,0,0,40,20.5v11A1.5,1.5,0,0,0,41.5,33h9A1.5,1.5,0,0,0,52,31.5ZM49.5,31h-7a.5.5,0,0,1-.5-.5v-9a.5.5,0,0,1,.5-.5h7a.5.5,0,0,1,.5.5v9A.5.5,0,0,1,49.5,31Z" fillRule="evenodd" /></Symbol>;
};

export function Preview(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M42 50a2 2 0 1 1 0 4H22a2 2 0 0 1 0-4zm14-6a4 4 0 0 1-4 4H12a4 4 0 0 1-4-4V18a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4zM42.447 30.105l-16-8A1 1 0 0 0 25 23v16a1 1 0 0 0 1.447.894l16-8a1 1 0 0 0 0-1.789z" /></Symbol>;
};

export function Private(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M53.58 44.172l-8.73 9.132A8.727 8.727 0 0 1 38.542 56H26.016a8.728 8.728 0 0 1-6.16-2.545L10.568 44.2a8.726 8.726 0 0 1-2.567-6.182V25.51a8.726 8.726 0 0 1 2.695-6.307l9.184-8.783A8.728 8.728 0 0 1 25.912 8h12.735a8.728 8.728 0 0 1 6.183 2.568l8.625 8.659A8.728 8.728 0 0 1 56 25.386V38.14a8.727 8.727 0 0 1-2.42 6.031zm-5.676-13.477A2.207 2.207 0 0 0 46 29.999c-2.532 0-5.146 4.432-5.83 5.675A.625.625 0 0 1 39 35.37V18.5a1.5 1.5 0 0 0-3 0V31a1 1 0 1 1-2 0V16.5a1.5 1.5 0 1 0-3 0V31a1 1 0 1 1-2 0V18.5a1.5 1.5 0 0 0-3 0V32a1 1 0 1 1-2 0v-8.5a1.5 1.5 0 1 0-3 0l-.003 15.493A9.998 9.998 0 0 0 31 48.987c5.544 0 9.902-2.449 12.002-8.994a32.65 32.65 0 0 1 4.903-8.714.498.498 0 0 0 0-.584z" /></Symbol>;
};

export function Promoted(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 52H14a6 6 0 0 1-6-6V18a6 6 0 0 1 6-6h36a6 6 0 0 1 6 6v28a6 6 0 0 1-6 6zm-6-30.002A2 2 0 0 0 41.998 20h-13.02a1.998 1.998 0 1 0 0 3.996h8.232L20.567 40.592a1.99 1.99 0 0 0 2.805 2.822l16.624-16.578v8.142a2.002 2.002 0 0 0 4.004 0z" /></Symbol>;
};

export function Public(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm3.5-41a3.5 3.5 0 1 0 3.5 3.5 3.5 3.5 0 0 0-3.5-3.5zm7.605 14.837l-4.374-.795a.5.5 0 0 1-.346-.247l-1.534-2.727a4 4 0 0 0-1.698-1.616l-1.715-.858a5.625 5.625 0 0 0-3.841-.436l-5.304 1.286A3 3 0 0 0 22 27.36v4.572a1.068 1.068 0 0 0 2.126.145l.585-4.29a.37.37 0 0 1 .328-.317l4.238-.446a.125.125 0 0 1 .125.178l-2.342 4.85a16.003 16.003 0 0 0-1.418 4.612L25.03 40.8a.752.752 0 0 1-.156.359l-3.545 4.43a1.498 1.498 0 0 0-.329.938A1.474 1.474 0 0 0 22.474 48a1.61 1.61 0 0 0 1.22-.466l3.67-3.853a8.007 8.007 0 0 0 1.436-2.093l1.633-3.447a.125.125 0 0 1 .19-.045l4.215 3.278a.5.5 0 0 1 .189.329l.787 5.973A1.581 1.581 0 0 0 37.406 49a1.563 1.563 0 0 0 1.604-1.604l-.429-6.556a4 4 0 0 0-1.232-2.689l-4.285-4.09a.125.125 0 0 1-.028-.14l1.88-4.233a.125.125 0 0 1 .212-.027l.821 1.026A3.5 3.5 0 0 0 38.682 32h4.228a1.09 1.09 0 0 0 .195-2.163z" /></Symbol>;
};

export function Redirect(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M55.3 33.52L37.275 48.518a2.003 2.003 0 0 1-3.285-1.536V39h-8.517l2.429 7.73A1.753 1.753 0 0 1 26.213 49H9a.994.994 0 0 1-.453-1.884l8.458-4.23L12.1 27.27A1.753 1.753 0 0 1 13.787 25h20.202v-8.015a2.003 2.003 0 0 1 3.285-1.536l18.027 14.998a1.998 1.998 0 0 1 0 3.073z" /></Symbol>;
};

export function Redo(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M30.014 48h16.013a2 2 0 1 1 0 4H30.014a18 18 0 1 1 0-36h9.997v-4.982a1.026 1.026 0 0 1 1.62-.82l9.956 6.99a.996.996 0 0 1 0 1.637l-9.956 6.99a1.026 1.026 0 0 1-1.62-.82V20h-9.997a14 14 0 1 0 0 28z" /></Symbol>;
};

export function Reference(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 36a4 4 0 1 1 4-4 4 4 0 0 1-4 4zm-40-4a4 4 0 1 1 4 4 4 4 0 0 1-4-4zm18-18a4 4 0 1 1 4 4 4 4 0 0 1-4-4zm8 36a4 4 0 1 1-4-4 4 4 0 0 1 4 4zm-21.446-.596a2.001 2.001 0 0 1 0-2.83l14.589-14.59-14.587-14.586a2.002 2.002 0 1 1 2.83-2.83l14.587 14.586 14.598-14.598a2.002 2.002 0 0 1 2.83 2.831L34.804 31.985l14.6 14.6a2.002 2.002 0 1 1-2.83 2.83l-14.6-14.6-14.59 14.589a2.001 2.001 0 0 1-2.83 0z" /></Symbol>;
};

export function Rename(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 20a6 6 0 0 1 6 6v12a6 6 0 0 1-6 6H40a2 2 0 1 1 0-4h10a2 2 0 0 0 2-2V26a2 2 0 0 0-2-2H40a2 2 0 1 1 0-4zM26 10h1a6.976 6.976 0 0 1 5 2.11A6.976 6.976 0 0 1 37 10h1a2 2 0 0 1 0 4h-1a3.003 3.003 0 0 0-3 3v30a3.003 3.003 0 0 0 3 3h1a2 2 0 1 1 0 4h-1a6.976 6.976 0 0 1-5-2.11A6.976 6.976 0 0 1 27 54h-1a2 2 0 1 1 0-4h1a3.003 3.003 0 0 0 3-3V17a3.003 3.003 0 0 0-3-3h-1a2 2 0 0 1 0-4zM14 40h10a2 2 0 1 1 0 4H14a6 6 0 0 1-6-6V26a6 6 0 0 1 6-6h10a2 2 0 1 1 0 4H14a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2z" /></Symbol>;
};

export function Repeat(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M42 44H22.01v4.995a1.024 1.024 0 0 1-1.618.82l-9.948-6.99a.997.997 0 0 1 0-1.638l9.948-6.989a1.024 1.024 0 0 1 1.619.82V40H42a8 8 0 0 0 8-8 2 2 0 1 1 4 0 12 12 0 0 1-12 12zm1.607-14.199a1.024 1.024 0 0 1-1.618-.819V24H22a8 8 0 0 0-8 8 2 2 0 0 1-4 0 12 12 0 0 1 12-12h19.99v-4.995a1.024 1.024 0 0 1 1.617-.82l9.949 6.99a.997.997 0 0 1 0 1.638z" /></Symbol>;
};

export function Reply(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52.309 51a1.671 1.671 0 0 1-1.602-1.165C48.454 42.532 42.212 39.003 32.01 39v8.011a2.003 2.003 0 0 1-3.285 1.538L10.699 33.537a2 2 0 0 1 0-3.076L28.726 15.45a2.003 2.003 0 0 1 3.285 1.538v8.014c14.274.727 21.465 9.365 21.987 24.225A1.7 1.7 0 0 1 52.308 51z" /></Symbol>;
};

export function ReplyAll(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M60.309 51a1.671 1.671 0 0 1-1.602-1.165C56.453 42.532 50.212 39.003 40.01 39v8.011a2.003 2.003 0 0 1-3.285 1.538L18.699 33.537a2 2 0 0 1 0-3.076L36.726 15.45a2.003 2.003 0 0 1 3.285 1.538v8.014c14.274.727 21.465 9.365 21.987 24.225A1.7 1.7 0 0 1 60.308 51zM17.41 28.93a4.002 4.002 0 0 0 0 6.148l5.15 4.29a4.005 4.005 0 0 1 1.441 3.077v4.568a2.003 2.003 0 0 1-3.284 1.54L2.693 33.542a2.001 2.001 0 0 1 0-3.075l18.023-15.012a2.003 2.003 0 0 1 3.285 1.542v4.567a4.005 4.005 0 0 1-1.441 3.077z" /></Symbol>;
};

export function Report(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm0-44a20 20 0 1 0 20 20 20 20 0 0 0-20-20zm-.006 34a3 3 0 1 1 3-3 3 3 0 0 1-3 3zm0-9a2.404 2.404 0 0 1-2.446-2.248l-.554-13.755a3 3 0 0 1 6 0l-.555 13.755A2.404 2.404 0 0 1 31.994 37z" /></Symbol>;
};

export function Repost(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51.595 49.181a2.002 2.002 0 0 1-3.201 0l-9.004-11.99a1.998 1.998 0 0 1 1.601-3.196h5.01v-9.991a2 2 0 0 0-2.002-1.999H27.986a2 2 0 0 1-2.002-1.998v-3.996a2 2 0 0 1 2.002-1.998H45a9 9 0 0 1 9.008 8.992v10.99h4.989a1.998 1.998 0 0 1 1.601 3.196zm-36.01-34.37l9.002 11.99a1.998 1.998 0 0 1-1.602 3.197h-5.008v9.991a2 2 0 0 0 2.002 1.998h16.014a2 2 0 0 1 2.002 1.999v3.996a2 2 0 0 1-2.002 1.998H18.978a9 9 0 0 1-9.007-8.992v-10.99H4.983a1.998 1.998 0 0 1-1.602-3.197l9.002-11.99a2.004 2.004 0 0 1 3.202 0z" /></Symbol>;
};

export function RotateClockwise(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32.01 54.001v4.994a1.028 1.028 0 0 1-1.622.82l-9.974-6.99a.996.996 0 0 1 0-1.638l9.974-6.988a1.027 1.027 0 0 1 1.623.819V50A18.002 18.002 0 0 0 45.96 20.636a2 2 0 0 1 3.05-2.589 22.002 22.002 0 0 1-17 35.954zm-3.26-30.929l9.85 2.094a3.02 3.02 0 0 1 2.327 3.583L38.833 38.6a3.021 3.021 0 0 1-3.583 2.327l-9.85-2.094a3.021 3.021 0 0 1-2.327-3.583l2.094-9.85a3.021 3.021 0 0 1 3.583-2.328zm4.862-3.27a1.027 1.027 0 0 1-1.623-.82V14A18.002 18.002 0 0 0 18.04 43.365a2 2 0 0 1-3.05 2.588A22.002 22.002 0 0 1 31.99 10V5.005a1.027 1.027 0 0 1 1.623-.82l9.974 6.99a.996.996 0 0 1 0 1.638z" /></Symbol>;
};

export function RotateCounter(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46.069 46.07a2.013 2.013 0 0 1-.109-2.705A18.002 18.002 0 0 0 32.01 14v4.983a1.027 1.027 0 0 1-1.622.819l-9.974-6.989a.996.996 0 0 1 0-1.638l9.974-6.989a1.027 1.027 0 0 1 1.623.82v4.993a22.002 22.002 0 0 1 17 35.954 1.994 1.994 0 0 1-2.943.117zM28.75 40.927a3.021 3.021 0 0 1-3.583-2.327l-2.094-9.85a3.021 3.021 0 0 1 2.327-3.584l9.85-2.094a3.021 3.021 0 0 1 3.583 2.327l2.094 9.851a3.021 3.021 0 0 1-2.327 3.583zM17.931 17.931a2.013 2.013 0 0 1 .109 2.705A18.002 18.002 0 0 0 31.99 50v-4.983a1.027 1.027 0 0 1 1.622-.82l9.974 6.99a.996.996 0 0 1 0 1.637l-9.974 6.99a1.027 1.027 0 0 1-1.622-.82v-4.994a22.002 22.002 0 0 1-17-35.954 1.994 1.994 0 0 1 2.941-.116z" /></Symbol>;
};

export function Safari(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm12.102-37.807l-14.724 9.828a4.85 4.85 0 0 0-1.345 1.346l-9.816 14.741a1.213 1.213 0 0 0 1.681 1.684l14.724-9.828a4.853 4.853 0 0 0 1.345-1.346l9.816-14.741a1.213 1.213 0 0 0-1.681-1.684zM32 35a3 3 0 1 1 3-3 3 3 0 0 1-3 3z" /></Symbol>;
};

export function Save(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46 54.003H18A6.008 6.008 0 0 1 12 48V44a2 2 0 0 1 4 0v4a2.002 2.002 0 0 0 2 2.001h28a2.003 2.003 0 0 0 2-2V44a2 2 0 0 1 4 0v4a6.008 6.008 0 0 1-6 6.003zM32 8a2 2 0 0 1 2 2v19.99h4.995a1.027 1.027 0 0 1 .82 1.622l-6.99 9.974a.996.996 0 0 1-1.638 0L24.2 31.612a1.027 1.027 0 0 1 .818-1.623H30V10a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Search(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M29 10a18.974 18.974 0 0 1 15.248 30.302l10.257 8.675a3.92 3.92 0 1 1-5.528 5.528L40.3 44.249A18.991 18.991 0 1 1 29 10zm0 34a15 15 0 1 0-15-15 15 15 0 0 0 15 15z" /></Symbol>;
};

export function Send(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50.08 8.187a1.305 1.305 0 0 1 1.935 1.376l-7.511 39.143a1.576 1.576 0 0 1-2.068 1.183l-11.688-4.376-4.982 9.962a.927.927 0 0 1-1.756-.414V44.403a4.144 4.144 0 0 1 .888-2.513C25.547 41.15 44 17 44 17S19.283 39.087 18.311 39.874a3.043 3.043 0 0 1-2.923.561L4.881 36.842a1.298 1.298 0 0 1-.228-2.353z" /></Symbol>;
};

export function Server(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 54H12a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2zm-30-7a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm30-9H12a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2zm-30-7a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1zm30-9H12a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2zm-30-7a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1z" /></Symbol>;
};

export function SendTo(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M35.274 48.549a2.003 2.003 0 0 1-3.285-1.538V39c-10.202.003-16.443 3.532-18.696 10.835A1.671 1.671 0 0 1 11.691 51a1.7 1.7 0 0 1-1.689-1.774C10.524 34.366 17.715 25.728 31.99 25v-8.014a2.003 2.003 0 0 1 3.285-1.538l18.027 15.012a2 2 0 0 1 0 3.076z" /></Symbol>;
};

export function Shield(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M14 14a27.913 27.913 0 0 0 16.709-5.527 2 2 0 0 1 2.582 0A28.183 28.183 0 0 0 50 14a2 2 0 0 1 2 2c0 22.06-5.933 34.35-19.236 39.849a2.005 2.005 0 0 1-1.528 0C17.933 50.35 12 38.06 12 16a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Share(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 29.994v20.007a6.008 6.008 0 0 1-6 6.002H20A6.008 6.008 0 0 1 14 50V29.994a6.008 6.008 0 0 1 6-6.002h4a2 2 0 0 1 0 4.002h-4a2.002 2.002 0 0 0-2 2v20.007a2.002 2.002 0 0 0 2 2h24a2.003 2.003 0 0 0 2-2V29.994a2.003 2.003 0 0 0-2-2h-4a2 2 0 0 1 0-4.002h4a6.008 6.008 0 0 1 6 6.002zM32 40a2 2 0 0 1-2-2V18.01h-4.995a1.027 1.027 0 0 1-.82-1.622l6.99-9.974a.996.996 0 0 1 1.638 0L39.8 16.39a1.027 1.027 0 0 1-.819 1.622H34V38a2 2 0 0 1-2 2z" /></Symbol>;
};

export function Shrink(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M12 10h20a2 2 0 1 1 0 4H12a2 2 0 1 1 0-4zm0 8h6a2 2 0 1 1 0 4h-6a2 2 0 1 1 0-4zm-2 10a2 2 0 0 1 2-2h12a2 2 0 0 1 0 4H12a2 2 0 0 1-2-2zm36 26H18a2 2 0 0 1 0-4h28a2 2 0 0 1 0 4zm-6.186-18.389l-6.989 9.975a.996.996 0 0 1-1.638 0l-6.989-9.974a1.027 1.027 0 0 1 .82-1.623h13.977a1.027 1.027 0 0 1 .82 1.623zM52 30H32a2 2 0 0 1 0-4h20a2 2 0 0 1 0 4zm0-8h-6a2 2 0 1 1 0-4h6a2 2 0 1 1 0 4zm2-10a2 2 0 0 1-2 2H40a2 2 0 1 1 0-4h12a2 2 0 0 1 2 2zm-14 8a2 2 0 0 1-2 2H26a2 2 0 1 1 0-4h12a2 2 0 0 1 2 2z" /></Symbol>;
};

export function Shuffle(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45.608 49.801a1.024 1.024 0 0 1-1.619-.819v-4.981h-4.163a9.985 9.985 0 0 1-7.797-3.753l-1.596-2.003a2.01 2.01 0 0 1 0-2.51 1.993 1.993 0 0 1 3.115 0l1.6 2.008a6.005 6.005 0 0 0 4.69 2.258h4.152v-4.996a1.024 1.024 0 0 1 1.618-.82l9.948 6.99a.997.997 0 0 1 0 1.638zm0-20a1.024 1.024 0 0 1-1.619-.819V24h-4.152a5.971 5.971 0 0 0-4.686 2.252L23.956 40.248A9.953 9.953 0 0 1 16.147 44H9.993a2 2 0 0 1 0-4h6.154a5.97 5.97 0 0 0 4.686-2.251l11.196-13.998A9.95 9.95 0 0 1 39.837 20h4.152v-4.994a1.024 1.024 0 0 1 1.619-.82l9.948 6.99a.997.997 0 0 1 0 1.638zm-21.645-6.04l1.59 1.994a2 2 0 0 1-3.115 2.51l-1.605-2.015a5.99 5.99 0 0 0-4.677-2.251h-6.06a2.072 2.072 0 0 1-2.076-1.665 2.002 2.002 0 0 1 1.973-2.336h6.154a10.01 10.01 0 0 1 7.816 3.763z" /></Symbol>;
};

export function SidebarLeftOpen(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 52H12a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v32a4 4 0 0 1-4 4zM22 17a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm23.586 14.175l-9.974-6.99a1.028 1.028 0 0 0-1.623.82v13.978a1.027 1.027 0 0 0 1.623.818l9.974-6.988a.996.996 0 0 0 0-1.638z" /></Symbol>;
};

export function SidebarRight(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M12 52h40a4 4 0 0 0 4-4V16a4 4 0 0 0-4-4H12a4 4 0 0 0-4 4v32a4 4 0 0 0 4 4zm30-35a1 1 0 0 1 1-1h8a1 1 0 0 1 1 1v30a1 1 0 0 1-1 1h-8a1 1 0 0 1-1-1zm-30 0a1 1 0 0 1 1-1h24a1 1 0 0 1 1 1v30a1 1 0 0 1-1 1H13a1 1 0 0 1-1-1z" /></Symbol>;
};

export function SidebarLeftClose(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 52H12a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v32a4 4 0 0 1-4 4zM22 17a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm22.011 8.018a1.024 1.024 0 0 0-1.618-.82l-9.949 6.99a.997.997 0 0 0 0 1.637l9.949 6.99a1.025 1.025 0 0 0 1.618-.82z" /></Symbol>;
};

export function SidebarLeft(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 52H12a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v32a4 4 0 0 1-4 4zM22 17a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1zm30 0a1 1 0 0 0-1-1H27a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h24a1 1 0 0 0 1-1z" /></Symbol>;
};

export function SidebarRightClose(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 52H12a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v32a4 4 0 0 1-4 4zM31.556 31.175l-9.948-6.99a1.024 1.024 0 0 0-1.619.82v13.978a1.024 1.024 0 0 0 1.618.818l9.949-6.988a.997.997 0 0 0 0-1.638zM52 17a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1z" /></Symbol>;
};

export function SignIn(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M26.01 52h17.994a6.006 6.006 0 0 0 5.999-6V18a6.006 6.006 0 0 0-5.998-6H26.01a6.006 6.006 0 0 0-5.998 6v6a2 2 0 1 0 3.998 0v-6a2.002 2.002 0 0 1 2-2h17.994a2.002 2.002 0 0 1 2 2v28a2.002 2.002 0 0 1-2 2H26.01a2.002 2.002 0 0 1-2-2v-6a2 2 0 1 0-3.999 0v6a6.006 6.006 0 0 0 5.999 6zM7.991 32a2 2 0 0 0 1.999 2h19.976v4.995a1.027 1.027 0 0 0 1.621.82l9.968-6.99a.996.996 0 0 0 0-1.638l-9.968-6.989a1.027 1.027 0 0 0-1.621.82V30H9.99a2 2 0 0 0-1.999 2z" /></Symbol>;
};

export function SidebarRightOpen(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52 52H12a4 4 0 0 1-4-4V16a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v32a4 4 0 0 1-4 4zM30.01 25.017a1.024 1.024 0 0 0-1.618-.818l-9.948 6.988a.997.997 0 0 0 0 1.638l9.948 6.99a1.024 1.024 0 0 0 1.619-.82zM52 17a1 1 0 0 0-1-1h-8a1 1 0 0 0-1 1v30a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1z" /></Symbol>;
};

export function SignOut(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M35.99 12H17.996a6.006 6.006 0 0 0-5.998 6v28a6.006 6.006 0 0 0 5.998 6H35.99a6.006 6.006 0 0 0 5.998-6v-6a2 2 0 1 0-3.999 0v6a2.002 2.002 0 0 1-1.999 2H17.996a2.002 2.002 0 0 1-2-2V18a2.002 2.002 0 0 1 2-2H35.99a2.002 2.002 0 0 1 2 2v6a2 2 0 1 0 3.999 0v-6a6.006 6.006 0 0 0-5.999-6zm-12 20a2 2 0 0 0 2 2h19.976v4.995a1.027 1.027 0 0 0 1.621.82l9.968-6.99a.996.996 0 0 0 0-1.638l-9.968-6.989a1.027 1.027 0 0 0-1.621.82V30H25.99a2 2 0 0 0-2 2z" /></Symbol>;
};

export function Slide(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M49 10a5 5 0 0 1 5 5v34a5 5 0 0 1-5 5H15a5 5 0 0 1-5-5V15a5 5 0 0 1 5-5zM14 21v22a1 1 0 0 0 1 1h34a1 1 0 0 0 1-1V21a1 1 0 0 0-1-1H15a1 1 0 0 0-1 1z" /></Symbol>;
};

export function SlidersVertical(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M16 10a2 2 0 0 1 2 2v24.367a5.966 5.966 0 0 1 0 11.266V52a2 2 0 0 1-4 0v-4.367a5.966 5.966 0 0 1 0-11.266V12a2 2 0 0 1 2-2zm16 0a2 2 0 0 1 2 2v4.367a5.966 5.966 0 0 1 0 11.266V52a2 2 0 1 1-4 0V27.633a5.966 5.966 0 0 1 0-11.266V12a2 2 0 0 1 2-2zm16 0a2 2 0 0 1 2 2v14.367a5.966 5.966 0 0 1 0 11.266V52a2 2 0 0 1-4 0V37.633a5.966 5.966 0 0 1 0-11.266V12a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Sort(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M20.813 10.414l6.989 9.974a1.027 1.027 0 0 1-.82 1.623H22V52a2 2 0 0 1-4 0V22.01h-4.995a1.027 1.027 0 0 1-.82-1.622l6.99-9.974a.996.996 0 0 1 1.638 0zm31.001 33.198l-6.989 9.974a.996.996 0 0 1-1.638 0L36.2 43.612a1.027 1.027 0 0 1 .819-1.623H42V12a2 2 0 0 1 4 0v29.99h4.995a1.027 1.027 0 0 1 .82 1.622z" /></Symbol>;
};

export function StarEmpty(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M45.608 36.446a.5.5 0 0 0-.186.557l4.361 13.88a3.892 3.892 0 0 1-6 4.34l-11.657-8.785a.25.25 0 0 0-.301 0l-11.618 8.77a3.89 3.89 0 0 1-6-4.36l4.378-14a.25.25 0 0 0-.093-.277l-11.85-8.51a3.915 3.915 0 0 1 2.434-7.09h14.407a.25.25 0 0 0 .238-.173l4.56-14.102a3.876 3.876 0 0 1 7.39.005l4.557 14.097a.25.25 0 0 0 .238.173h14.46a3.915 3.915 0 0 1 2.43 7.092zm8.417-11.475H38.1a1 1 0 0 1-.952-.692l-4.936-15.27a.25.25 0 0 0-.475 0L26.8 24.279a1 1 0 0 1-.951.692H9.972a.25.25 0 0 0-.146.453l12.905 9.268a1 1 0 0 1 .371 1.11l-4.764 15.237a.25.25 0 0 0 .39.274l12.644-9.544a1 1 0 0 1 1.204 0l12.678 9.555a.25.25 0 0 0 .389-.275l-4.799-15.27a1 1 0 0 1 .374-1.113l12.952-9.242a.25.25 0 0 0-.145-.453z" /></Symbol>;
};

export function Star(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M42.986 35.876l4.906 15.614a1.89 1.89 0 0 1-2.902 2.134l-13-9.798-12.975 9.794a1.89 1.89 0 0 1-2.893-2.144l4.87-15.576L7.82 26.44a1.912 1.912 0 0 1 1.088-3.466h16.226L30.198 7.31a1.876 1.876 0 0 1 3.582 0l5.064 15.663h16.28a1.912 1.912 0 0 1 1.087 3.467z" /></Symbol>;
};

export function SwapHorizontally(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M43.607 25.802a1.024 1.024 0 0 1-1.618-.82V20H23a5 5 0 1 0 0 10h18a9 9 0 0 1 0 18H22.01v4.995a1.025 1.025 0 0 1-1.618.82l-9.948-6.99a.997.997 0 0 1 0-1.638l9.948-6.988a1.024 1.024 0 0 1 1.619.818V44H41a5 5 0 0 0 0-10H23a9 9 0 1 1 0-18h18.99v-4.995a1.024 1.024 0 0 1 1.617-.82l9.949 6.99a.997.997 0 0 1 0 1.638z" /></Symbol>;
};

export function SwapVertically(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46.825 53.556a.997.997 0 0 1-1.638 0L38.2 43.6a1.025 1.025 0 0 1 .819-1.62H44V22.979a5 5 0 1 0-10 0v18.013a9 9 0 1 1-18 0V21.988h-4.995a1.025 1.025 0 0 1-.82-1.62l6.99-9.955a.997.997 0 0 1 1.638 0l6.989 9.956a1.025 1.025 0 0 1-.82 1.62H20V40.99a5 5 0 1 0 10 0V22.978a9 9 0 1 1 18 0v19.003h4.995a1.025 1.025 0 0 1 .82 1.619z" /></Symbol>;
};

export function Sync(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52.825 43.586a.996.996 0 0 1-1.638 0L44.2 33.612a1.027 1.027 0 0 1 .819-1.623H50A18.002 18.002 0 0 0 20.636 18.04a2 2 0 0 1-2.589-3.05 22.003 22.003 0 0 1 35.955 17h4.993a1.027 1.027 0 0 1 .82 1.622zM12.813 20.414l6.988 9.974a1.027 1.027 0 0 1-.818 1.623H14A18.002 18.002 0 0 0 43.365 45.96a2 2 0 0 1 2.588 3.05A22.002 22.002 0 0 1 10 32.01H5.005a1.027 1.027 0 0 1-.82-1.622l6.99-9.974a.996.996 0 0 1 1.638 0z" /></Symbol>;
};

export function TV(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M8 44V18a4 4 0 0 1 4-4h40a4 4 0 0 1 4 4v26a4 4 0 0 1-4 4H12a4 4 0 0 1-4-4zm4-1a1 1 0 0 0 1 1h38a1 1 0 0 0 1-1V19a1 1 0 0 0-1-1H13a1 1 0 0 0-1 1zm8 9a2 2 0 0 1 2-2h20a2 2 0 1 1 0 4H22a2 2 0 0 1-2-2z" /></Symbol>;
};

export function Table(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51 52H13a5 5 0 0 1-5-5V17a5 5 0 0 1 5-5h38a5 5 0 0 1 5 5v30a5 5 0 0 1-5 5zM30 22H13a1 1 0 0 0-1 1v5h18zm0 10H12v6h18zm0 10H12v5a1 1 0 0 0 1 1h17zm22-19a1 1 0 0 0-1-1H34v6h18zm0 9H34v6h18zm0 10H34v6h17a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Tablet(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M15 51V13a5 5 0 0 1 5-5h24a5 5 0 0 1 5 5v38a5 5 0 0 1-5 5H20a5 5 0 0 1-5-5zm30-3V16a1 1 0 0 0-1-1H20a1 1 0 0 0-1 1v32a1 1 0 0 0 1 1h24a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Tabs(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50 44h-6v6a6 6 0 0 1-6 6H14a6 6 0 0 1-6-6V26a6 6 0 0 1 6-6h6v-6a6 6 0 0 1 6-6h24a6 6 0 0 1 6 6v24a6 6 0 0 1-6 6zM40 26a2 2 0 0 0-2-2H14a2 2 0 0 0-2 2v24a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2z" /></Symbol>;
};

export function Tag(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M16.51 11.994l13.573.004a6 6 0 0 1 4.243 1.76l20.77 20.785a3.009 3.009 0 0 1 0 4.253L38.793 55.115a3.004 3.004 0 0 1-4.25 0L13.769 34.327a6 6 0 0 1-1.756-4.24l-.004-13.592a4.5 4.5 0 0 1 4.501-4.501zm4.524 12.04a3.007 3.007 0 1 0-3.005-3.008 3.006 3.006 0 0 0 3.005 3.008z" /></Symbol>;
};

export function TextSize(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M11.623 23.988a1.507 1.507 0 0 1 1.417-.987h2.656a1.508 1.508 0 0 1 1.412.975l9.87 26.28a.55.55 0 0 1-.518.743h-2.876a1.131 1.131 0 0 1-1.06-.733L20.186 44H8.014l-2.368 6.347a1.005 1.005 0 0 1-.943.652H2.536a.55.55 0 0 1-.52-.738zM19.067 41L14.1 27.686 9.133 41zm17.494 1l-2.993 8.022a1.508 1.508 0 0 1-1.414.978h-2.4a.75.75 0 0 1-.707-1.01L42.263 14.31A2.01 2.01 0 0 1 44.15 13h2.69a2.01 2.01 0 0 1 1.887 1.309L61.944 49.99A.75.75 0 0 1 61.236 51H57.86a1.508 1.508 0 0 1-1.414-.978L53.455 42zm8.447-22.642L38.053 38h13.91z" /></Symbol>;
};

export function ThumbsUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M11.996 30h4.032a4.001 4.001 0 0 0 3.17-1.561l7.973-10.36A4 4 0 0 0 28 15.64V8a1.971 1.971 0 0 1 2.001-2c2.567 0 6.002 3 6.002 9a26.74 26.74 0 0 1-1.645 10.28.504.504 0 0 0 .447.72h13.038a4.113 4.113 0 0 1 4.119 3.392 3.978 3.978 0 0 1-2.056 4.108 3.971 3.971 0 0 1-.571 7.257 3.93 3.93 0 0 1 1.507 2.113 3.996 3.996 0 0 1-.32 2.883 3.6 3.6 0 0 1-3.104 2.186 3.942 3.942 0 0 1 .579 2.326A4.167 4.167 0 0 1 43.769 54H28.845a16.007 16.007 0 0 1-8.878-2.687l-.961-.641a4.002 4.002 0 0 0-2.22-.672h-4.79a2 2 0 0 1-2-2V32a2 2 0 0 1 2-2z" /></Symbol>;
};

export function ThumbsDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M11.996 14h4.79a4 4 0 0 0 2.22-.672l.961-.64A16.008 16.008 0 0 1 28.845 10h14.923a4.167 4.167 0 0 1 4.229 3.735 3.942 3.942 0 0 1-.58 2.326 3.601 3.601 0 0 1 3.105 2.186 3.996 3.996 0 0 1 .32 2.883 3.93 3.93 0 0 1-1.508 2.113 3.971 3.971 0 0 1 .572 7.257 3.978 3.978 0 0 1 2.056 4.109A4.112 4.112 0 0 1 47.843 38H34.804a.504.504 0 0 0-.446.72A26.74 26.74 0 0 1 36.003 49c0 6-3.436 9-6.002 9A1.971 1.971 0 0 1 28 56v-7.64a3.998 3.998 0 0 0-.83-2.439l-7.972-10.36A4.002 4.002 0 0 0 16.028 34h-4.032a2 2 0 0 1-2-2V16a2 2 0 0 1 2-2z" /></Symbol>;
};

export function Token(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M32 56a24 24 0 1 1 24-24 24 24 0 0 1-24 24zm3.962-29.008l-2.695-8.07a1.345 1.345 0 0 0-2.536 0l-2.696 8.07h-8.698a1.294 1.294 0 0 0-.78 2.353l7.028 4.939-2.503 7.97a1.314 1.314 0 0 0 2.008 1.499L32 38.78l6.907 4.984a1.314 1.314 0 0 0 2.01-1.496l-2.503-7.987 7.026-4.935a1.295 1.295 0 0 0-.78-2.354z" /></Symbol>;
};

export function Transfers(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M29.01 38v-4.982a1.027 1.027 0 0 0-1.622-.82l-9.974 6.99a1 1 0 0 0 0 1.637l9.974 6.99a1.027 1.027 0 0 0 1.623-.82V42h23.58a.754.754 0 0 1 .66 1.112A23.986 23.986 0 0 1 8 32a23.562 23.562 0 0 1 .557-5.036A1.244 1.244 0 0 1 9.778 26H34.99v4.982a1.027 1.027 0 0 0 1.623.82l9.974-6.99a1 1 0 0 0 0-1.637l-9.974-6.99a1.027 1.027 0 0 0-1.623.82V22H11.41a.754.754 0 0 1-.66-1.112A23.986 23.986 0 0 1 56 32a23.565 23.565 0 0 1-.557 5.036 1.244 1.244 0 0 1-1.22.964z" /></Symbol>;
};

export function Undo(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M34 52H18a2 2 0 1 1 0-4h16a14 14 0 1 0 0-28h-9.99v4.995a1.024 1.024 0 0 1-1.617.82l-9.949-6.99a.997.997 0 0 1 0-1.638l9.949-6.989a1.024 1.024 0 0 1 1.618.82V16H34a18 18 0 0 1 0 36z" /></Symbol>;
};

export function Unfollow(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M55.446 27.382L50.83 32l4.586 4.586a2 2 0 0 1-2.829 2.829L48 34.828l-4.586 4.587a2 2 0 0 1-2.829-2.829L45.172 32l-4.618-4.618a2 2 0 0 1 2.828-2.828L48 29.172l4.618-4.618a2 2 0 0 1 2.828 2.828zM39.995 52H7.005a1.01 1.01 0 0 1-1.003-1.07c.263-4.53 2.323-5.688 6.852-7.369 4.35-1.614 5.839-3.21 6.094-7.163a9.488 9.488 0 0 1-1.994-1.716 9.552 9.552 0 0 1-1.905-4.713 1.178 1.178 0 0 1-.144.019c-1.084 0-1.905-2.282-1.905-3.886s.557-2.093 1.089-2.093c.113 0 .208.014.31.022A13.565 13.565 0 0 1 14 20.938C14 15.247 16.303 12 23.5 12c3.428 0 3.66.848 4.319 2.031a2.063 2.063 0 0 1 1.295-.406c1.83 0 3.886 2.21 3.886 7.313a13.571 13.571 0 0 1-.399 3.094c.101-.01.196-.023.31-.023.532 0 1.09.488 1.09 2.093s-.822 3.886-1.906 3.886a1.192 1.192 0 0 1-.144-.02 9.55 9.55 0 0 1-1.905 4.714 9.491 9.491 0 0 1-1.994 1.716c.255 3.953 1.745 5.55 6.094 7.163 4.529 1.681 6.588 2.84 6.852 7.37A1.011 1.011 0 0 1 39.995 52z" /></Symbol>;
};

export function Upward(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M25.022 52V32.011h-8.004a2.003 2.003 0 0 1-1.536-3.285L30.48 10.699a1.998 1.998 0 0 1 3.073 0L48.55 28.726a2.003 2.003 0 0 1-1.536 3.285h-8.006V52a1.999 1.999 0 0 1-1.998 2H27.02a1.999 1.999 0 0 1-1.998-2z" /></Symbol>;
};

export function User(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M18.587 46.568c5.909-1.945 7.33-3.856 7.41-9.084a9.764 9.764 0 0 1-1.576-1.567 12.164 12.164 0 0 1-2.237-5.967 1.124 1.124 0 0 1-.279.036c-1.084 0-1.905-2.662-1.905-4.534s.557-2.441 1.089-2.441a3.025 3.025 0 0 1 .42.03A18.55 18.55 0 0 1 21 19c0-7.002 2.667-11 11-11 3.97 0 4.237 1.043 5 2.5a2.306 2.306 0 0 1 1.5-.5c2.12 0 4.5 2.72 4.5 9a18.548 18.548 0 0 1-.51 4.04 3.033 3.033 0 0 1 .42-.03c.532 0 1.09.57 1.09 2.442s-.821 4.534-1.906 4.534a1.123 1.123 0 0 1-.278-.036 12.164 12.164 0 0 1-2.237 5.967 9.768 9.768 0 0 1-1.576 1.567c.079 5.228 1.5 7.14 7.409 9.084 5.528 1.82 8.135 3.091 8.595 7.771a1.512 1.512 0 0 1-1.5 1.661H11.493a1.512 1.512 0 0 1-1.5-1.66c.46-4.68 3.067-5.952 8.594-7.772z" /></Symbol>;
};

export function Upload(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M46 44h-6a2 2 0 0 1 0-4h5.906A8.058 8.058 0 0 0 54 32.073c.038-5.378-4.133-7.227-6.59-8.064a1.992 1.992 0 0 1-1.34-1.704l-.103-1.081A10.106 10.106 0 0 0 36.137 12c-4.668-.054-7.12 2.677-8.707 4.683a1.993 1.993 0 0 1-1.993.696 9.816 9.816 0 0 0-2.634-.377 6.005 6.005 0 0 0-5.792 5.871l-.031 1.304a1.987 1.987 0 0 1-1.255 1.807c-2.593 1.016-5.886 2.66-5.718 7.357A6.97 6.97 0 0 0 17 40h7a2 2 0 0 1 0 4h-7a10.995 10.995 0 0 1-3.988-21.243A9.983 9.983 0 0 1 25.1 13.225a13.991 13.991 0 0 1 24.833 7.448A11.994 11.994 0 0 1 46 44zm-21.814-9.612l6.989-9.974a.996.996 0 0 1 1.638 0l6.988 9.974a1.027 1.027 0 0 1-.819 1.623H34V56a2 2 0 1 1-4 0V36.01h-4.995a1.027 1.027 0 0 1-.82-1.622z" /></Symbol>;
};

export function Unlock(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M15 49.667V32.333A4.177 4.177 0 0 1 19 28h1V16.995c0-4.03.846-6.379 2.729-8.262s4.23-2.73 8.258-2.73h2.026c4.029 0 6.376.847 8.258 2.73S44 12.965 44 16.995V20a2 2 0 0 1-4 0v-3.003c0-2.565-.538-4.06-1.737-5.258s-2.691-1.737-5.255-1.737h-2.016c-2.564 0-4.058.54-5.255 1.737S24 14.432 24 16.997V28h21a4.177 4.177 0 0 1 4 4.333v17.334A4.177 4.177 0 0 1 45 54H19a4.177 4.177 0 0 1-4-4.333z" /></Symbol>;
};

export function UserMan(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M18.587 46.568c5.909-1.945 7.33-3.856 7.41-9.084a9.768 9.768 0 0 1-1.576-1.567 12.163 12.163 0 0 1-2.237-5.967 1.124 1.124 0 0 1-.278.036c-1.085 0-1.906-2.662-1.906-4.534s.557-2.441 1.089-2.441a3.033 3.033 0 0 1 .42.03A18.548 18.548 0 0 1 21 19c0-4.804 1.268-8.13 5.14-8.722C28.283 8.698 31.524 8 35.662 8c2.648 0 .442 1.925 2.93 2.487C40.132 10.835 43 11.996 43 19a18.567 18.567 0 0 1-.509 4.04 3.025 3.025 0 0 1 .42-.03c.532 0 1.089.57 1.089 2.442s-.821 4.534-1.906 4.534a1.125 1.125 0 0 1-.278-.036 12.164 12.164 0 0 1-2.237 5.967 9.76 9.76 0 0 1-1.575 1.567c.078 5.228 1.5 7.14 7.408 9.084 5.528 1.82 8.135 3.091 8.596 7.771a1.512 1.512 0 0 1-1.5 1.661H11.492a1.511 1.511 0 0 1-1.5-1.66c.46-4.68 3.067-5.952 8.594-7.772z" /></Symbol>;
};

export function UserWomanAlternate(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M18.588 46.568c5.871-1.933 7.312-3.833 7.406-8.988A10.523 10.523 0 0 1 23 38c-2.843 0-6-3.552-6-12 0-11.118 3.752-15 8-15a8.88 8.88 0 0 1 2.115.298C28.285 9.354 31.613 8 35 8c7.079 0 12 5.954 12 16 0 4.191-1 6.977-1 9a2.268 2.268 0 0 0 2 2.5c1.62 0 2-1.488 2 .5 0 2.22-2.574 3-6 3a15.748 15.748 0 0 1-5.995-1.437c.092 5.168 1.529 7.07 7.407 9.005 5.528 1.82 8.136 3.091 8.596 7.771A1.512 1.512 0 0 1 52.507 56H11.493a1.512 1.512 0 0 1-1.5-1.66c.46-4.68 3.067-5.952 8.595-7.772z" /></Symbol>;
};

export function Vegetables(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M58 17.007a3.003 3.003 0 0 1-3 3.002c-6.019 0-8.339.131-10.728 2.518-.774.773-1.573.526-2.46-.36s-1.133-1.682-.358-2.455C43.843 17.325 44 15.014 44 9.002a2.994 2.994 0 0 1 5.957-.427 3.974 3.974 0 0 1 5.465 5.458A3.01 3.01 0 0 1 58 17.007zM41.44 36.669c-10.503 10.51-30.977 21.23-33.15 19.056s8.54-22.662 19.044-33.172c2.836-2.837 7.191-4.399 12.84 1.253s4.102 10.025 1.267 12.863z" /></Symbol>;
};

export function Verified(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54.975 34.391l-3.999 3.999a3.411 3.411 0 0 0-.999 2.412v5.756a3.411 3.411 0 0 1-3.41 3.41H40.81a3.412 3.412 0 0 0-2.412 1L34.4 54.966a3.411 3.411 0 0 1-4.824 0l-3.999-3.998a3.412 3.412 0 0 0-2.412-1H17.41a3.411 3.411 0 0 1-3.411-3.41v-5.756a3.411 3.411 0 0 0-1-2.412L9.002 34.39a3.411 3.411 0 0 1 0-4.824l3.998-3.998a3.412 3.412 0 0 0 1-2.413v-5.755a3.411 3.411 0 0 1 3.411-3.411h5.755a3.411 3.411 0 0 0 2.413-1l3.998-3.998a3.411 3.411 0 0 1 4.824 0l3.998 3.998a3.411 3.411 0 0 0 2.412 1h5.756a3.411 3.411 0 0 1 3.411 3.41v5.756a3.412 3.412 0 0 0 1 2.413l3.998 3.998a3.411 3.411 0 0 1 0 4.824zM45.4 20.581a2.001 2.001 0 0 0-3.049.278L29.507 39.25l-5.874-8.411a2.001 2.001 0 0 0-3.05-.278 2.085 2.085 0 0 0-.17 2.67l6.624 9.485a3.014 3.014 0 0 0 4.94 0l13.594-19.464a2.085 2.085 0 0 0-.171-2.67z" /></Symbol>;
};

export function VideoCamera(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56.766 45.848a2 2 0 0 1-2.18-.434l-10-10A2 2 0 0 1 44 34v-4a2 2 0 0 1 .586-1.414l10-10A2 2 0 0 1 58 20v24a2 2 0 0 1-1.234 1.848zM12 16h24a6 6 0 0 1 6 6v20a6 6 0 0 1-6 6H12a6 6 0 0 1-6-6V22a6 6 0 0 1 6-6z" /></Symbol>;
};

export function VideoCameraOff(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M5.997 21.977a.966.966 0 0 1 1.649-.681l25.059 25.051a.956.956 0 0 1-.676 1.632h-20.03a6.001 6.001 0 0 1-6.002-6zm50.78-3.848a2 2 0 0 1 1.234 1.848v24.002a2 2 0 0 1-3.415 1.414l-10.002-10a2 2 0 0 1-.586-1.415v-4a2 2 0 0 1 .586-1.414l10.002-10.002a2 2 0 0 1 2.18-.433zm-20.771-2.153a6.001 6.001 0 0 1 6.001 6v20.05a.967.967 0 0 1-1.652.656L15.274 17.608a.956.956 0 0 1 .676-1.632zM5.404 10.562l40.018 40.013a2 2 0 1 1-2.83 2.829L2.575 13.392a2 2 0 0 1 2.83-2.83z" /></Symbol>;
};

export function Username(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M31.998 50.01a17.894 17.894 0 0 0 8.423-2.094 2 2 0 0 1 1.948 3.495A22.01 22.01 0 1 1 53.99 32.003c0 6.077-3.58 11.004-7.997 11.004a7.99 7.99 0 0 1-6.96-4.07 10.13 10.13 0 0 1-8.535 5.07c-5.797 0-10.497-5.375-10.497-12.004s4.7-12.005 10.497-12.005a9.868 9.868 0 0 1 7.497 3.612v-.503a2.075 2.075 0 0 1 1.664-2.08A2.001 2.001 0 0 1 41.995 23v12.004a4 4 0 0 0 3.999 4.001c2.208 0 3.999-3.135 3.999-7.002A17.997 17.997 0 1 0 31.997 50.01zm-1-26.01c-3.864 0-6.997 3.583-6.997 8.003s3.133 8.003 6.997 8.003 6.998-3.583 6.998-8.003S34.863 24 30.998 24z" /></Symbol>;
};

export function View(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M31.985 48c-14.676 0-21.966-11.672-23.78-15.078a1.957 1.957 0 0 1 0-1.844C10.019 27.672 17.309 16 31.985 16s21.965 11.672 23.78 15.078a1.956 1.956 0 0 1 0 1.844C53.95 36.328 46.66 48 31.985 48zM32 21a11 11 0 1 0 11 11 11 11 0 0 0-11-11zm0 18a7 7 0 1 1 7-7 7 7 0 0 1-7 7z" /></Symbol>;
};

export function ViewOff(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48.67 41.727a1.003 1.003 0 0 1-1.362-.05l-4.782-4.783a.998.998 0 0 1-.227-1.057 10.988 10.988 0 0 0-14.135-14.141.998.998 0 0 1-1.058-.227l-3.194-3.196a.625.625 0 0 1 .244-1.036 25.03 25.03 0 0 1 7.828-1.233c14.67 0 21.955 11.669 23.77 15.073a1.956 1.956 0 0 1 0 1.844 31.778 31.778 0 0 1-7.083 8.806zM13.413 10.595l39.993 40.012a2 2 0 0 1-2.828 2.83L10.585 13.423a2 2 0 0 1 2.827-2.829zm3.244 11.728l4.804 4.806a.998.998 0 0 1 .228 1.055A10.99 10.99 0 0 0 35.81 42.312a.998.998 0 0 1 1.055.228l3.185 3.187a.625.625 0 0 1-.244 1.036 25.032 25.032 0 0 1-7.823 1.23c-14.67 0-21.955-11.668-23.77-15.072a1.955 1.955 0 0 1 0-1.844 31.767 31.767 0 0 1 7.08-8.803 1.004 1.004 0 0 1 1.364.05z" /></Symbol>;
};

export function VoteDown(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M52.593 31.257L33.583 55.24a2.002 2.002 0 0 1-3.137 0l-19.01-23.984a2 2 0 0 1 1.569-3.241H22V14a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v14.016h9.025a2 2 0 0 1 1.568 3.24z" /></Symbol>;
};

export function VoteUp(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M50.995 35.984H42V50a2 2 0 0 1-2 2H24a2 2 0 0 1-2-2V35.984h-9.025a2 2 0 0 1-1.568-3.24l19.01-23.985a2.002 2.002 0 0 1 3.137 0l19.01 23.984a2 2 0 0 1-1.57 3.241z" /></Symbol>;
};

export function VoteUpEmpty(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M10.643 32.126L29.65 8.15a3 3 0 0 1 4.703 0l19.005 23.976a2.999 2.999 0 0 1-2.351 4.86h-7.993v13.011a3.002 3.002 0 0 1-3 3H24.019a3.002 3.002 0 0 1-3-3v-13.01h-8.022a3 3 0 0 1-2.353-4.861zm5.454.862h8.42a.5.5 0 0 1 .5.5v15.01a.5.5 0 0 0 .5.5h12.997a.5.5 0 0 0 .5-.5v-15.01a.5.5 0 0 1 .5-.5h8.39a.5.5 0 0 0 .392-.81L32.393 12.115a.5.5 0 0 0-.784 0L15.705 32.178a.5.5 0 0 0 .392.81z" /></Symbol>;
};

export function VoteDownEmpty(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M53.357 31.874L34.35 55.85a3 3 0 0 1-4.703 0L10.643 31.874a2.999 2.999 0 0 1 2.351-4.86h7.993V14.003a3.002 3.002 0 0 1 3-3h15.995a3.002 3.002 0 0 1 3 3v13.01h8.022a3 3 0 0 1 2.353 4.861zm-5.454-.862h-8.42a.5.5 0 0 1-.5-.5v-15.01a.5.5 0 0 0-.5-.5H25.485a.5.5 0 0 0-.5.5v15.01a.5.5 0 0 1-.5.5h-8.39a.5.5 0 0 0-.391.81l15.902 20.063a.5.5 0 0 0 .784 0l15.904-20.062a.5.5 0 0 0-.392-.81z" /></Symbol>;
};

export function Walk(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48.483 31.964h-5.881a4.83 4.83 0 0 1-3.803-1.87 14.36 14.36 0 0 1-1.142-1.616.172.172 0 0 0-.295.04l-2.617 6.021a.181.181 0 0 0 .04.201l5.961 5.821a5.75 5.75 0 0 1 1.715 3.826l.597 9.33a2.237 2.237 0 0 1-4.447.398l-1.095-8.5a.715.715 0 0 0-.262-.467l-5.865-4.666a.172.172 0 0 0-.264.065l-2.272 4.905a11.403 11.403 0 0 1-2 2.978c-.816.878-4.712 5.436-5.106 5.86a2.213 2.213 0 0 1-1.697.662A2.074 2.074 0 0 1 18 52.855a2.164 2.164 0 0 1 .457-1.334c.297-.378 4.817-6.533 4.932-6.68a1.082 1.082 0 0 0 .218-.51l.852-5.885a23.154 23.154 0 0 1 1.973-6.563l3.258-6.903a.177.177 0 0 0-.175-.254c-.135.014-6.067.36-6.303.385a.52.52 0 0 0-.456.452l-.814 6.103a1.484 1.484 0 0 1-2.957-.205v-6.507a4.25 4.25 0 0 1 3.19-4.149c1.872-.464 7.181-1.43 7.785-1.58a7.665 7.665 0 0 1 5.344.62l2.387 1.221a5.62 5.62 0 0 1 2.362 2.3c.545.991 2.032 3.852 2.134 4.037a.698.698 0 0 0 .482.351l6.085 1.132a1.555 1.555 0 0 1-.271 3.078zM37.998 18a5.006 5.006 0 1 1 5.012-5.006 5.009 5.009 0 0 1-5.012 5.006z" /></Symbol>;
};

export function Warning(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M6.437 47.432L29.43 9.422a2.998 2.998 0 0 1 5.13 0l22.992 38.01a3 3 0 0 1-2.565 4.555H9.003a3 3 0 0 1-2.566-4.555zM31.994 48a3 3 0 1 0-3-3 3 3 0 0 0 3 3zm-2.446-11.248a2.454 2.454 0 0 0 4.891 0l.555-13.755a3 3 0 0 0-6 0z" /></Symbol>;
};

export function Window(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51 54H13a5 5 0 0 1-5-5V15a5 5 0 0 1 5-5h38a5 5 0 0 1 5 5v34a5 5 0 0 1-5 5zM14 14a2 2 0 1 0 2 2 2 2 0 0 0-2-2zm6 0a2 2 0 1 0 2 2 2 2 0 0 0-2-2zm6 0a2 2 0 1 0 2 2 2 2 0 0 0-2-2zm26 9a1 1 0 0 0-1-1H13a1 1 0 0 0-1 1v26a1 1 0 0 0 1 1h38a1 1 0 0 0 1-1z" /></Symbol>;
};

export function Woman(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M42.985 36h-.918a1.277 1.277 0 0 1-1.19-.797l-3.536-8.896-1.313 3.214a.981.981 0 0 0 .122.874 23.163 23.163 0 0 1 3.847 12.555A1.014 1.014 0 0 1 38.993 44H35.06l-.954 13.071a1.013 1.013 0 0 1-1.02.929h-2.175a1.013 1.013 0 0 1-1.019-.929L28.94 44h-3.931a1.014 1.014 0 0 1-1.005-1.05 23.164 23.164 0 0 1 3.847-12.555.981.981 0 0 0 .122-.874l-1.313-3.214-3.535 8.896a1.277 1.277 0 0 1-1.191.797h-.918a1.006 1.006 0 0 1-.998-1.217l2.745-12.085A6.104 6.104 0 0 1 28.748 18h6.504a6.104 6.104 0 0 1 5.986 4.698l2.745 12.085A1.006 1.006 0 0 1 42.985 36zM32 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function WomanAndMan(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M54.989 38h-1.095a1 1 0 0 1-.995-.9l-1.087-10.865-.618 4.333a15.982 15.982 0 0 0-.15 2.877l.905 23.517A1 1 0 0 1 50.95 58h-2.056a1 1 0 0 1-.995-.9L46 38l-1.899 19.1a1 1 0 0 1-.995.9H41.05a1 1 0 0 1-1-1.038l.905-23.517a16 16 0 0 0-.149-2.877l-.619-4.333L39.101 37.1a1 1 0 0 1-.995.9h-1.095a1 1 0 0 1-1-1V23a5 5 0 0 1 5-5h9.978a5 5 0 0 1 5 5v14a1 1 0 0 1-1 1zM46 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5zM28.985 36h-.918a1.277 1.277 0 0 1-1.19-.797l-3.536-8.896-1.313 3.214a.981.981 0 0 0 .122.874 23.163 23.163 0 0 1 3.847 12.555A1.014 1.014 0 0 1 24.993 44H21.06l-.954 13.071a1.013 1.013 0 0 1-1.02.929h-2.175a1.013 1.013 0 0 1-1.019-.929L14.939 44h-3.932a1.014 1.014 0 0 1-1.004-1.05 23.164 23.164 0 0 1 3.847-12.555.98.98 0 0 0 .122-.874l-1.313-3.214-3.536 8.896a1.277 1.277 0 0 1-1.19.797h-.918a1.006 1.006 0 0 1-.998-1.217l2.745-12.085A6.104 6.104 0 0 1 14.748 18h6.504a6.104 6.104 0 0 1 5.986 4.698l2.745 12.085A1.006 1.006 0 0 1 28.985 36zM18 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function Women(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M56.985 36h-.918a1.277 1.277 0 0 1-1.19-.797l-3.536-8.896-1.313 3.214a.982.982 0 0 0 .122.874 23.164 23.164 0 0 1 3.847 12.555A1.014 1.014 0 0 1 52.993 44H49.06l-.954 13.071a1.013 1.013 0 0 1-1.02.929h-2.175a1.013 1.013 0 0 1-1.019-.929L42.939 44h-3.932a1.014 1.014 0 0 1-1.004-1.05 23.163 23.163 0 0 1 3.847-12.555.98.98 0 0 0 .122-.874l-1.314-3.214-3.534 8.896a1.277 1.277 0 0 1-1.191.797h-.918a1.006 1.006 0 0 1-.998-1.217l2.745-12.085A6.104 6.104 0 0 1 42.748 18h6.504a6.104 6.104 0 0 1 5.986 4.698l2.745 12.085A1.006 1.006 0 0 1 56.985 36zM46 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5zM28.985 36h-.918a1.277 1.277 0 0 1-1.19-.797l-3.536-8.896-1.313 3.214a.981.981 0 0 0 .122.874 23.163 23.163 0 0 1 3.847 12.555A1.014 1.014 0 0 1 24.993 44H21.06l-.954 13.071a1.013 1.013 0 0 1-1.02.929h-2.175a1.013 1.013 0 0 1-1.019-.929L14.94 44h-3.931a1.014 1.014 0 0 1-1.005-1.05 23.164 23.164 0 0 1 3.847-12.555.981.981 0 0 0 .122-.874l-1.313-3.214-3.535 8.896A1.277 1.277 0 0 1 7.933 36h-.918a1.006 1.006 0 0 1-.998-1.217l2.745-12.085A6.104 6.104 0 0 1 14.748 18h6.505a6.104 6.104 0 0 1 5.985 4.698l2.745 12.085A1.006 1.006 0 0 1 28.985 36zM18 16a5 5 0 1 1 5-5 5 5 0 0 1-5 5z" /></Symbol>;
};

export function Wrench(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M51.866 43.162a6.172 6.172 0 1 1-8.71 8.701l-14.25-16.5a1.01 1.01 0 0 0-1.144-.265 12.908 12.908 0 0 1-7.004.714 13.073 13.073 0 0 1-10.423-9.957 12.79 12.79 0 0 1-.197-4.606 1.004 1.004 0 0 1 1.71-.56l6.758 6.906a1 1 0 0 0 .969.268l5.996-1.578a1 1 0 0 0 .712-.713l1.578-5.996a1 1 0 0 0-.268-.969l-6.907-6.756a1.004 1.004 0 0 1 .56-1.711 12.79 12.79 0 0 1 4.614.198 13.073 13.073 0 0 1 9.952 10.427 12.909 12.909 0 0 1-.715 6.999 1.004 1.004 0 0 0 .262 1.141l16.507 14.257zm-3.864 7.84a2.999 2.999 0 1 0-2.999-3 2.999 2.999 0 0 0 2.999 3z" /></Symbol>;
};

export function ZoomIn(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48.977 54.505L40.3 44.249a19.052 19.052 0 1 1 3.948-3.947l10.257 8.675a3.92 3.92 0 1 1-5.527 5.528zM29 14a15 15 0 1 0 15 15 15 15 0 0 0-15-15zm2 17v6a2 2 0 1 1-4 0v-6h-6a2 2 0 0 1 0-4h6v-6a2 2 0 1 1 4 0v6h6a2 2 0 0 1 0 4z" /></Symbol>;
};

export function ZoomOut(props: Props): JSX.Element {
	return <Symbol {...props} viewBox="0 0 64 64"><path fillRule="evenodd" d="M48.977 54.505L40.3 44.249a19.052 19.052 0 1 1 3.948-3.947l10.257 8.675a3.92 3.92 0 1 1-5.527 5.528zM29 14a15 15 0 1 0 15 15 15 15 0 0 0-15-15zm-8 17a2 2 0 0 1 0-4h16a2 2 0 0 1 0 4z" /></Symbol>;
};
