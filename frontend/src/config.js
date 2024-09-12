import {GoogleAuthProvider} from "firebase/auth";

export const firebaseConfig = {
	apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
	authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
	projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
	storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
	messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID,
	appId: import.meta.env.VITE_FIREBASE_APP_ID,
};

export const provider = new GoogleAuthProvider();
provider.setCustomParameters({
	'hd': import.meta.env.VITE_GOOGLEOAUTH_MAIL_DOMAIN,
});

export const apiPath = import.meta.env.VITE_API_PATH;
