import { registerLanguagePack } from '../langs';

export default () => {
	registerLanguagePack("uk", {
		"authToContinue": "Авторизуйтеся, щоб продовжити",
		"signout": "Вийти",
		"signinGoogle": "Увійти через Google",
		"notFoundMessage": "Схоже, ви заблукали, повертайтеся додому",
		"notFoundButton": "Ходімо",
		"post": "Відправити",
		"postPlaceholder": "Введіть текст посту",
		"emptyPostError": "Текст посту порожній",
		"tooLargePostError": "Пост занадто великий (> 20 Мб)",
		"postIsSendingError": "Зачекайте поки минулий пост відправиться",
		"postRequestError": "Пост не було відправлено через невідому помилку",
		"postIsDeletingError": "Зачекайте поки минулий пост видалиться",
		"postDeleteError": "Пост не було видалено через невідому помилку",
		"buttonUnfold": "Показати повністю...",
		"tryAgain": "Спробувати знову",
	});
}
