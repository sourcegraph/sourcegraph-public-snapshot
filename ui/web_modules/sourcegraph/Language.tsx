const names: { [key: string]: string } = {
	golang: "Go",
	java: "Java",
	python: "Python",
	javascript: "JavaScript",
	other: "Other",
};
export const langName = (lang: string) => names[lang];
