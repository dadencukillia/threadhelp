import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'


const envKeys = [
	"FIREBASE_API_KEY",
	"FIREBASE_AUTH_DOMAIN",
	"FIREBASE_PROJECT_ID",
	"FIREBASE_STORAGE_BUCKET",
	"FIREBASE_MESSAGING_SENDER_ID",
	"FIREBASE_APP_ID",
	"OAUTH_ALLOWED_EMAIL_DOMAIN",
];
const envDefine = {};
for (const envKey of envKeys) {
	envDefine["import.meta.env." + envKey] = JSON.stringify(process.env[envKey]);
}

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [svelte()],
	define: {
		...envDefine
	},
});
