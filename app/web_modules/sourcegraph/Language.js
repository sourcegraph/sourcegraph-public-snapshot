// @flow

export type LanguageID = "golang" | "java" | "python" | "javascript";
export const allLangs: Array<LanguageID> = ["golang", "java", "python", "javascript"];

export const supportedLangs: Array<LanguageID> = ["golang", "java"];
export const langIsSupported = (lang: LanguageID) => supportedLangs.includes(lang);

const names: {[key: LanguageID]: string} = {
	golang: "Go",
	java: "Java",
	python: "Python",
	javascript: "JavaScript",
};
export const langName = (lang: LanguageID) => names[lang];
