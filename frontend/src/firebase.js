import { initializeApp } from "firebase/app";
import { getAuth, signInWithPopup } from "firebase/auth";
import {firebaseConfig, provider} from "./config";

const app = initializeApp(firebaseConfig);
const auth = getAuth(app);

export default auth;

export async function makeAuth() {
	return signInWithPopup(auth, provider);
}

export async function makeSignout() {
	return auth.signOut();
}
