# ThreadHelp
The customizable self-hosted website for publishing thematic posts for a group of users.

# Downloading
1. To get started, you need to install **git**, **docker**, **docker compose** and **make**. Command for Ubuntu: 
```
sudo apt install git docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```
2. Clone the repository: 
```
git clone github.com/dadencukillia/threadhelp
```

# Setting up
1. Rename the `.env_sample` file to `.env`.
2. Create a project in [Firebase](https://console.firebase.google.com).
3. On the main project page, click on the tag icon (**</>**) and give the project a name and click on the “Register app” button.
4. In the **.env** file, change the values of the variables starting with `FIREBASE_` to the corresponding values in the code of the second paragraph.

**Code example**
```JavaScript
...
const firebaseConfig = {
  apiKey: "AIzaSyCTLAro5hXRm7roTH7E9knkan1-Gs0K6MM",
  authDomain: "test-f7ad5.firebaseapp.com",
  projectId: "test-f7ad5",
  storageBucket: "test-f7ad5.appspot.com",
  messagingSenderId: "323094117308",
  appId: "1:323094117308:web:8f5971c583eca4ec7d89b9",
  measurementId: "NOT_NEEDED"
};
...
```

**Relevant contents of the .env file**
```dotenv
...
FIREBASE_API_KEY=AIzaSyCTLAro5hXRm7roTH7E9knkan1
FIREBASE_AUTH_DOMAIN=test-f7ad5.firebaseapp.com
FIREBASE_PROJECT_ID=test-f7ad5
FIREBASE_STORAGE_BUCKET=test-f7ad5.appspot.com
FIREBASE_MESSAGING_SENDER_ID=323094117308
FIREBASE_APP_ID=1:323094117308:web:8f5971c583eca4ec7d89b9
...
```
5. In the “**Authentication**” tab, enable the “**Google**” provider.
6. In the settings, go to the “Service accounts” tab and click on the “Generate new private key” button. Rename the downloaded file to `firebaseSecretKey.json` and place it in the repository folder (next to the “.env” and “docker-compose.yml” files).

# Important
The site does not have built-in support for the HTTPS protocol, as required by Google OAuth2. Use ngrok and certbot to add support for the HTTPS protocol.

# Launching a website
To start the site, just run the command `sudo make`, and to stop it - `sudo make stop`. The site will be available on port `80` and at `127.0.0.1` (if running locally).