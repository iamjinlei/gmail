# Gmail Client

Gmail oauth client with a listing API. More APIs can be easily extended.

# Usage

Before to use the client, some configuration is needed (refer to https://stackoverflow.com/questions/37534548/how-to-access-a-gmail-account-i-own-using-gmail-api).
We need to create OAuth client in google api console (https://console.developers.google.com/apis).
When creating OAuth client, set type as web application, and use "https://developers.google.com/oauthplayground" as redirect URI.
Then use the client id and secret to generate *refresh token* in auth playground (https://developers.google.com/oauthplayground).
Download the corresonding *credentials json file*.
Use the refresh token and the path to credentials file as parameters to create client.
User id is the gmail address.
