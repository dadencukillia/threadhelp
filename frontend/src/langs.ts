import loadEn from './langs/en';
import loadUk from './langs/uk';

const defaultLang: string = "en";
const userLanguages: string[] = navigator.languages.map(e => e.split("-")[0].toLowerCase());

const languagePacks: Map<string, Dictionary> = new Map<string, Dictionary>();

interface Dictionary {
	authToContinue: string,
	signout: string,
	signinGoogle: string,
	notFoundMessage: string,
	notFoundButton: string,
	post: string,
	postPlaceholder: string,
	emptyPostError: string,
	tooLargePostError: string,
	postIsSendingError: string,
	postRequestError: string,
	postIsDeletingError: string,
	postDeleteError: string,
	buttonUnfold: string,
	tryAgain: string,
	enterPasswordToContinue: string,
	continue: string,
	incorrectPassword: string,
}

let activeDictionary: Dictionary | null = null;

export function registerLanguagePack(langSlug: string, dictionary: Dictionary) {
	languagePacks.set(langSlug, dictionary);
}

export function updateLanguagePacks() {
	for (const langShortName of userLanguages.concat([defaultLang])) {
		if (languagePacks.has(langShortName)) {
			activeDictionary = languagePacks.get(langShortName) ?? null;
			break;
		}
	}
}

export function getLangString(key: string): string {
	return activeDictionary![key];
}

loadEn();
loadUk();

updateLanguagePacks();
