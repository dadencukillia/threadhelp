<script>
	import {Editor} from '@tadashi/svelte-editor-quill';

	import getLangString from '../lang';
	import '../assets/quill.snow.css';
	
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

	const onTextChange = event => {
		;({text, html} = event?.detail ?? {});
		data = html;
	};

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

		data = "";
	}

	export let onSend = html => {};
</script>
<form on:submit|preventDefault={onPostSend}>
	<Editor
	  {options}
	  {data}
	  on:text-change={onTextChange}
	/>
	<button type="submit" class="post">{ getLangString("post") }</button>
</form>

<style>
	form {
		width: 100%;
	}

	button.post {
		background: rgb(214,170,48);
		background: linear-gradient(to right, #bf953f, #fcf6ba, #b38728, #fbf5b7, #aa771c) fixed;
		border: 1px solid #ccc;
		border-top: none;
		border-radius: 0 0 10px 10px;
		filter: none;
		width: 100%;
	}

	:global(div.ql-toolbar) {
		border-radius: 10px 10px 0 0;
	}
</style>
