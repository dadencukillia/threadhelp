# ThreadHelp
The customizable self-hosted website for publishing thematic posts for a group of users.

# Downloading
1. To get started, you need to install **git**, **docker**, **docker compose** and **make**. Command for Ubuntu: 
```
# From https://docs.docker.com/engine/install/ubuntu/
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt install git docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin docker-compose
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

## HTTPS protocol by Let's Encrypt
Google OAuth2 recommends using the HTTPS protocol. ThreadHelp provides the ability to set the HTTPS protocol through the **.env** file. To do this, you must have a domain name for the site. The **.env** file contains the parameters `USE_HTTPS`, `HTTPS_EMAIL` and `HTTPS_DOMAIN`. To enable the HTTPS protocol, you need to set `USE_HTTPS` to `true`, enter your email in `HTTPS_EMAIL`, and enter the domain name of your site in `HTTPS_DOMAIN`. Example:
```
...
USE_HTTPS=true
HTTPS_EMAIL=you@gmail.com
HTTPS_DOMAIN=example.com
...
```

# Important
The website is intended for a narrow circle of people and may be vulnerable to high traffic. Use it for group communication!

# Launching a website
To start the site, just run the command `sudo make`, and to stop it - `sudo make stop`. The site will be available on port `80` and at `127.0.0.1` (if running locally).
