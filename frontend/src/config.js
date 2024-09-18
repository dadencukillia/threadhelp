import { GoogleAuthProvider } from "firebase/auth";

export const firebaseConfig = {
	apiKey: import.meta.env.FIREBASE_API_KEY,
	authDomain: import.meta.env.FIREBASE_AUTH_DOMAIN,
	projectId: import.meta.env.FIREBASE_PROJECT_ID,
	storageBucket: import.meta.env.FIREBASE_STORAGE_BUCKET,
	messagingSenderId: import.meta.env.FIREBASE_MESSAGING_SENDER_ID,
	appId: import.meta.env.FIREBASE_APP_ID,
};

export const provider = new GoogleAuthProvider();
provider.setCustomParameters({
	'hd': import.meta.env.OAUTH_ALLOWED_EMAIL_DOMAIN,
});

export const apiPath = "/api";
