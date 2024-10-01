<script>
    import { onMount } from 'svelte';
	import auth from '../firebase.js';
    import getLangString from '../lang.js';

	export let user;
	export const postId = "";
	export let publishTime = "";
	export let userId = "";
	export let userDisplayName = "Unknown";
	export let post = "<h1>Nothing here</h1>";
	export let onDelete = () => {};

	const isAdmin = user["isAdmin"];

	let isLarge = false;
	let isUnfolded = false;

	let postBodyElement;
	onMount(() => {
		const observer = new ResizeObserver(() => {
			const height = postBodyElement.offsetHeight;
			if (height > 400) {
				isLarge = true;
			} else {
				isLarge = false;
			}
		});

		observer.observe(postBodyElement);
	});
</script>

<article class="post">
	<div class="topbar">
		{#if auth.currentUser.uid === userId || isAdmin }
			<button on:click={onDelete} class={isAdmin && auth.currentUser.uid !== userId ? "adminblue" : ""}>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-trash-fill" viewBox="0 0 16 16">
					<path d="M2.5 1a1 1 0 0 0-1 1v1a1 1 0 0 0 1 1H3v9a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V4h.5a1 1 0 0 0 1-1V2a1 1 0 0 0-1-1H10a1 1 0 0 0-1-1H7a1 1 0 0 0-1 1zm3 4a.5.5 0 0 1 .5.5v7a.5.5 0 0 1-1 0v-7a.5.5 0 0 1 .5-.5M8 5a.5.5 0 0 1 .5.5v7a.5.5 0 0 1-1 0v-7A.5.5 0 0 1 8 5m3 .5v7a.5.5 0 0 1-1 0v-7a.5.5 0 0 1 1 0"/>
				</svg>
			</button>
		{/if}
		<span class="user">{ userDisplayName }</span>
		<span class="date">{ new Date(publishTime).toLocaleString(navigator.language) }</span>
	</div>
	<div class={"body" + (isLarge && !isUnfolded ? " cut-content" : "")} bind:this={postBodyElement}>
		<p>{ @html post }</p>
		{#if isLarge && !isUnfolded }
		<button on:click={() => { isUnfolded = true }}>{ getLangString("buttonUnfold") }</button>
		{/if}
	</div>
</article>

<style>
	.post {
		animation: slidein 0.2s ease-out;
		background-color: white;
		width: 100%;
		border: 1px solid #eee;
		padding: 10px;
		border-radius: 10px;

	}

	@keyframes slidein {
		from {
			scale: 0;
			filter: opacity(0.0);
		}
		
		to {
			scale: 1;
			filter: opacity(1.0);
		}
	}

	.topbar {
		display: flex;
		flex-direction: row;
		align-items: center;
		border-bottom: 1px solid #eee;
		padding-bottom: 5px;
		gap: 5px;
	}

	.topbar .user {
		font-weight: bold;
		text-wrap: nowrap;
		overflow-x: hidden;
		text-overflow: ellipsis;
	}

	.topbar .date {
		font-size: 12px;
		color: #333;
		margin-left: auto;
		text-wrap: nowrap;
	}

	.topbar button {
		color: white;
		background-color: red;
	}

	.topbar button:hover {
		color: #efefef;
		background-color: #e11;
	}

	.topbar button.adminblue {
		background-color: #3498db;
	}

	.topbar button.adminblue:hover {
		background-color: #3083bb;
	}

	.body {
		padding: 10px 0;
		overflow: auto;
		width: 100%;
	}

	.body.cut-content {
		position: relative;
		max-height: 450px;
		overflow-y: hidden;

		& button {
			position: absolute;
			bottom: 5px;
			left: 50%;
			transform: translateX(-50%);
		}
	}
	
	:global(img) {
		max-width: 100%;
	}
</style>
