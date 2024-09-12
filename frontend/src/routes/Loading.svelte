<main>
	<div class="loader"></div>
</main>

<style>
	main {
		display: flex;
		align-items: center;
		justify-content: center;
		position: fixed;
		width: 100%;
		height: 100%;
		top: 0;
		left: 0;
	}

	.loader {
		border: 5px solid #f3f3f3;
		border-top: 5px solid #3498db;
		border-radius: 50%;
		width: clamp(20px, 50vw, 300px);
		height: clamp(20px, 50vw, 300px);
		animation: spin 1s cubic-bezier(0, 0, 0, 1.06) infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
</style>
