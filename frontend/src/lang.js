const defaultLang = "en";
const userLanguages = navigator.languages.map(e => e.split("-").at(0).toLowerCase());

const translateDict = new Map();

const ukDict = new Map();
ukDict.set("authToContinue", "Авторизуйтеся, щоб продовжити");
ukDict.set("signout", "Вийти");
ukDict.set("signinGoogle", "Увійти через Google");
ukDict.set("notFoundMessage", "Схоже, ви заблукали, повертайтеся додому");
ukDict.set("notFoundButton", "Ходімо");
ukDict.set("post", "Відправити");
ukDict.set("postPlaceholder", "Введіть текст посту");
ukDict.set("emptyPostError", "Текст посту порожній");
ukDict.set("tooLargePostError", "Пост занадто великий (> 20 Мб)");
ukDict.set("postIsSendingError", "Зачекайте поки минулий пост відправиться");
ukDict.set("postRequestError", "Пост не було відправлено через невідому помилку");
ukDict.set("postIsDeletingError", "Зачекайте поки минулий пост видалиться");
ukDict.set("postDeleteError", "Пост не було видалено через невідому помилку");
ukDict.set("buttonUnfold", "Показати повністю...");
translateDict.set("uk", ukDict);

const enDict = new Map();
enDict.set("authToContinue", "Sign in to continue");
enDict.set("signout", "Sign out");
enDict.set("signinGoogle", "Continue with Google");
enDict.set("notFoundMessage", "You seem to be lost, go back home");
enDict.set("notFoundButton", "Let's go");
enDict.set("post", "Send post");
enDict.set("postPlaceholder", "Enter the text of the post");
enDict.set("emptyPostError", "The text of the post is empty");
enDict.set("tooLargePostError", "The post is too large (> 20 Mb)");
enDict.set("postIsSendingError", "Wait until the previous post is sent");
enDict.set("postRequestError", "The post was not sent due to an unknown error");
enDict.set("postIsDeletingError", "Wait until the previous post is deleted");
enDict.set("postDeleteError", "The post was not deleted due to an unknown error");
enDict.set("buttonUnfold", "Show more...");
translateDict.set("en", enDict);

let applyLang = defaultLang;
for (const lang of userLanguages) {	
	if (translateDict.has(lang)) {
		applyLang = lang;
		break;
	}
}

export default function getLangString(key) {
	return translateDict.get(applyLang).get(key);
}
