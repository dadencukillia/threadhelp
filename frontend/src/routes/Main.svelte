<script>
	import { onMount, onDestroy } from 'svelte';

	import TextEditor from '../lib/TextEditor.svelte';
	import {APIGetRequest, APIPostRequest, connectPath} from '../api';
	import Header from '../lib/Header.svelte';
	import Post from '../lib/Post.svelte';
    import { getLangString } from '../langs';
    import {EventSourcePolyfill} from 'event-source-polyfill';

	let posts = [];
	let sending = false;
	let sendBuffer = [];
	let deleting = false;
    
	let sse;

	export let user;

	async function loadNewPosts() {
		return APIGetRequest("getNextTenPosts/" + posts.at(posts.length - 1).postId).then(r => r.json()).then(r => {
			posts = posts.concat(r);
			return r.length;
		}).catch(e => console.log(e));
	}

	onDestroy(() => {
		sse.close();
	});

	onMount(() => {
		APIGetRequest("tenNewestPosts").then(r => r.json()).then(async r => {
			posts = r;
			
			let postsToRemove = [];
			let postsToAdd = [];
			let lastPostLoaded = false;

			sse = new EventSourcePolyfill(
				connectPath("events"),
				{
					"headers": {
						"Auth-Token": await user.getIdToken()
					}
				}
			);

			sse.onmessage = ev => {
				const data = ev.data;
				console.log(data);
				if (data.startsWith("delPost;")) { 
					const postId = data.split(";")[1];
					console.log("Delete post: ", postId);

					if (lastPostLoaded) {
						posts = posts.filter(e => e.postId !== postId);
					} else {
						postsToRemove.push(postId);
					}
				} else if (data.startsWith("newPost;")) {
					const json = JSON.parse(data.split(";").slice(1).join(";"));
					console.log("New post: ", json);

					if (sending) {
						if (sendBuffer.includes(json.postId)) {
							sending = false;
							sendBuffer.length = 0;
						} else {
							sendBuffer.push(json.postId);
						}
					}

					if (lastPostLoaded) {
						posts = [json].concat(posts);
					} else {
						postsToAdd.push(json);
					}
				}
			}

			sse.onerror = () => {
				location.reload();
			}

			(async () => {
				while (!lastPostLoaded) {
					await (new Promise((resolve) => setTimeout(resolve, 10)));

					const postsEl = document.querySelectorAll("article.post");
					if (postsEl.length > 0) {
						const lastPost = postsEl.item(postsEl.length - 1);

						if (lastPost !== null && lastPost.getBoundingClientRect().top < (window.innerHeight || document.documentElement.clientHeight)) {
							const loaded = await loadNewPosts();
							console.log(`Loaded ${loaded} posts.`)
							if (loaded === 0) {
								lastPostLoaded = true;
							}
						}
					}

					if (postsToAdd.length !== 0) {
						posts = postsToAdd.concat(posts);
						postsToAdd = [];
					}
					if (postsToRemove.length !== 0) {
						for (let postId of postsToRemove) {
							posts = posts.filter(e => e.postId !== postId);
						}

						postsToRemove = [];
					}
				}
			})();
		}).catch(e => {
			console.log(e);
			location.reload();
		});
	});
</script>

<main>
	<Header user={user} />
	<div class="wrapper">
		<container>
			<TextEditor onSend={html => {
				if (sending) {
					alert(getLangString("postIsSendingError"));
					return;
				}

				sending = true;
				APIPostRequest("sendPost", {
					"content": html,
				}).then(r => {
					if (r.status !== 200) {
						return Promise.reject(new Error("invalid status code"));
					}
					return r.text();
				}).then(r => {
					if (sendBuffer.includes(r)) {
						sending = false;
						sendBuffer.length = 0;
					} else {
						sendBuffer.push(r);
					}
				}).catch(() => {
					sendBuffer.length = 0;
					sending = false;
					alert(getLangString("postRequestError"));
				});
			}} />
			{#each posts as {userId, userDisplayName, post, postId, pubTime} (postId)}
				<Post onDelete={() => {
					if (deleting) {
						alert(getLangString("postIsDeletingError"));
						return;
					}
					deleting = true;

					APIPostRequest("deletePost", {
						"id": postId,
					}).then(r => {
						deleting = false;
						if (r.status === 200) {
							posts = posts.filter(e => e.postId !== postId);
						} else {
							alert(getLangString("postDeleteError"));
							deleting = false;
						}
					}).catch(() => {
						alert(getLangString("postDeleteError"));
						deleting = false;
					});
				}} user={user} postId={postId} userId={userId} userDisplayName={userDisplayName} post={post} publishTime={pubTime} />
			{/each}
		</container>
	</div>
</main>

<style>
	main {
		display: flex;
		flex-direction: column;
		align-items: stretch;
		height: 100vh;
		max-height: 100vh;
	}

	.wrapper {
		width: 100%;
		height: 100%;
		flex-grow: 1;
		overflow: scroll;
	}

	.wrapper::-webkit-scrollbar {
		display: none;
	}

	container {
		display: flex;
		flex-direction: column;
		gap: 20px;
		width: 100%;
		max-width: 1200px;
		min-height: 100%;
		padding: clamp(5px, 3vw, 20px);
		padding-bottom: 200px;
		margin: 0 auto;
	}
</style>
