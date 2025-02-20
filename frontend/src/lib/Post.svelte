<script>
    import { onMount, onDestroy } from 'svelte';
    import { getLangString } from '../langs';
    import { APIGetRequest, APIPostRequest } from '../api';

	export let user;
	export let postId = "";
	export let publishTime = "";
	export let userId = "";
	export let userDisplayName = "Unknown";
	export let onDelete = () => {};
	export let outCommunication = {};

	outCommunication["updateLikes"] = async () => {
		try {
			const r = await APIGetRequest("getPostLikes/" + postId);
			const json = await r.json();
			likes = json["likes"];
			liked = json["liked"];
		} catch {}
	};

	let post = "<h1>Nothing here</h1>";
	let loading = true;
	let loadError = false;

	let liked = true;
	let likesLoading = true;
	let likes = 0;

	const isAdmin = user["isAdmin"] ?? false;
	const selfUid = user["provider"] === "oauth" ? user.uid : user.info.id;

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
			return true;
		} catch {
			loading = false;
			loadError = true;
			return false;
		}
	}

	async function loadLikes() {
		likesLoading = true;
		loadError = false;
		
		try {
			const r = await APIGetRequest("getPostLikes/" + postId);
			const json = await r.json();
			likes = json["likes"];
			liked = json["liked"];
			loadError = false;
			likesLoading = false;
			return true;
		} catch {
			loading = false;
			loadError = true;
			return false;
		}
	}

	async function setLiked(newValue) {
		likesLoading = true;

		if (newValue) {
			try {
				const r = await APIPostRequest("likePost", {"id": postId});
				if (r.status === 200) {
					liked = newValue;
				}
			} catch {}
			likesLoading = false;
		} else {
			try {
				const r = await APIPostRequest("unlikePost", {"id": postId});
				if (r.status === 200) {
					liked = newValue;
				}
			} catch {}
			likesLoading = false;
		}
	}

	async function loadEverything() {
		if (!await loadContent()) {
			return;
		}
		await loadLikes();
	}

	$: if (postBodyElement && !loading && !loadError) {
		try {
			observer.disconnect();
		} catch {}

		observer = new ResizeObserver(() => {
			if (postBodyElement) {
				const height = postBodyElement.offsetHeight;
				if (height > 400) {
					isLarge = true;
				} else {
					isLarge = false;
				}
			}
		});
		observer.observe(postBodyElement);
	}

	onMount(() => {
		loadEverything();
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
	<button on:click={ loadEverything }>{ getLangString("tryAgain") }</button>
</article>
{:else if loading}
<article class="post loading"/>
{:else}
<article class="post">
	<div class="topbar">
		{#if selfUid === userId || isAdmin }
			<button on:click={onDelete} class={isAdmin && selfUid !== userId ? "adminblue" : ""}>
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
	<div class="footer">
		{#if likesLoading}
			<div class="like-loader" />
		{:else}
			<button class={"like"+(liked?" liked":"")} on:click={() => setLiked(!liked)}>
				{#if liked}
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-hand-thumbs-up-fill" viewBox="0 0 16 16">
					  <path d="M6.956 1.745C7.021.81 7.908.087 8.864.325l.261.066c.463.116.874.456 1.012.965.22.816.533 2.511.062 4.51a10 10 0 0 1 .443-.051c.713-.065 1.669-.072 2.516.21.518.173.994.681 1.2 1.273.184.532.16 1.162-.234 1.733q.086.18.138.363c.077.27.113.567.113.856s-.036.586-.113.856c-.039.135-.09.273-.16.404.169.387.107.819-.003 1.148a3.2 3.2 0 0 1-.488.901c.054.152.076.312.076.465 0 .305-.089.625-.253.912C13.1 15.522 12.437 16 11.5 16H8c-.605 0-1.07-.081-1.466-.218a4.8 4.8 0 0 1-.97-.484l-.048-.03c-.504-.307-.999-.609-2.068-.722C2.682 14.464 2 13.846 2 13V9c0-.85.685-1.432 1.357-1.615.849-.232 1.574-.787 2.132-1.41.56-.627.914-1.28 1.039-1.639.199-.575.356-1.539.428-2.59z"/>
					</svg>
				{:else}
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-hand-thumbs-up" viewBox="0 0 16 16">
					  <path d="M8.864.046C7.908-.193 7.02.53 6.956 1.466c-.072 1.051-.23 2.016-.428 2.59-.125.36-.479 1.013-1.04 1.639-.557.623-1.282 1.178-2.131 1.41C2.685 7.288 2 7.87 2 8.72v4.001c0 .845.682 1.464 1.448 1.545 1.07.114 1.564.415 2.068.723l.048.03c.272.165.578.348.97.484.397.136.861.217 1.466.217h3.5c.937 0 1.599-.477 1.934-1.064a1.86 1.86 0 0 0 .254-.912c0-.152-.023-.312-.077-.464.201-.263.38-.578.488-.901.11-.33.172-.762.004-1.149.069-.13.12-.269.159-.403.077-.27.113-.568.113-.857 0-.288-.036-.585-.113-.856a2 2 0 0 0-.138-.362 1.9 1.9 0 0 0 .234-1.734c-.206-.592-.682-1.1-1.2-1.272-.847-.282-1.803-.276-2.516-.211a10 10 0 0 0-.443.05 9.4 9.4 0 0 0-.062-4.509A1.38 1.38 0 0 0 9.125.111zM11.5 14.721H8c-.51 0-.863-.069-1.14-.164-.281-.097-.506-.228-.776-.393l-.04-.024c-.555-.339-1.198-.731-2.49-.868-.333-.036-.554-.29-.554-.55V8.72c0-.254.226-.543.62-.65 1.095-.3 1.977-.996 2.614-1.708.635-.71 1.064-1.475 1.238-1.978.243-.7.407-1.768.482-2.85.025-.362.36-.594.667-.518l.262.066c.16.04.258.143.288.255a8.34 8.34 0 0 1-.145 4.725.5.5 0 0 0 .595.644l.003-.001.014-.003.058-.014a9 9 0 0 1 1.036-.157c.663-.06 1.457-.054 2.11.164.175.058.45.3.57.65.107.308.087.67-.266 1.022l-.353.353.353.354c.043.043.105.141.154.315.048.167.075.37.075.581 0 .212-.027.414-.075.582-.05.174-.111.272-.154.315l-.353.353.353.354c.047.047.109.177.005.488a2.2 2.2 0 0 1-.505.805l-.353.353.353.354c.006.005.041.05.041.17a.9.9 0 0 1-.121.416c-.165.288-.503.56-1.066.56z"/>
					</svg>
				{/if}
				{likes}
			</button>
		{/if}
		{#if isLarge && !isUnfolded }
			<button on:click={() => { isUnfolded = true }} class="unfold">{ getLangString("buttonUnfold") }</button>
		{/if}
	</div>
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

			& pre {
				min-width: 100%;
				width: fit-content;
				display: block;
				border-radius: 3px;
				padding: 5px 10px;
				background-color: #23241f;
				color: #f8f8f2;
				margin: 5px 0;
			}

			&.cut-content {
				max-height: 450px;
				overflow-y: hidden;
			}
		}

		& .footer {
			display: flex;
			flex-direction: row;
			align-items: center;
			border-top: 1px solid #eee;
			padding-top: 5px;
			gap: 5px;

			& .like {
				background-color: transparent;
				padding: 0;
				width: fit-content;
				height: 16px;
				
				&.liked {
					color: #3498db;
				}
			}

			& .like-loader {
				display: block;
				width: 16px;
				height: 16px;
				border: 5px solid #f3f3f3;
				border-top: 5px solid #3498db;
				border-radius: 50%;
				animation: spin 1s cubic-bezier(0, 0, 0, 1.06) infinite;
			}

			& .unfold {
				margin-left: auto;
			}
		}
	}

	:global(img) {
		max-width: 100%;
	}
</style>
