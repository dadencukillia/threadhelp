<script>
	import Quill from 'quill';

	import { getLangString } from '../langs';
	import '../assets/quill.snow.css';
    import {onMount} from 'svelte';
	
	let editorContainer;
	let quill;

	let options = {
		modules: {
			toolbar: [
				[{ header: [1, 2, false] }],
				['bold', 'italic', 'underline'],
				['image', 'code-block', 'link'],
			],
		},
		placeholder: getLangString("postPlaceholder"),
		theme: 'snow',
	};

	let data = '';
	let html = '';
	let text = '';

	const onPostSend = () => {
		if (!html.includes("<img") && text.replaceAll(" ", "").replaceAll("\n", "").replaceAll("\t", "").length < 5) {
			alert(getLangString("emptyPostError"));
			return;
		}
		if (html.length > 1024*1024*1024*20) { // 20 Mb
			alert(getLangString("tooLargePostError"));
			return;
		}

		onSend(html);

		if (quill) {
			quill.setText("");
		}
		data = "";
		html = "";
		text = "";
	}

	export let onSend = (html) => {};

	onMount(async () => {
		quill = new Quill(editorContainer, options);
		quill.setText(data);
		quill.on("text-change", (delta, oldDelta, source) => {
			data = quill.getText();
			text = data;
			html = quill.getSemanticHTML();
		});
	});
</script>
<form on:submit|preventDefault={onPostSend}>
	<div bind:this={editorContainer} />
	<button type="submit" class="post">{ getLangString("post") }</button>
</form>

<style>
	form {
		width: 100%;
	}

	button.post {
		color: white;
		background-color: #3498db;
		border: 1px solid #ccc;
		border-top: none;
		border-radius: 0 0 10px 10px;
		filter: none;
		width: 100%;
	}

	button.post:hover {
		color: #efefef;
		background-color: #3083bb;
	}

	:global(div.ql-toolbar) {
		border-radius: 10px 10px 0 0;
	}
</style>
