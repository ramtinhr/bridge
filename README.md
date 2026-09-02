# HTTP Bridge (single POST endpoint)

One route: `POST /request`. Send it the method, url, headers, and body you
want fired off, and the HTTP response you get back from this call IS the
upstream response — same status code, same headers, same body. Nothing
wrapped in extra JSON.

## Request format

```
POST /request
X-Api-Key: <your API_KEY>
Content-Type: application/json

{
  "method": "POST",
  "url": "https://api.example.com/v1/things",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer sometoken"
  },
  "body": "{\"name\":\"test\"}"
}
```

## Example

```bash
curl -X POST http://YOUR_SERVER:8080/request \
  -H "X-Api-Key: <your API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "url": "https://api.ipify.org?format=json",
    "headers": {}
  }'
```

The response you see is exactly what `https://api.ipify.org` returned —
status code, headers, and body untouched.

Health check (no auth): `GET /healthz` → `200 ok`.

## Deploying on a server

1. Install Docker:
   ```bash
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER   # log out/in after this
   ```

2. Copy this folder to the server:
   ```bash
   scp -r ./bridge2 youruser@YOUR_SERVER:/opt/bridge
   ```

3. Set the API key:
   ```bash
   cd /opt/bridge
   cp .env.example .env
   nano .env   # set API_KEY
   ```

4. Build and start:
   ```bash
   docker compose up -d --build
   docker compose logs -f
   ```

5. Open the port in your firewall if needed:
   ```bash
   sudo ufw allow 8080/tcp
   ```

## Calling it from your other project

Just make a normal POST request with a JSON body — same as the curl example
above, in whatever language/HTTP client you're already using. No proxy
settings, no special client config — it's a regular API call.
# bridge
