<script>
    import { onMount, onDestroy } from 'svelte';
	import auth from '../firebase';
    import { getLangString } from '../langs';
    import { APIGetRequest } from '../api';

	export let user;
	export let postId = "";
	export let publishTime = "";
	export let userId = "";
	export let userDisplayName = "Unknown";
	export let onDelete = () => {};

	let post = "<h1>Nothing here</h1>";
	let loading = true;
	let loadError = false;

	const isAdmin = user["isAdmin"];

	let isLarge = false;
	let isUnfolded = false;

	let postBodyElement;
	let observer;

	async function loadContent() {
		loading = true;
		loadError = false;
		
		try {
			const r = await APIGetRequest("getPostContent/" + postId);
			post = await r.text();
			loadError = false;
			loading = false;	
		} catch {
			loading = false;
			loadError = true;
		}
	}

	$: if (postBodyElement) {
		try {
			observer.disconnect();
		} catch {}

		observer = new ResizeObserver(() => {
			const height = postBodyElement.offsetHeight;
			if (height > 400) {
				isLarge = true;
			} else {
				isLarge = false;
			}
		});
		observer.observe(postBodyElement);
	}

	onMount(async () => {
		await loadContent();
	});

	onDestroy(() => {
		observer.disconnect();
	});
</script>

{#if loadError}
<article class="post error">
	<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" class="bi bi-exclamation-circle" viewBox="0 0 16 16">
		<path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/>
		<path d="M7.002 11a1 1 0 1 1 2 0 1 1 0 0 1-2 0M7.1 4.995a.905.905 0 1 1 1.8 0l-.35 3.507a.552.552 0 0 1-1.1 0z"/>
	</svg>
	<button on:click={ loadContent }>{ getLangString("tryAgain") }</button>
</article>
{:else if loading}
<article class="post loading"/>
{:else}
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
	</div>
	{#if isLarge && !isUnfolded }
		<button on:click={() => { isUnfolded = true }}>{ getLangString("buttonUnfold") }</button>
	{/if}
</article>
{/if}

<style>
	@keyframes fadein {
		from {
			scale: 0;
			filter: opacity(0.0);
		}
		
		to {
			scale: 1;
			filter: opacity(1.0);
		}
	}

	@keyframes shine {
		0% {
			background-position: 300%;
		}
		100% {
			background-position: 0%;
		}
	}

	.post.error {
		width: 100%;
		border-radius: 10px;
		color: white;
		background-color: red;
		text-align: center;

		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 10px;
		padding: 20px;

		& svg {
			width: clamp(32px, 25%, 128px);
		}

		& button {
			filter: none;
		}
	}

	.post.loading {
		animation: fadein 0.2s ease-out;
		width: 100%;
		height: 400px;
		border-radius: 10px;	

		background: linear-gradient(125deg,#0000 33%, rgba(255,255,255,0.3) 50%,#0000 66%) #3498db;
		background-size: 300% 100%;
		animation: shine 2s infinite linear;
	}

	.post:not(.loading):not(.error) {
		position: relative;
		animation: fadein 0.2s ease-out;
		background-color: white;
		width: 100%;
		border: 1px solid #eee;
		padding: 10px;
		border-radius: 10px;	

		&>.topbar {
			display: flex;
			flex-direction: row;
			align-items: center;
			border-bottom: 1px solid #eee;
			padding-bottom: 5px;
			gap: 5px;

			& .user {
				font-weight: bold;
				text-wrap: nowrap;
				overflow-x: hidden;
				text-overflow: ellipsis;
			}

			& .date {
				font-size: 12px;
				color: #333;
				margin-left: auto;
				text-wrap: nowrap;
			}

			& button {
				color: white;
				background-color: red;

				&:hover {
					color: #efefef;
					background-color: #e11;
				}

				&.adminblue {
					background-color: #3498db;
				}

				&.adminbutton:hover {
					background-color: #3083bb;
				}
			}
		}

		& .body {
			padding: 10px 0;
			overflow: auto;
			width: 100%;

			&.cut-content {
				max-height: 450px;
				overflow-y: hidden;
			}
		}

		&>button {
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
