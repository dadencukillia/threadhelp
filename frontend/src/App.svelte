<script>
    import { onAuthStateChanged } from 'firebase/auth';
	import auth, { makeSignout } from './firebase.js';
	import { setGlobalVar } from './config.js';

	import Main from './routes/Main.svelte';
	import Auth from './routes/Auth.svelte';
	import NotFound from './routes/NotFound.svelte';
	import Loading from './routes/Loading.svelte';

	import { Router, Route } from "svelte-routing";
    import { APIGetRequest } from './api.js';

	let authUser = null;
	let isLoading = true;
	let curAuth = 0;

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
			} catch {
				setTimeout(() => {
					location.reload();
				}, 10000);
				return;
			}

		}

		if (authThread === curAuth) {
			isLoading = false;
			authUser = user;
		}
	});

	export let url = "";
</script>

{#if isLoading}
	<Loading />
{:else}
	<Router {url}>
		{#if authUser === null}
			<Route path="/" component={Auth} />
		{:else}
			<Route path="/"><Main user={authUser} /></Route>
		{/if}
		<Route component={NotFound} />
	</Router>
{/if}

<style>
</style>
