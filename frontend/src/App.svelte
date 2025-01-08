<script>
    import { onAuthStateChanged } from 'firebase/auth';
	import auth, { initAuth, makeSignout } from './firebase.js';

	import Main from './routes/Main.svelte';
	import OAuth from './routes/OAuth.svelte';
	import Passcode from './routes/Passcode.svelte';
	import NotFound from './routes/NotFound.svelte';
	import Loading from './routes/Loading.svelte';

	import { Router, Route } from "svelte-routing";
    import { APIPostRequest, APIGetRequest, connectPath } from './api.js';
    import {onMount} from 'svelte';

	let authUser = null;
	let isLoading = true;
	let curAuth = 0;
	let provider = "";

	onMount(async () => {
		try {
			provider = await fetch(connectPath("provider")).then(e => e.text());
		} catch {
			setTimeout(() => location.reload(), 10000);
			return;
		}

		if (provider === "oauth") {
			initAuth();
			onAuthStateChanged(auth, async (user) => {
				const authThread = ++curAuth;
				isLoading = true;

				if (user !== null && authUser !== user) {
					try {
						const res = await APIGetRequest("check");

						if (res.status !== 200) {
							makeSignout();
							return;
						}

						const json = await res.json();
						user["isAdmin"] = json["admin"];
						user["provider"] = "oauth";
					} catch {
						setTimeout(() => location.reload(), 10000);
						return;
					}
				}

				if (authThread === curAuth) {
					isLoading = false;
					authUser = user;
				}
			});
		} else if (provider === "passcode") {
			try {
				const res = await APIPostRequest("check");
				if (res.status === 200) {
					authUser = {
						"provider": "passcode",
						"logout": () => {authUser = null},
						"info": await res.json()
					};
				} else {
					authUser = null;
				}

				isLoading = false;
			} catch {
				setTimeout(() => location.reload(), 10000);
			}
		} else {
			setTimeout(() => location.reload(), 10000);
		}

	});
	
	export let url = "";
</script>

{#if isLoading}
	<Loading />
{:else}
	<Router {url}>
		{#if authUser === null}
			{#if provider == "oauth"}
				<Route path="/" component={OAuth} />
			{:else if provider == "passcode"}
				<Route path="/">
					<Passcode setUserInfo={(newInfo) => {
						authUser = {
							"provider": "passcode",
							"logout": () => {authUser = null},
							"info": newInfo
						};
					}} />
				</Route>
			{:else}
				<Loading />
			{/if}
		{:else}
			<Route path="/"><Main user={authUser} /></Route>
		{/if}
		<Route component={NotFound} />
	</Router>
{/if}

<style>
</style>
