<script>
	import { getLangString } from '../langs';
    import {makeSignout} from "../firebase";
    import {APIGetRequest} from '../api';

	export let user = {};

	async function signOut() {
		const provider = user["provider"];
		if (provider === "oauth") {
			await makeSignout();
		} else if (provider === "passcode") {
			await APIGetRequest("logout");
			document.cookie = "Auth-Token=; expires=Thu, 01 Jan 1970 00:00:00 GMT;";
			user["logout"]();
		}
	}
</script>

<header>
	<button on:click={signOut}>
		<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-door-open-fill" viewBox="0 0 16 16"><path d="M1.5 15a.5.5 0 0 0 0 1h13a.5.5 0 0 0 0-1H13V2.5A1.5 1.5 0 0 0 11.5 1H11V.5a.5.5 0 0 0-.57-.495l-7 1A.5.5 0 0 0 3 1.5V15zM11 2h.5a.5.5 0 0 1 .5.5V15h-1zm-2.5 8c-.276 0-.5-.448-.5-1s.224-1 .5-1 .5.448.5 1-.224 1-.5 1"/></svg>
		<span>{ getLangString("signout") }</span>
	</button>
	<span>{user.provider === "oauth" ? user.displayName : user.info.name}</span>
	{#if user.provider === "oauth"}
		<img class="avatar" src={user.photoURL} alt="avatar">
	{/if}
</header>

<style>
	@media (max-width: 350px) {
		button span {
			display: none;
		}
	}

	header {
		background-color: #f3f3f3;
		display: flex;
		align-items: center;
		flex-direction: row;
		width: 100%;
		padding: 10px;
		gap: 10px;
		border-bottom: 1px solid #ccc;
	}

	button {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: center;
		gap: 5px;
	}

	span {
		margin-left: auto;
		text-wrap: nowrap;
		overflow-x: hidden;
		text-overflow: ellipsis;
	}

	img {
		border-radius: 10px;
		height: 48px;
	}
</style>
