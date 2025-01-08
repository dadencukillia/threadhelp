import { initializeApp } from "firebase/app";
import { getAuth, signInWithPopup } from "firebase/auth";
import { firebaseConfig, provider } from "./config";

let app = null;
let auth = null;

export function initAuth() {
	app = initializeApp(firebaseConfig);
	auth = getAuth(app);
}

export default auth;

export async function makeAuth() {
	if (auth === null) return;

	return signInWithPopup(auth, provider);
}

export async function makeSignout() {
	if (auth === null) return;

	return auth.signOut();
}
