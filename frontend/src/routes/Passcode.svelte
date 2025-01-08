<script>
    import { getLangString } from '../langs';
	import { APIPostRequest } from '../api';

	let passwordField = "";
	let loggingIn = false;

	export let setUserInfo = (newInfo) => {};

	async function makeLogin() {
		if (loggingIn) {
			return;
		}
		loggingIn = true;

		const resp = await APIPostRequest("check", {
			"password": passwordField
		});

		if (resp.status === 200) {
			setUserInfo(await resp.json());
		} else {
			alert(getLangString("incorrectPassword"));
		}

		loggingIn = false;
	}
</script>

<main>
	<h1>{getLangString("enterPasswordToContinue")}</h1>
	<form on:submit|preventDefault={makeLogin}>
		<input bind:value={passwordField} type="password">
		<button>{getLangString("continue")}</button>
	</form>	
</main>

<style>
	main {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		text-align: center;
		gap: 30px;
		padding: clamp(5px, 3%, 20px);
	}

	form {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	input {
		height: 30px;
		border-radius: 10px;
		border: 1px solid black;
		padding: 10px;
	}

	button {
		font-size: 16px;
		font-weight: 600;
		width: 100%;
		overflow: hidden;
	}
</style>
