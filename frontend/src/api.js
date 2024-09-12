import { apiPath } from "./config";
import auth, {makeSignout} from "./firebase";

export function connectPath(path) {
	return apiPath + ((!apiPath.endsWith("/") && !path.startsWith("/"))?"/":"") + path;
}

export async function APIGetRequest(endpoint) {
	let authToken = auth.currentUser !== null ? await auth.currentUser.getIdToken() : null;

	let res = await fetch(connectPath(endpoint), {
		method: "GET",
		headers: {
			"Auth-Token": authToken
		}
	})

	if (res.status === 403 || res.status === 401) {
		await makeSignout()
	}

	return res
}

export async function APIPostRequest(endpoint, data) {
	let authToken = auth.currentUser !== null ? await auth.currentUser.getIdToken() : null;

	let res = await fetch(connectPath(endpoint), {
		method: "POST",
		headers: {
			"Auth-Token": authToken
		},
		body: JSON.stringify(data)
	})

	if (res.status === 403 || res.status === 401) {
		await makeSignout()
	}

	return res
}
