import { registerLanguagePack } from '../langs';

export default () => {
	registerLanguagePack("en", {
		"authToContinue": "Sign in to continue",
		"signout": "Sign out",
		"signinGoogle": "Continue with Google",
		"notFoundMessage": "You seem to be lost, go back home",
		"notFoundButton": "Let's go",
		"post": "Send post",
		"postPlaceholder": "Enter the text of the post",
		"emptyPostError": "The text of the post is empty",
		"tooLargePostError": "The post is too large (> 20 Mb)",
		"postIsSendingError": "Wait until the previous post is sent",
		"postRequestError": "The post was not sent due to an unknown error",
		"postIsDeletingError": "Wait until the previous post is deleted",
		"postDeleteError": "The post was not deleted due to an unknown error",
		"buttonUnfold": "Show more...",
		"tryAgain": "Try again",
	});
}
