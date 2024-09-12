<script>
    import {navigate} from "svelte-routing";
import getLangString from "../lang";

</script>

<main>
	<h1>{ getLangString("notFoundMessage") }</h1>
	<button on:click={() => {
		navigate("/");
	}}>
		<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-house-door-fill" viewBox="0 0 16 16"><path d="M6.5 14.5v-3.505c0-.245.25-.495.5-.495h2c.25 0 .5.25.5.5v3.5a.5.5 0 0 0 .5.5h4a.5.5 0 0 0 .5-.5v-7a.5.5 0 0 0-.146-.354L13 5.793V2.5a.5.5 0 0 0-.5-.5h-1a.5.5 0 0 0-.5.5v1.293L8.354 1.146a.5.5 0 0 0-.708 0l-6 6A.5.5 0 0 0 1.5 7.5v7a.5.5 0 0 0 .5.5h4a.5.5 0 0 0 .5-.5"/></svg>
		<span>{ getLangString("notFoundButton") }</span>
	</button>
</main>

<style>
	main {
		position: fixed;
		width: 100%;
		height: 100%;

		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: center;
		text-align: center;
		justify-content: center;
	}

	button {
		display: flex;
		flex-direction: row;
		gap: 5px;
		align-items: center;
		justify-content: center;
		padding: 10px 20px;
	}
</style>
